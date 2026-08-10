// Code generated-style deepcopy helpers for the experimental API.
// This file is maintained manually until controller-gen becomes part of the build.
package v1alpha1

import runtime "k8s.io/apimachinery/pkg/runtime"

func (in *HealingPolicy) DeepCopyInto(out *HealingPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

func (in *HealingPolicy) DeepCopy() *HealingPolicy {
	if in == nil {
		return nil
	}
	out := new(HealingPolicy)
	in.DeepCopyInto(out)
	return out
}

func (in *HealingPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *HealingPolicyList) DeepCopyInto(out *HealingPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]HealingPolicy, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *HealingPolicyList) DeepCopy() *HealingPolicyList {
	if in == nil {
		return nil
	}
	out := new(HealingPolicyList)
	in.DeepCopyInto(out)
	return out
}

func (in *HealingPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *HealingPolicyStatus) DeepCopyInto(out *HealingPolicyStatus) {
	*out = *in
	if in.LastObservedTime != nil {
		out.LastObservedTime = in.LastObservedTime.DeepCopy()
	}
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		copy(out.Conditions, in.Conditions)
	}
}
