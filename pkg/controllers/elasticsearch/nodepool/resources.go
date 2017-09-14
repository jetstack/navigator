package nodepool

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	apps "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/util"
)

const (
	sharedVolumeName      = "shared"
	sharedVolumeMountPath = "/shared"

	esDataVolumeName      = "elasticsearch-data"
	esDataVolumeMountPath = "/usr/share/elasticsearch/data"

	esConfigVolumeName      = "config"
	esConfigVolumeMountPath = "/config"
)

func nodePoolStatefulSet(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) (*apps.StatefulSet, error) {
	statefulSetName := util.NodePoolResourceName(c, np)

	elasticsearchPodTemplate, err := elasticsearchPodTemplateSpec(statefulSetName, c, np)

	if err != nil {
		return nil, fmt.Errorf("error building elasticsearch container: %s", err.Error())
	}

	nodePoolHash, err := nodePoolHash(np)

	if err != nil {
		return nil, fmt.Errorf("error hashing node pool object: %s", err.Error())
	}

	ss := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            statefulSetName,
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(c)},
			Annotations: map[string]string{
				util.NodePoolHashAnnotationKey: nodePoolHash,
			},
			Labels: util.NodePoolLabels(c, np.Name, np.Roles...),
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    util.Int32Ptr(int32(np.Replicas)),
			ServiceName: statefulSetName,
			Selector: &metav1.LabelSelector{
				MatchLabels: util.NodePoolLabels(c, np.Name, np.Roles...),
			},
			Template: *elasticsearchPodTemplate,
		},
	}

	if np.Persistence.Enabled {
		volumeClaimTemplateAnnotations := map[string]string{}
		volumeResourceRequests := apiv1.ResourceList{}

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

		ss.Spec.VolumeClaimTemplates = []apiv1.PersistentVolumeClaim{
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
		}
	}

	return ss, nil
}

func elasticsearchPodTemplateSpec(controllerName string, c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) (*apiv1.PodTemplateSpec, error) {
	volumes := []apiv1.Volume{
		apiv1.Volume{
			Name: sharedVolumeName,
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		},
		apiv1.Volume{
			Name: esConfigVolumeName,
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: util.ConfigMapName(c),
					},
				},
			},
		},
	}

	if !np.Persistence.Enabled {
		volumes = append(volumes, apiv1.Volume{
			Name: esDataVolumeName,
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		})
	}

	rolesBytes, err := json.Marshal(np.Roles)

	if err != nil {
		return nil, fmt.Errorf("error marshaling roles: %s", err.Error())
	}

	pluginsBytes := []byte("[]")
	if len(c.Spec.Plugins) > 0 {
		pluginsBytes, err = json.Marshal(c.Spec.Plugins)

		if err != nil {
			return nil, fmt.Errorf("error marshaling plugins: %s", err.Error())
		}
	}

	return &apiv1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: util.NodePoolLabels(c, np.Name, np.Roles...),
		},
		Spec: apiv1.PodSpec{
			TerminationGracePeriodSeconds: util.Int64Ptr(1800),
			// TODO
			ServiceAccountName: "",
			NodeSelector:       np.NodeSelector,
			SecurityContext: &apiv1.PodSecurityContext{
				FSGroup: util.Int64Ptr(c.Spec.Image.FsGroup),
			},
			Volumes:        volumes,
			InitContainers: buildInitContainers(c, np),
			Containers: []apiv1.Container{
				{
					Name:            "elasticsearch",
					Image:           c.Spec.Image.Repository + ":" + c.Spec.Image.Tag,
					ImagePullPolicy: apiv1.PullPolicy(c.Spec.Image.PullPolicy),
					Command: []string{
						"/bin/sh",
						"-c",
						fmt.Sprintf(`#!/bin/sh
exec %s/pilot \
  start \
  --podName=$(POD_NAME) \
  --clusterURL=$(CLUSTER_URL) \
  --namespace=$(NAMESPACE) \
  --controllerKind=StatefulSet \
  --plugins='%s' \
  --roles='%s' \
  --controllerName='%s'`, sharedVolumeMountPath, string(pluginsBytes), string(rolesBytes), controllerName),
					},
					Env: []apiv1.EnvVar{
						// TODO: Tidy up generation of discovery & client URLs
						{
							Name:  "DISCOVERY_SERVICE",
							Value: util.DiscoveryServiceName(c),
						},
						{
							Name:  "CLUSTER_URL",
							Value: "http://" + util.ClientServiceName(c) + ":9200",
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
							Name:      esDataVolumeName,
							MountPath: esDataVolumeMountPath,
							ReadOnly:  false,
						},
						{
							Name:      sharedVolumeName,
							MountPath: sharedVolumeMountPath,
							ReadOnly:  false,
						},
						{
							Name:      esConfigVolumeName,
							MountPath: esConfigVolumeMountPath,
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}, nil
}

func buildInitContainers(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) []apiv1.Container {
	containers := make([]apiv1.Container, len(c.Spec.Sysctl)+1)
	containers[0] = apiv1.Container{
		Name:            "install-pilot",
		Image:           fmt.Sprintf("%s:%s", c.Spec.Pilot.Repository, c.Spec.Pilot.Tag),
		ImagePullPolicy: apiv1.PullPolicy(c.Spec.Pilot.PullPolicy),
		Command:         []string{"cp", "/pilot", fmt.Sprintf("%s/pilot", sharedVolumeMountPath)},
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
	for i, sysctl := range c.Spec.Sysctl {
		containers[i+1] = apiv1.Container{
			Name:            fmt.Sprintf("tune-sysctl-%d", i),
			Image:           "busybox:latest",
			ImagePullPolicy: apiv1.PullIfNotPresent,
			SecurityContext: &apiv1.SecurityContext{
				Privileged: util.BoolPtr(true),
			},
			Command: []string{
				"sysctl", "-w", sysctl,
			},
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("10m"),
					apiv1.ResourceMemory: resource.MustParse("8Mi"),
				},
			},
		}
	}
	return containers
}

func nodePoolHash(np *v1alpha1.ElasticsearchClusterNodePool) (string, error) {
	d, err := json.Marshal(np)
	if err != nil {
		return "", err
	}
	hasher := md5.New()
	hasher.Write(d)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
