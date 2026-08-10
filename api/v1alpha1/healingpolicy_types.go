package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ObservationModeObserve = "Observe"

	ConditionTargetResolved  = "TargetResolved"
	ConditionObservedHealthy = "ObservedHealthy"
	ConditionObserveOnly      = "ObserveOnly"
)

// TargetReference identifies the workload observed by a HealingPolicy.
// The MVP intentionally supports only apps/v1 Deployments.
type TargetReference struct {
	Namespace string `json:"namespace,omitempty"`

	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// HealingPolicySpec declares what should be observed. It does not grant
// remediation authority and contains no mutation-capable action fields.
type HealingPolicySpec struct {
	Target TargetReference `json:"target"`

	// Mode is deliberately restricted to Observe in the first controller.
	// +kubebuilder:default:=Observe
	// +kubebuilder:validation:Enum=Observe
	Mode string `json:"mode,omitempty"`

	// ObserveIntervalSeconds controls the periodic read-only reconciliation.
	// +kubebuilder:default:=30
	// +kubebuilder:validation:Minimum=5
	// +kubebuilder:validation:Maximum=3600
	ObserveIntervalSeconds int32 `json:"observeIntervalSeconds,omitempty"`
}

// HealingPolicyStatus is an evidence-like observation snapshot of the target.
// It records what the controller saw without mutating the target workload.
type HealingPolicyStatus struct {
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	TargetFound        bool  `json:"targetFound,omitempty"`

	DesiredReplicas   int32 `json:"desiredReplicas,omitempty"`
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
	ReadyReplicas     int32 `json:"readyReplicas,omitempty"`
	UpdatedReplicas   int32 `json:"updatedReplicas,omitempty"`

	LastObservedTime          *metav1.Time       `json:"lastObservedTime,omitempty"`
	LastEvidenceID            string             `json:"lastEvidenceID,omitempty"`
	LastEvidenceDigestSHA256  string             `json:"lastEvidenceDigestSHA256,omitempty"`
	Conditions                []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=hp
// +kubebuilder:printcolumn:name="Target",type=string,JSONPath=`.spec.target.name`
// +kubebuilder:printcolumn:name="Found",type=boolean,JSONPath=`.status.targetFound`
// +kubebuilder:printcolumn:name="Available",type=integer,JSONPath=`.status.availableReplicas`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.status.desiredReplicas`
// HealingPolicy defines a read-only observation policy for infrastructure.
type HealingPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HealingPolicySpec   `json:"spec,omitempty"`
	Status HealingPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// HealingPolicyList contains a list of HealingPolicy objects.
type HealingPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HealingPolicy `json:"items"`
}
