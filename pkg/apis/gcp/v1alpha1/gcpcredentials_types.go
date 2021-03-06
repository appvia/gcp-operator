package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GCPCredentialsSpec defines the desired state of GCPCredentials
// +k8s:openapi-gen=true
type GCPCredentialsSpec struct {
	// Key is the credential used to create GCP projects
	// You must create a service account with resourcemanager.projectCreator
	// and billing.user roles at the organization level and use the JSON payload here
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:validation:Required
	Key string `json:"key"`
	// ProjectId is the GCP project ID these credentials belong to
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:validation:Required
	ProjectId string `json:"projectId"`
	// Organization is the GCP org you wish the projects to reside within
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:validation:Required
	OrganizationId string `json:"organizationId"`
}

// GCPCredentialsStatus defines the observed state of GCPCredentials
// +k8s:openapi-gen=true
type GCPCredentialsStatus struct {
	// Verified checks that the credentials are ok and valid
	Verified bool `json:"verified,omitempty"`
	// Status provides a overall status
	Status string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GCPCredentials is the Schema for the gcpcredentials API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=gcpcredentials,scope=Namespaced
type GCPCredentials struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPCredentialsSpec   `json:"spec,omitempty"`
	Status GCPCredentialsStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GCPCredentialsList contains a list of GCPCredentials
type GCPCredentialsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPCredentials `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GCPCredentials{}, &GCPCredentialsList{})
}
