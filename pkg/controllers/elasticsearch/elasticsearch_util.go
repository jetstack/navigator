package elasticsearch

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	apps "k8s.io/client-go/pkg/apis/apps/v1beta1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/jetstack-experimental/navigator/pkg/apis/marshal"
	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/marshal/v1alpha1"
)

const (
	typeName = "es"
	kindName = "ElasticsearchCluster"

	nodePoolVersionAnnotationKey = "elasticsearch.marshal.io/deployed-version"
)

var (
	trueVar  = true
	falseVar = false
)

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func elasticsearchPodTemplateSpec(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) (*apiv1.PodTemplateSpec, error) {
	initContainers := buildInitContainers(c, np)

	initContainersJSON, err := json.Marshal(initContainers)

	if err != nil {
		return nil, fmt.Errorf("error marshaling init containers: %s", err.Error())
	}

	volumes := []apiv1.Volume{}

	if np.Persistence == nil {
		volumes = append(volumes, apiv1.Volume{
			Name: "elasticsearch-data",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		})
	}

	rolesBytes, err := json.Marshal(np.Roles)

	if err != nil {
		return nil, fmt.Errorf("error marshaling roles: %s", err.Error())
	}

	pluginsBytes, err := json.Marshal(c.Spec.Plugins)

	if err != nil {
		return nil, fmt.Errorf("error marshaling plugins: %s", err.Error())
	}

	return &apiv1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: buildNodePoolLabels(c, np.Name, np.Roles...),
			Annotations: map[string]string{
				"pod.beta.kubernetes.io/init-containers": string(initContainersJSON),
			},
		},
		Spec: apiv1.PodSpec{
			TerminationGracePeriodSeconds: int64Ptr(1800),
			// TODO
			ServiceAccountName: "",
			SecurityContext: &apiv1.PodSecurityContext{
				FSGroup: int64Ptr(c.Spec.Image.FsGroup),
			},
			Volumes: volumes,
			Containers: []apiv1.Container{
				{
					Name:            "elasticsearch",
					Image:           c.Spec.Image.Repository + ":" + c.Spec.Image.Tag,
					ImagePullPolicy: apiv1.PullPolicy(c.Spec.Image.PullPolicy),
					Args: []string{
						"start",
						"--podName=$(POD_NAME)",
						"--clusterURL=$(CLUSTER_URL)",
						"--namespace=$(NAMESPACE)",
						"--plugins=" + string(pluginsBytes),
						"--roles=" + string(rolesBytes),
					},
					Env: []apiv1.EnvVar{
						// TODO: Tidy up generation of discovery & client URLs
						{
							Name:  "DISCOVERY_HOST",
							Value: clusterService(c, "discovery", false, nil, "master").Name,
						},
						{
							Name:  "CLUSTER_URL",
							Value: "http://" + clusterService(c, "clients", true, nil, "client").Name + ":9200",
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
							Name: "NAMESPACE",
							ValueFrom: &apiv1.EnvVarSource{
								FieldRef: &apiv1.ObjectFieldSelector{
									FieldPath: "metadata.namespace",
								},
							},
						},
					},
					SecurityContext: &apiv1.SecurityContext{
						Capabilities: &apiv1.Capabilities{
							Add: []apiv1.Capability{
								apiv1.Capability("IPC_LOCK"),
							},
						},
					},
					ReadinessProbe: &apiv1.Probe{
						Handler: apiv1.Handler{
							HTTPGet: &apiv1.HTTPGetAction{
								Port: intstr.FromInt(12001),
								Path: "/",
							},
						},
						InitialDelaySeconds: int32(60),
						PeriodSeconds:       int32(10),
						TimeoutSeconds:      int32(5),
					},
					LivenessProbe: &apiv1.Probe{
						Handler: apiv1.Handler{
							HTTPGet: &apiv1.HTTPGetAction{
								Port: intstr.FromInt(12000),
								Path: "/",
							},
						},
						InitialDelaySeconds: int32(60),
						PeriodSeconds:       int32(10),
						TimeoutSeconds:      int32(5),
					},
					Resources: apiv1.ResourceRequirements{
						Requests: np.Resources.Requests,
						Limits:   np.Resources.Limits,
					},
					Ports: []apiv1.ContainerPort{
						{
							Name:          "transport",
							ContainerPort: int32(9300),
						},
						{
							Name:          "http",
							ContainerPort: int32(9200),
						},
					},
					VolumeMounts: []apiv1.VolumeMount{
						{
							Name:      "elasticsearch-data",
							MountPath: "/usr/share/elasticsearch/data",
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}, nil
}

func buildInitContainers(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) []apiv1.Container {
	containers := make([]apiv1.Container, len(c.Spec.Sysctl))
	for i, sysctl := range c.Spec.Sysctl {
		containers[i] = apiv1.Container{
			Name:            fmt.Sprintf("tune-sysctl-%d", i),
			Image:           "busybox:latest",
			ImagePullPolicy: apiv1.PullIfNotPresent,
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &trueVar,
			},
			Command: []string{
				"sysctl", "-w", sysctl,
			},
		}
	}
	return containers
}

func buildNodePoolLabels(c v1alpha1.ElasticsearchCluster, poolName string, roles ...string) map[string]string {
	labels := map[string]string{
		"cluster": c.Name,
		"app":     "elasticsearch",
	}
	if poolName != "" {
		labels["pool"] = poolName
	}
	for _, role := range roles {
		labels[role] = "true"
	}
	return labels
}

func clusterService(c v1alpha1.ElasticsearchCluster, name string, http bool, annotations map[string]string, roles ...string) *apiv1.Service {
	svc := apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            c.Name + "-" + name,
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference(c)},
			Labels:          buildNodePoolLabels(c, "", roles...),
			Annotations:     annotations,
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeClusterIP,
			Ports: []apiv1.ServicePort{
				{
					Name:       "transport",
					Port:       int32(9300),
					TargetPort: intstr.FromInt(9300),
				},
			},
			Selector: buildNodePoolLabels(c, "", roles...),
		},
	}

	if http {
		svc.Spec.Ports = append(svc.Spec.Ports, apiv1.ServicePort{
			Name:       "http",
			Port:       int32(9200),
			TargetPort: intstr.FromInt(9200),
		})
	}

	return &svc
}

