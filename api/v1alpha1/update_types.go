package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UpdateSpec struct {
	Versioning UpdateVersioning `json:"versioning"`
}

type UpdateVersioning struct {
	Sources []UpdateSource `json:"sources"`
}

type UpdateSource struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Source  string `json:"source"`
	Version string `json:"version"`
}
type UpdateStatus struct {
	Conditions    []UpdateSource `json:"conditions"`
	Phase         string         `json:"phase"`
	SyncTimestamp string         `json:"syncTimestamp"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.versioning.sources[0].type`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.versioning.sources[0].version`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Synced",type=date,JSONPath=`.status.syncTimestamp`

// Update is the Schema for the updates API
type Update struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UpdateSpec   `json:"spec,omitempty"`
	Status UpdateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UpdateList contains a list of Update
type UpdateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Update `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Update{}, &UpdateList{})
}

func (c *Update) CheckStatus() *Update {
	c.Status.Phase = "HiHi"
	return c
}
