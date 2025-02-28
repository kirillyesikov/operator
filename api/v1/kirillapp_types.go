/*
Copyright 2025.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KirillAppSpec defines the desired state of KirillApp
type KirillAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Replicas int32                  `json:"replicas,omitempty"`
	Selector *metav1.LabelSelector  `json:"selector,omitempty"`
	Template corev1.PodTemplateSpec `json:"template,omitempty"`
	// Foo is an example field of KirillApp. Edit kirillapp_types.go to remove/update

}

// KirillAppStatus defines the observed state of KirillApp
type KirillAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	AvailableReplicas int32              `json:"availableReplicas"`
	ReadyReplicas     int32              `json:"readyReplicas"`
	PodNames          []string           `json:"podNames,omitempty"`
	Conditions        []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KirillApp is the Schema for the kirillapps API
type KirillApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              KirillAppSpec   `json:"spec,omitempty"`
	Status            KirillAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KirillAppList contains a list of KirillApp
type KirillAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KirillApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KirillApp{}, &KirillAppList{})
}
