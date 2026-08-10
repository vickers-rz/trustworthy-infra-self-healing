// Package v1alpha1 contains the first experimental InfraHeal Kubernetes API.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	GroupVersion = schema.GroupVersion{Group: "infraheal.io", Version: "v1alpha1"}
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}
	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(&HealingPolicy{}, &HealingPolicyList{})
}

// APIVersion returns the canonical group/version string for this API.
func APIVersion() string { return GroupVersion.String() }

// TypeMetaFor is a small helper used by tests and examples.
func TypeMetaFor(kind string) metav1.TypeMeta {
	return metav1.TypeMeta{APIVersion: APIVersion(), Kind: kind}
}
