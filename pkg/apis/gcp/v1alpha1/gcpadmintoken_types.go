package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GCPAdminTokenSpec defines the desired state of GCPAdminToken
// +k8s:openapi-gen=true
type GCPAdminTokenSpec struct {
	// Token is the bearer token used to setup the initial GCP admin project and service account
	// You must grab a token using `gcloud auth print-access-token you@example.com`
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:validation:Required
	Token string `json:"token"`
}

// GCPAdminTokenStatus defines the observed state of GCPAdminToken
// +k8s:openapi-gen=true
type GCPAdminTokenStatus struct {
	// Verified checks that the token is ok and valid
	Verified bool `json:"verified"`
	// Status provides a overall status
	Status string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GCPAdminToken is the Schema for the gcpadmintokens API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=gcpadmintokens,scope=Namespaced
type GCPAdminToken struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPAdminTokenSpec   `json:"spec,omitempty"`
	Status GCPAdminTokenStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GCPAdminTokenList contains a list of GCPAdminToken
type GCPAdminTokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPAdminToken `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GCPAdminToken{}, &GCPAdminTokenList{})
}
