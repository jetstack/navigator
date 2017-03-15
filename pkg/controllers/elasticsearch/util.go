package elasticsearch

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"gitlab.jetstack.net/marshal/colonel/pkg/api/v1"
)

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func elasticsearchPodTemplateSpec(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) (*apiv1.PodTemplateSpec, error) {
	initContainers := buildInitContainers(c, np)

	initContainersJSON, err := json.Marshal(initContainers)

	if err != nil {
		return nil, fmt.Errorf("error marshaling init containers: %s", err.Error())
	}

	elasticsearchContainerRequests, elasticsearchContainerLimits :=
		apiv1.ResourceList{},
		apiv1.ResourceList{}

	if np.Resources != nil {
		if req := np.Resources.Requests; req != nil {
			elasticsearchContainerRequests, err = parseResources(req)

			if err != nil {
				return nil, fmt.Errorf("error parsing container resource requests: %s", err.Error())
			}
		}
		if req := np.Resources.Limits; req != nil {
			elasticsearchContainerLimits, err = parseResources(req)

			if err != nil {
				return nil, fmt.Errorf("error parsing container resource limits: %s", err.Error())
			}
		}
	}

	volumes := []apiv1.Volume{
	// {
	// 	Name: "sidecar-config",
	// 	VolumeSource: apiv1.VolumeSource{
	// 		ConfigMap: &apiv1.ConfigMapVolumeSource{
	// 			LocalObjectReference: apiv1.LocalObjectReference{
	// 				Name: nodePoolConfigMapName(c, np),
	// 			},
	// 		},
	// 	},
	// },
	}

	if np.State == nil ||
		np.State.Persistence == nil ||
		!np.State.Persistence.Enabled {
		volumes = append(volumes, apiv1.Volume{
			Name: "elasticsearch-data",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		})
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
					Args:            []string{"start"},
					Env: []apiv1.EnvVar{
						{
							Name:  "SERVICE",
							Value: clusterNodesServiceName(c),
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
						Requests: elasticsearchContainerRequests,
						Limits:   elasticsearchContainerLimits,
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
							ReadOnly:  true,
						},
					},
				},
			},
		},
	}, nil
}

func buildInitContainers(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) []apiv1.Container {
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

func buildNodePoolLabels(c *v1.ElasticsearchCluster, poolName string, roles ...string) map[string]string {
	labels := map[string]string{
		"app": "elasticsearch",
	}
	if poolName != "" {
		labels["pool"] = poolName
	}
	for _, role := range roles {
		labels[role] = "true"
	}
	return labels
}

func parseResources(rs *v1.ElasticsearchClusterResources_ResourceSet) (apiv1.ResourceList, error) {
	list := apiv1.ResourceList{}
	var err error
	var cpu, mem resource.Quantity

	if cpu, err = resource.ParseQuantity(rs.Cpu); err != nil {
		return list, fmt.Errorf("error parsing cpu specification '%s': %s", rs.Cpu, err.Error())
	}

	list[apiv1.ResourceCPU] = cpu

	if mem, err = resource.ParseQuantity(rs.Memory); err != nil {
		return list, fmt.Errorf("error parsing memory specification '%s': %s", rs.Memory, err.Error())
	}

	list[apiv1.ResourceMemory] = mem

	return list, nil
}

func clientNodesService(c *v1.ElasticsearchCluster) apiv1.Service {
	return apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            clientNodesServiceName(c),
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference(c)},
			Labels:          buildNodePoolLabels(c, "", "client"),
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
			Selector: buildNodePoolLabels(c, "", "client"),
		},
	}
}

func clientNodesServiceName(c *v1.ElasticsearchCluster) string {
	return fmt.Sprintf("%s", c.Name)
}

func clusterNodesService(c *v1.ElasticsearchCluster) apiv1.Service {
	return apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            clusterNodesServiceName(c),
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference(c)},
			Labels:          buildNodePoolLabels(c, "", "client", "data", "master"),
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
			Selector: buildNodePoolLabels(c, "", "client", "data", "master"),
		},
	}
}

func clusterNodesServiceName(c *v1.ElasticsearchCluster) string {
	return fmt.Sprintf("%s-cluster", c.Name)
}

func nodePoolConfigMapName(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) string {
	return fmt.Sprintf("%s-%s-config", c.Name, np.Name)
}
