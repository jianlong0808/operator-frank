/*
Copyright 2022.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//+kubebuilder:validation:Enum=backend;frontend

// Pdl is an example type. Edit frank_types.go to remove/update
type Pdl string

const (
	Backend  Pdl = "backend"
	Frontend Pdl = "frontend"
)

// FrankSpec defines the desired state of Frank
type FrankSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Pdl *Pdl `json:"pdl"`
	//+kubebuilder:validation:MinLength=0
	// Image,Replica is an example field of Frank. Edit frank_types.go to remove/update
	Image *string `json:"image,omitempty"`
	//+kubebuilder:validation:Minimum=0
	Replica *int32 `json:"replica,omitempty"`
}

// FrankStatus defines the observed state of Frank
type FrankStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	//+kubebuilder:validation:Minimum=0
	RealReplica int32 `json:"realReplica,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replica,statuspath=.status.realReplica,selectorpath=.spec.selector
//+kubebuilder:printcolumn:name="RealReplica",type=integer,JSONPath=`.status.realReplica`
//+kubebuilder:printcolumn:name="Pdl",type=string,priority=1,JSONPath=`.spec.pdl`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`

// Frank is the Schema for the franks API
type Frank struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrankSpec   `json:"spec,omitempty"`
	Status FrankStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FrankList contains a list of Frank
type FrankList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Frank `json:"items"`
}

func init() {
	//将这个 Go 类型添加到 API 组中。这允许我们将这个 API 组中的类型可以添加到任何Scheme。
	SchemeBuilder.Register(&Frank{}, &FrankList{})
}
