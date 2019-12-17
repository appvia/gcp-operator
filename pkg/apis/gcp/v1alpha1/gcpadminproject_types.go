package v1alpha1

import (
	core "github.com/appvia/hub-apis/pkg/apis/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GCPAdminProjectSpec defines the desired state of GCPAdminProject
// +k8s:openapi-gen=true
type GCPAdminProjectSpec struct {
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
	// e.g. `012345-567890-ABCDEF`
	// +kubebuilder:validation:Required
	// +k8s:openapi-gen=false
	BillingAccountName string `json:"billingAccountName"`
	// GCPCredentials is a reference to the gcp credentials object to use
	// +kubebuilder:validation:Required
	// +k8s:openapi-gen=false
	Use core.Ownership `json:"use"`
}

// GCPAdminProjectStatus defines the observed state of GCPAdminProject
// +k8s:openapi-gen=true
type GCPAdminProjectStatus struct {
	// Status provides a overall status
	Status core.Status `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GCPAdminProject is the Schema for the gcpadminprojects API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=gcpadminprojects,scope=Namespaced
type GCPAdminProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPAdminProjectSpec   `json:"spec,omitempty"`
	Status GCPAdminProjectStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GCPAdminProjectList contains a list of GCPAdminProject
type GCPAdminProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPAdminProject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GCPAdminProject{}, &GCPAdminProjectList{})
}
