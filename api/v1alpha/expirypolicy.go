package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type TargetKind struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

type ExpiryPolicySpec struct {
	Selector metav1.LabelSelector `json:"selector"`

	// +kubebuilder:validation:MinItems=1
	TargetKinds []TargetKind `json:"targetKinds"`


	TTL metav1.Duration `json:"ttl"`

	// +kubebuilder:default=false
	DryRun bool `json:"dryRun,omitempty"`
}

type ExpiryPolicyStatus struct {
	MatchedCount  int          `json:"matchedCount,omitempty"`
	DeletedCount  int          `json:"deletedCount,omitempty"`
	LastSweepTime *metav1.Time `json:"lastSweepTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="TTL",type=string,JSONPath=`.spec.ttl`
// +kubebuilder:printcolumn:name="Matched",type=integer,JSONPath=`.status.matchedCount`
// +kubebuilder:printcolumn:name="Deleted",type=integer,JSONPath=`.status.deletedCount`
type ExpiryPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExpiryPolicySpec   `json:"spec"`
	Status ExpiryPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ExpiryPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExpiryPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExpiryPolicy{}, &ExpiryPolicyList{})
}