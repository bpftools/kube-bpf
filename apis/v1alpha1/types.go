package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	BPFKind     = "BPF"
	BPFResource = "bpf"
	BPFPlural   = "bpfs"
	Group       = "bpf.sh"
	Version     = "v1alpha1"
)

type BPF struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BPFSpec `json:"spec"`
}

func (b *BPF) GetObjectKind() schema.ObjectKind {
	panic("not implemented")
}

func (b *BPF) DeepCopyObject() runtime.Object {
	panic("not implemented")
}

type BPFSpec struct {
	Program Program `json:"program"`
}

func (b *BPFSpec) GetObjectKind() schema.ObjectKind {
	panic("not implemented")
}

func (b *BPFSpec) DeepCopyObject() runtime.Object {
	panic("not implemented")
}

type BPFList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []*BPF `json:"items"`
}

func (l *BPFList) GetObjectKind() schema.ObjectKind {
	panic("not implemented")
}

func (l *BPFList) DeepCopyObject() runtime.Object {
	panic("not implemented")
}

type Program struct {
	// String value for program
	// Defaults to ""
	// +optional
	Value string `json:"value,omitempty"`
	// Source for the program value. Cannot be used if value is not empty.
	// +optional
	ValueFrom *ProgramSource `json:"valueFrom,omitempty"`
}

// ProgramSource represents a source for the value of an Program.
type ProgramSource struct {
	// Selects a key of a ConfigMap in the BPF's namespace
	// +optional
	ConfigMapKeyRef *corev1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	// Selects a key of a secret in the BPF's namespace
	// +optional
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`

	// TODO(us): VolumeKeyRef
}