func nodePoolDeployment(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) (*extensions.Deployment, error) {
	elasticsearchPodTemplate, err := elasticsearchPodTemplateSpec(c, np)

	if err != nil {
		return nil, fmt.Errorf("error building elasticsearch container: %s", err.Error())
	}

	deploymentName := nodePoolResourceName(c, np)
	depl := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            deploymentName,
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference(c)},
			Annotations: map[string]string{
				nodePoolVersionAnnotationKey: c.Spec.Version,
			},
			Labels: buildNodePoolLabels(c, np.Name, np.Roles...),
		},
		Spec: extensions.DeploymentSpec{
			Replicas: int32Ptr(int32(np.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: buildNodePoolLabels(c, np.Name, np.Roles...),
			},
			Template: *elasticsearchPodTemplate,
		},
	}

	// TODO: make this safer?
	depl.Spec.Template.Spec.Containers[0].Args = append(
		depl.Spec.Template.Spec.Containers[0].Args,
		"--controllerKind=Deployment",
		"--controllerName="+deploymentName,
	)
	return depl, nil
}

func nodePoolStatefulSet(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) (*apps.StatefulSet, error) {
	volumeClaimTemplateAnnotations, volumeResourceRequests := map[string]string{}, apiv1.ResourceList{}

	if np.Persistence != nil {
		if np.Persistence.StorageClass != "" {
			volumeClaimTemplateAnnotations["volume.beta.kubernetes.io/storage-class"] = np.Persistence.StorageClass
		}

		if size := np.Persistence.Size; size != "" {
			storageRequests, err := resource.ParseQuantity(size)

			if err != nil {
				return nil, fmt.Errorf("error parsing storage size quantity '%s': %s", size, err.Error())
			}

			volumeResourceRequests[apiv1.ResourceStorage] = storageRequests
		}
	}

	elasticsearchPodTemplate, err := elasticsearchPodTemplateSpec(c, np)

	if err != nil {
		return nil, fmt.Errorf("error building elasticsearch container: %s", err.Error())
	}

	statefulSetName := nodePoolResourceName(c, np)

	ss := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            statefulSetName,
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference(c)},
			Annotations: map[string]string{
				nodePoolVersionAnnotationKey: c.Spec.Version,
			},
			Labels: buildNodePoolLabels(c, np.Name, np.Roles...),
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    int32Ptr(int32(np.Replicas)),
			ServiceName: statefulSetName,
			Selector: &metav1.LabelSelector{
				MatchLabels: buildNodePoolLabels(c, np.Name, np.Roles...),
			},
			Template: *elasticsearchPodTemplate,
			VolumeClaimTemplates: []apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "elasticsearch-data",
						Annotations: volumeClaimTemplateAnnotations,
					},
					Spec: apiv1.PersistentVolumeClaimSpec{
						AccessModes: []apiv1.PersistentVolumeAccessMode{
							apiv1.ReadWriteOnce,
						},
						Resources: apiv1.ResourceRequirements{
							Requests: volumeResourceRequests,
						},
					},
				},
			},
		},
	}

	// TODO: make this safer?
	ss.Spec.Template.Spec.Containers[0].Args = append(
		ss.Spec.Template.Spec.Containers[0].Args,
		"--controllerKind=StatefulSet",
		"--controllerName="+statefulSetName,
	)

	return ss, nil
}

