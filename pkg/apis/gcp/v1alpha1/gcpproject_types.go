package v1alpha1

import (
	core "github.com/appvia/hub-apis/pkg/apis/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GCPProjectSpec defines the desired state of GCPProject
// +k8s:openapi-gen=true
type GCPProjectSpec struct {
	// ProjectId is the GCP project ID
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:validation:Required
	ProjectId string `json:"projectId"`
	// ProjectName is the GCP project name
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:validation:Required
	ProjectName string `json:"projectName"`
	// ParentType is the type of parent this project has
	// Valid types are: "organization", "folder", and "project"
	// +kubebuilder:validation:Enum=organization;folder;project
	// +kubebuilder:validation:Required
	ParentType string `json:"parentType"`
	// ParentId is the type specific ID of the parent this project has
	// +kubebuilder:validation:Required
	ParentId string `json:"parentId"`
	// BillingAccountName is the resource name of the billing account associated with the project
	// +kubebuilder:validation:Required
	// +k8s:openapi-gen=false
	BillingAccountName string `json:"billingAccountName"`
	// ServiceAccountName is the name used when creating the service account
	// e.g. `hub-admin`
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:validation:Required
	ServiceAccountName string `json:"serviceAccountName"`
	// GCPCredentials is a reference to the gcp credentials object to use
	// +kubebuilder:validation:Required
	// +k8s:openapi-gen=false
	Use core.Ownership `json:"use"`
}

// GCPProjectStatus defines the observed state of GCPProject
// +k8s:openapi-gen=true
type GCPProjectStatus struct {
	// Status provides a overall status
	Status core.Status `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GCPProject is the Schema for the gcpprojects API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=gcpprojects,scope=Namespaced
type GCPProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPProjectSpec   `json:"spec,omitempty"`
	Status GCPProjectStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GCPProjectList contains a list of GCPProject
type GCPProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPProject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GCPProject{}, &GCPProjectList{})
}
