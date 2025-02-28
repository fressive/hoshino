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

func (cm *ContainerManager) CreateChallengeContainer(c *store.Challenge, u *store.User, flag string, store *store.Store) (string, error) {
	hash := util.SHA256(c.UUID + u.UUID)[:16]
	name := "container-" + hash
	// Create a container in k8s
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"challenge": c.UUID,
				"user":      u.UUID,
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: ptr.To[int32](1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"challenge": c.UUID,
					"user":      u.UUID,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"challenge": c.UUID,
						"user":      u.UUID,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "challenge",
							Image: c.Image.Name,
							Env: []v1.EnvVar{
								{
									Name:  "HOSHINO_FLAG",
									Value: flag,
								},
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:                    resource.MustParse(c.Image.CPULimit),
									v1.ResourceMemory:                 resource.MustParse(c.Image.MemoryLimit),
									v1.ResourceLimitsEphemeralStorage: resource.MustParse(c.Image.StorageLimit),
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
		return "", err
	}

	slog.Info(fmt.Sprintf("Deployment created: %s", deploymentObj.Name))

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-service",
			Labels: map[string]string{
				"challenge": c.UUID,
				"user":      u.UUID,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"challenge": c.UUID,
				"user":      u.UUID,
			},
			Ports: []v1.ServicePort{
				{
					Port:       int32(c.ExposedPort),
					TargetPort: intstr.FromInt(c.ExposedPort),
				},
			},
		},
	}

	serviceObj, err := cm.K8SClient.CoreV1().Services("challenge-containers").Create(context.Background(), service, metav1.CreateOptions{})
	if err != nil {
		// handle error
		slog.Error(err.Error())
		return "", err
	}

	slog.Info(fmt.Sprintf("Service created: %s, %d", serviceObj.Name, serviceObj.Spec.Ports[0].NodePort))

	containerUUID := util.UUID()

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-ingress",
			Labels: map[string]string{
				"challenge": c.UUID,
				"user":      u.UUID,
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: fmt.Sprintf("%s.%s", containerUUID, store.GetSettingString("node_domain")),
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
												Number: int32(c.ExposedPort),
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
		return "", err
	}

	slog.Info(fmt.Sprintf("Ingress created: %s", ingressObj.Name))
	return containerUUID, nil
}

func (cm *ContainerManager) DisposeChallengeContainer(c *store.Challenge, u *store.User) error {
	hash := util.SHA256(c.UUID + u.UUID)[:16]
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
