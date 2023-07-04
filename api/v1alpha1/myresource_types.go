/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MyResourceSpec defines the desired state of MyResource
type MyResourceSpec struct {
	App        AppConfig         `json:"app"`
	Service    ServiceConfig     `json:"service"`
	Monitoring *MonitoringConfig `json:"monitoring,omitempty"`
}

type AppConfig struct {
	// +kubebuilder:validation:Required
	//字段至少为1  如果字段的值是其类型的零值 并且在json中被标记为omitempty 那么该字段将不会被发送到API服务器，也就不会被验证
	//如果你想要字段既要满足最小值为1，又要必填，Required是必须的
	// +kubebuilder:validation:Minimum=1
	Size int32 `json:"size,omitempty"`
	// +kubebuilder:validation:Required
	// #+kubebuilder:validation:Pattern=`^[a-z0-9]+(\.[a-z0-9]+)*(:[0-9a-zA-Z_\.]*)?$`
	Image            string                        `json:"image"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Envs             []corev1.EnvVar               `json:"envs,omitempty"`
	Resources        corev1.ResourceRequirements   `json:"resources,omitempty"`
	Ports            []corev1.ContainerPort        `json:"ports"`
}

type ServiceConfig struct {
	Type corev1.ServiceType `json:"type,omitempty"`
	// +kubebuilder:validation:MinItems=1
	//数组至少有一个元素
	Ports []corev1.ServicePort `json:"ports"`
}

type MonitoringConfig struct {
	Enabled      bool           `json:"enabled"`
	Image        string         `json:"image,omitempty"`
	ExporterPort int32          `json:"exporterPort,omitempty"`
	Service      *ServiceConfig `json:"service,omitempty"`
}

// MyResourceStatus defines the observed state of MyResource
type MyResourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase      string             `json:"phase"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MyResource is the Schema for the myresources API
type MyResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MyResourceSpec `json:"spec,omitempty"`

	Status MyResourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MyResourceList contains a list of MyResource
type MyResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MyResource{}, &MyResourceList{})
}
