// Copyright 2025 Rina
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	"context"
	"fmt"
	"log/slog"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
	"rina.icu/hoshino/internal/util"
	"rina.icu/hoshino/store"
)

type ContainerManager struct {
	K8SClient *kubernetes.Clientset
	Store     *store.Store
}

type ContainerCreatePayload struct {
	Image                   string `json:"image" validate:"required"`
	MemoryLimit             string `json:"memory_limit" default:"128Mi"`
	CPULimit                string `json:"cpu_limit" default:"100m"`
	StorageLimit            string `json:"storage_limit" default:"1Gi"`
	ExposedPort             int    `json:"exposed_port"`
	RegistryAccessTokenUUID string `json:"registry_access_token_uuid"`
}

type ContainerInfo struct {
	Identifier string            `json:"-"`
	Labels     map[string]string `json:"-"`
	Flag       string            `json:"-"`
}

func (cm *ContainerManager) CreateContainer(payload *ContainerCreatePayload, info *ContainerInfo, store *store.Store) (string, string, error) {
	hash := util.SHA256(info.Identifier)[:16]
	name := "container-" + hash
	// Create a container in k8s

	info.Labels["identifier"] = info.Identifier

	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: info.Labels,
		},
		Spec: appv1.DeploymentSpec{
			Replicas: ptr.To[int32](1),
			Selector: &metav1.LabelSelector{
				MatchLabels: info.Labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: info.Labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "challenge",
							Image: payload.Image,
							Env: []v1.EnvVar{
								{
									Name:  "HOSHINO_FLAG",
									Value: info.Flag,
								},
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:              resource.MustParse(payload.CPULimit),
									v1.ResourceMemory:           resource.MustParse(payload.MemoryLimit),
									v1.ResourceEphemeralStorage: resource.MustParse(payload.StorageLimit),
								},
							},
						},
					},
				},
			},
		},
	}

	deploymentObj, err := cm.K8SClient.AppsV1().Deployments("challenge-containers").Create(context.Background(), deployment, metav1.CreateOptions{})

	if err != nil {
		// handle error
		slog.Error(err.Error())
		return "", "", err
	}

	slog.Info(fmt.Sprintf("Deployment created: %s", deploymentObj.Name))

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name + "-service",
			Labels: info.Labels,
		},
		Spec: v1.ServiceSpec{
			Selector: info.Labels,
			Ports: []v1.ServicePort{
				{
					Port:       int32(payload.ExposedPort),
					TargetPort: intstr.FromInt(payload.ExposedPort),
				},
			},
		},
	}

	serviceObj, err := cm.K8SClient.CoreV1().Services("challenge-containers").Create(context.Background(), service, metav1.CreateOptions{})
	if err != nil {
		// handle error
		slog.Error(err.Error())
		return "", "", err
	}

	slog.Info(fmt.Sprintf("Service created: %s, %d", serviceObj.Name, serviceObj.Spec.Ports[0].NodePort))

	containerUUID := util.UUID()

	pods, err := cm.K8SClient.CoreV1().Pods("challenge-containers").List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("identifier=%s", info.Identifier),
	})

	if err != nil {
		// handle error
		slog.Error(err.Error())
		return "", "", err
	}

	if len(pods.Items) == 0 {
		slog.Warn("No pod found")
		return "", "", fmt.Errorf("No pod found")
	}

	pod := pods.Items[0]
	nodeName := pod.Spec.NodeName
	nodeDomain := ""
	node, err := cm.K8SClient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})

	if domain, ok := node.Labels["node-domain"]; ok {
		nodeDomain = domain
	} else {
		slog.Warn("Node domain not found, using default domain")
		nodeDomain = store.GetSettingString("node_domain")
	}

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name + "-ingress",
			Labels: info.Labels,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: fmt.Sprintf("%s.%s", containerUUID, nodeDomain),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: func() *networkingv1.PathType { pt := networkingv1.PathTypePrefix; return &pt }(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: name + "-service",
											Port: networkingv1.ServiceBackendPort{
												Number: int32(payload.ExposedPort),
											},
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
	ingressObj, err := cm.K8SClient.NetworkingV1().Ingresses("challenge-containers").Create(context.Background(), ingress, metav1.CreateOptions{})
	if err != nil {
		// handle error
		slog.Error(err.Error())
		return "", "", err
	}

	slog.Info(fmt.Sprintf("Ingress created: %s", ingressObj.Name))
	return containerUUID, nodeDomain, nil
}

func (cm *ContainerManager) DisposeContainer(identifier string) error {
	hash := util.SHA256(identifier)[:16]
	name := "container-" + hash
	// Delete a container in k8s
	err := cm.K8SClient.AppsV1().Deployments("challenge-containers").Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		// handle error
		slog.Error(err.Error())
		return err
	}

	err = cm.K8SClient.CoreV1().Services("challenge-containers").Delete(context.Background(), name+"-service", metav1.DeleteOptions{})
	if err != nil {
		// handle error
		slog.Error(err.Error())
		return err
	}

	err = cm.K8SClient.NetworkingV1().Ingresses("challenge-containers").Delete(context.Background(), name+"-ingress", metav1.DeleteOptions{})
	if err != nil {
		// handle error
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (cm *ContainerManager) CreateChallengeContainer(c *store.Challenge, u *store.User, flag string, s *store.Store) (string, string, error) {
	payload := &ContainerCreatePayload{
		Image:                   c.Image.Name,
		MemoryLimit:             c.Image.MemoryLimit,
		CPULimit:                c.Image.CPULimit,
		StorageLimit:            c.Image.StorageLimit,
		ExposedPort:             c.Image.ExposedPort,
		RegistryAccessTokenUUID: c.Image.RegistryAccessTokenUUID,
	}

	info := &ContainerInfo{
		Identifier: c.UUID + u.UUID,
		Labels: map[string]string{
			"challenge": c.UUID,
			"user":      u.UUID,
		},
		Flag: flag,
	}

	return cm.CreateContainer(payload, info, s)
}

func (cm *ContainerManager) DisposeChallengeContainer(c *store.Challenge, u *store.User) error {
	return cm.DisposeContainer(c.UUID + u.UUID)
}
