package nodepool

import (
	"fmt"
	"path"

	apps "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
	"github.com/jetstack/navigator/pkg/controllers/common"
	"github.com/jetstack/navigator/pkg/util/ptr"
)

const (
	sharedVolumeName      = "shared"
	sharedVolumeMountPath = "/shared"

	cassDataVolumeName      = "cassandra-data"
	cassDataVolumeMountPath = "/var/lib/cassandra"

	cassSnitch = "GossipingPropertyFileSnitch"

	// See https://jolokia.org/reference/html/agents.html#jvm-agent
	jolokiaHost    = "127.0.0.1"
	jolokiaPort    = 8778
	jolokiaContext = "/jolokia"
)

func StatefulSetForCluster(
	cluster *v1alpha1.CassandraCluster,
	np *v1alpha1.CassandraClusterNodePool,
) *apps.StatefulSet {
	statefulSetName := util.NodePoolResourceName(cluster, np)
	nodePoolLabels := util.NodePoolLabels(cluster, np.Name)
	podLabels := util.NodePoolLabels(cluster, np.Name)
	podLabels[v1alpha1.PilotLabel] = ""
	image := cassImageToUse(&cluster.Spec)

	set := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            statefulSetName,
			Namespace:       cluster.Namespace,
			Labels:          nodePoolLabels,
			Annotations:     make(map[string]string),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
		Spec: apps.StatefulSetSpec{
			Replicas: ptr.Int32(0),
			Selector: &metav1.LabelSelector{
				MatchLabels: nodePoolLabels,
			},
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: apps.OrderedReadyPodManagement,
			ServiceName:         statefulSetName,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
					Annotations: map[string]string{
						"prometheus.io/port":   "8080",
						"prometheus.io/path":   "/",
						"prometheus.io/scrape": "true",
					},
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: util.ServiceAccountName(cluster),
					NodeSelector:       np.NodeSelector,
					SchedulerName:      np.SchedulerName,
					Volumes: []apiv1.Volume{
						apiv1.Volume{
							Name: sharedVolumeName,
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
					},
					SecurityContext: &apiv1.PodSecurityContext{
						FSGroup: cluster.Spec.NavigatorClusterConfig.SecurityContext.RunAsUser,
					},
					InitContainers: []apiv1.Container{
						pilotInstallationContainer(&cluster.Spec.NavigatorClusterConfig.PilotImage),
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
								"--leader-election-config-map=$(LEADER_ELECTION_CONFIG_MAP)",
								// Trailing slash is important.
								// Allows url.ResolveReference to link to a
								// descendant rather than a sibling.
								fmt.Sprintf(
									"--jolokia-url=http://%s:%d%s/",
									jolokiaHost,
									jolokiaPort,
									jolokiaContext,
								),
							},
							Image: fmt.Sprintf(
								"%s:%s",
								image.Repository,
								image.Tag,
							),
							ImagePullPolicy: image.PullPolicy,
							ReadinessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									HTTPGet: &apiv1.HTTPGetAction{
										Port: intstr.FromInt(12000),
									},
								},
								// Test logs show that Cassandra begins
								// listening for CQL connections ~45s after startup.
								InitialDelaySeconds: 60,
								TimeoutSeconds:      10,
								PeriodSeconds:       15,
								SuccessThreshold:    3,
								FailureThreshold:    1,
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
									HTTPGet: &apiv1.HTTPGetAction{
										Port: intstr.FromInt(12001),
									},
								},
								// Don't start performing liveness probes until
								// readiness probe has had a chance to succeed
								// at least 3 times.
								InitialDelaySeconds: 90,
								TimeoutSeconds:      10,
								PeriodSeconds:       30,
								SuccessThreshold:    1,
								FailureThreshold:    6,
							},
							Resources: np.Resources,
							SecurityContext: &apiv1.SecurityContext{
								RunAsUser: cluster.Spec.NavigatorClusterConfig.SecurityContext.RunAsUser,
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
									ContainerPort: int32(9042),
								},
								{
									Name:          "prometheus",
									ContainerPort: int32(8080),
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      sharedVolumeName,
									MountPath: sharedVolumeMountPath,
									ReadOnly:  true,
								},
								{
									Name:      cassDataVolumeName,
									MountPath: cassDataVolumeMountPath,
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
								// Deliberately set to a single space to force Cassandra to do a host name lookup.
								// See https://github.com/apache/cassandra/blob/cassandra-3.11.2/conf/cassandra.yaml#L592
								{
									Name:  "CASSANDRA_LISTEN_ADDRESS",
									Value: " ",
								},
								{
									Name:  "CASSANDRA_BROADCAST_ADDRESS",
									Value: " ",
								},
								{
									Name:  "CASSANDRA_RPC_ADDRESS",
									Value: " ",
								},
								// Set a non-existent default seed.
								// The Kubernetes Seed Provider will fall back to a default seed host if it can't look up seeds via the CASSANDRA_SERVICE.
								// And if the CASSANDRA_SEEDS environment variable is not set, it defaults to localhost.
								// Which could cause confusion if a non-seed node is temporarily unable to lookup the seed nodes from the service.
								// We want the list of seeds to be strictly controlled by the service.
								// See:
								// https://github.com/docker-library/cassandra/blame/master/3.11/docker-entrypoint.sh#L31 and
								// https://github.com/apache/cassandra/blob/cassandra-3.11.2/conf/cassandra.yaml#L416 and
								// https://github.com/kubernetes/examples/blob/cabf8b8e4739e576837111e156763d19a64a3591/cassandra/java/src/main/java/io/k8s/cassandra/KubernetesSeedProvider.java#L69 and
								// https://github.com/kubernetes/examples/blob/cabf8b8e4739e576837111e156763d19a64a3591/cassandra/go/main.go#L51
								{
									Name:  "CASSANDRA_SEEDS",
									Value: "black-hole-dns-name",
								},
								{
									Name:  "CASSANDRA_ENDPOINT_SNITCH",
									Value: cassSnitch,
								},
								{
									Name:  "CASSANDRA_SERVICE",
									Value: util.SeedsServiceName(cluster),
								},
								{
									Name:  "CASSANDRA_CLUSTER_NAME",
									Value: cluster.Name,
								},
								{
									Name:  "CASSANDRA_DC",
									Value: ptr.DerefString(np.Datacenter),
								},
								{
									Name:  "CASSANDRA_RACK",
									Value: ptr.DerefString(np.Rack),
								},
								{
									Name: "JVM_OPTS",
									Value: fmt.Sprintf(
										"-javaagent:%s/jolokia.jar=host=%s,port=%d,agentContext=%s "+
											"-javaagent:%s/jmx_prometheus_javaagent.jar=8080:%s/jmx_prometheus_javaagent.yaml "+
											common.JVMCgroupOpts,
										sharedVolumeMountPath,
										jolokiaHost,
										jolokiaPort,
										jolokiaContext,
										sharedVolumeMountPath,
										sharedVolumeMountPath,
									),
								},
								{
									Name: "CLASSPATH",
									Value: path.Join(
										sharedVolumeMountPath,
										"libcassandra-kubernetes-seed-provider.jar",
									),
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
								{
									Name: "LEADER_ELECTION_CONFIG_MAP",
									// TODO: trim the length of this string
									Value: fmt.Sprintf("cassandra-%s-leaderelection", cluster.Name),
								},
							},
						},
					},
				},
			},
		},
	}
	if np.Persistence != nil {
		volumeClaimTemplateAnnotations := map[string]string{}

		if np.Persistence.StorageClass != nil {
			volumeClaimTemplateAnnotations["volume.beta.kubernetes.io/storage-class"] = *np.Persistence.StorageClass
		}

		set.Spec.VolumeClaimTemplates = []apiv1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:        cassDataVolumeName,
					Annotations: volumeClaimTemplateAnnotations,
				},
				Spec: apiv1.PersistentVolumeClaimSpec{
					AccessModes: []apiv1.PersistentVolumeAccessMode{
						apiv1.ReadWriteOnce,
					},
					StorageClassName: np.Persistence.StorageClass,
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceStorage: np.Persistence.Size,
						},
					},
				},
			},
		}
	} else {
		set.Spec.Template.Spec.Volumes = append(
			set.Spec.Template.Spec.Volumes,
			apiv1.Volume{
				Name: cassDataVolumeName,
				VolumeSource: apiv1.VolumeSource{
					EmptyDir: &apiv1.EmptyDirVolumeSource{},
				},
			},
		)
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
			"cp",
			"/pilot",
			"/jolokia.jar",
			"/jmx_prometheus_javaagent.jar",
			"/jmx_prometheus_javaagent.yaml",
			"/libcassandra-kubernetes-seed-provider.jar",
			fmt.Sprintf("%s/", sharedVolumeMountPath),
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
				apiv1.ResourceMemory: resource.MustParse("50Mi"),
			},
			Limits: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("10m"),
				apiv1.ResourceMemory: resource.MustParse("50Mi"),
			},
		},
	}
}