func clusterServiceAccount(c v1alpha1.ElasticsearchCluster) *apiv1.ServiceAccount {
	return &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:            resourceBaseName(c),
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference(c)},
		},
	}
}

func isManagedByCluster(c v1alpha1.ElasticsearchCluster, meta metav1.ObjectMeta) bool {
	clusterOwnerRef := ownerReference(c)
	for _, o := range meta.OwnerReferences {
		if clusterOwnerRef.APIVersion == o.APIVersion &&
			clusterOwnerRef.Kind == o.Kind &&
			clusterOwnerRef.Name == o.Name &&
			clusterOwnerRef.UID == o.UID {
			return true
		}
	}
	return false
}

func managedOwnerRef(meta metav1.ObjectMeta) *metav1.OwnerReference {
	for _, ref := range meta.OwnerReferences {
		if ref.APIVersion == marshal.GroupName+"/v1alpha1" && ref.Kind == kindName {
			return &ref
		}
	}
	return nil
}

func ownerReference(c v1alpha1.ElasticsearchCluster) metav1.OwnerReference {
	// Really, this should be able to use the TypeMeta of the ElasticsearchCluster.
	// There is an issue open on client-go about this here: https://github.com/kubernetes/client-go/issues/60
	return metav1.OwnerReference{
		APIVersion: marshal.GroupName + "/v1alpha1",
		Kind:       kindName,
		Name:       c.Name,
		UID:        c.UID,
	}
}

func resourceBaseName(c v1alpha1.ElasticsearchCluster) string {
	return typeName + "-" + c.Name
}

func nodePoolResourceName(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) string {
	return fmt.Sprintf("%s-%s", resourceBaseName(c), np.Name)
}

func nodePoolVersionAnnotation(m map[string]string) string {
	return m[nodePoolVersionAnnotationKey]
}

func nodePoolIsStateful(np v1alpha1.ElasticsearchClusterNodePool) bool {
	return np.Persistence != nil
}
