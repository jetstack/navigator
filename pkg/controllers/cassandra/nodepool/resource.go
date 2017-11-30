package nodepool

import (
	"fmt"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
	apps "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	sharedVolumeName      = "shared"
	sharedVolumeMountPath = "/shared"
)

func StatefulSetForCluster(
	cluster *v1alpha1.CassandraCluster,
	np *v1alpha1.CassandraClusterNodePool,
) *apps.StatefulSet {

	statefulSetName := util.NodePoolResourceName(cluster, np)
	seedProviderServiceName := util.SeedProviderServiceName(cluster)
	nodePoolLabels := util.NodePoolLabels(cluster, np.Name)
	set := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            statefulSetName,
			Namespace:       cluster.Namespace,
			Labels:          util.ClusterLabels(cluster),
			Annotations:     make(map[string]string),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    util.Int32Ptr(int32(np.Replicas)),
			ServiceName: seedProviderServiceName,
			Selector: &metav1.LabelSelector{
				MatchLabels: nodePoolLabels,
			},
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: apps.ParallelPodManagement,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: nodePoolLabels,
				},
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{
						apiv1.Volume{
							Name: sharedVolumeName,
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
					},
					InitContainers: []apiv1.Container{
						pilotInstallationContainer(&cluster.Spec.PilotImage),
					},
					Containers: []apiv1.Container{
						{
							Name: "cassandra",
							Command: []string{
								fmt.Sprintf("%s/pilot", sharedVolumeMountPath),
							},
							Args: []string{
								"--v=4",
								"--logtostderr",
								"--pilot-name=$(POD_NAME)",
								"--pilot-namespace=$(POD_NAMESPACE)",
							},
							Image: fmt.Sprintf(
								"%s:%s",
								cluster.Spec.Image.Repository,
								cluster.Spec.Image.Tag,
							),
							ImagePullPolicy: apiv1.PullPolicy(
								cluster.Spec.Image.PullPolicy,
							),
							ReadinessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									Exec: &apiv1.ExecAction{
										// XXX The ready-probe.sh script is only
										// available in the
										// gcr.io/google-samples/cassandra image.
										// Replace this when we have a Cassandra pilot.
										// It can perform similar ready probe.
										Command: []string{
											"/usr/bin/timeout",
											"10",
											"/ready-probe.sh",
										},
									},
								},
								// Test logs show that Cassandra begins
								// listening for CQL connections ~45s after startup.
								InitialDelaySeconds: 60,
								// XXX Kubernetes ignores the TimeoutSeconds for Exec probes.
								// See https://github.com/kubernetes/kubernetes/issues/26895
								TimeoutSeconds:   10,
								PeriodSeconds:    15,
								SuccessThreshold: 3,
								FailureThreshold: 1,
							},
							// XXX: You might imagine that LivenessProbes begin
							// only after a successful ReadinessProbe,
							// but in fact they start at the same time.
							// Set a large initial delay to avoid declaring
							// the database dead before it has had a chance to
							// initialise.
							// See: https://github.com/kubernetes/kubernetes/issues/27114
							LivenessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									Exec: &apiv1.ExecAction{
										Command: []string{
											"/usr/bin/timeout",
											"10",
											"/ready-probe.sh",
										},
									},
								},
								// Don't start performing liveness probes until
								// readiness probe has had a chance to succeed
								// at least 3 times.
								InitialDelaySeconds: 90,
								// XXX Kubernetes ignores the TimeoutSeconds for Exec probes.
								// See https://github.com/kubernetes/kubernetes/issues/26895
								TimeoutSeconds:   10,
								PeriodSeconds:    30,
								SuccessThreshold: 1,
								FailureThreshold: 6,
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "intra-node",
									ContainerPort: int32(7000),
								},
								{
									Name:          "intra-node-tls",
									ContainerPort: int32(7001),
								},
								{
									Name:          "jmx",
									ContainerPort: int32(7199),
								},
								{
									Name:          "cql",
									ContainerPort: util.DefaultCqlPort,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      sharedVolumeName,
									MountPath: sharedVolumeMountPath,
									ReadOnly:  false,
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "MAX_HEAP_SIZE",
									Value: "512M",
								},
								{
									Name:  "HEAP_NEWSIZE",
									Value: "100M",
								},
								{
									Name: "CASSANDRA_SEEDS",
									Value: fmt.Sprintf(
										"%s-0.%s.%s.svc.cluster.local",
										statefulSetName,
										seedProviderServiceName,
										cluster.Namespace,
									),
								},
								{
									Name:  "CASSANDRA_CLUSTER_NAME",
									Value: cluster.Name,
								},
								{
									Name:  "CASSANDRA_DC",
									Value: "DC1-K8Demo",
								},
								{
									Name:  "CASSANDRA_RACK",
									Value: "Rack1-K8Demo",
								},
								{
									Name:  "CASSANDRA_AUTO_BOOTSTRAP",
									Value: "false",
								},
								{
									Name: "POD_IP",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
								apiv1.EnvVar{
									Name: "POD_NAME",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								apiv1.EnvVar{
									Name: "POD_NAMESPACE",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return set
}

func pilotInstallationContainer(
	image *v1alpha1.ImageSpec,
) apiv1.Container {
	return apiv1.Container{
		Name: "install-pilot",
		Image: fmt.Sprintf(
			"%s:%s",
			image.Repository, image.Tag),
		ImagePullPolicy: apiv1.PullPolicy(image.PullPolicy),
		Command: []string{
			"cp", "/pilot", fmt.Sprintf("%s/pilot", sharedVolumeMountPath),
		},
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:      sharedVolumeName,
				MountPath: sharedVolumeMountPath,
				ReadOnly:  false,
			},
		},
		Resources: apiv1.ResourceRequirements{
			Requests: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("10m"),
				apiv1.ResourceMemory: resource.MustParse("8Mi"),
			},
		},
	}
}
