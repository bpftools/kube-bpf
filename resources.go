package bpfoperator

import (
	"github.com/bpftools/bpf-operator/apis/v1alpha1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BPFCRDName is the name of the custom resource definition for BPFs
const BPFCRDName = "bpfs.bpf.sh"

var crds = []*extensionsobj.CustomResourceDefinition{
	// BPFs
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:   BPFCRDName,
			Labels: map[string]string{},
		},
		Spec: extensionsobj.CustomResourceDefinitionSpec{
			Group:   v1alpha1.Group,
			Version: v1alpha1.Version,
			Scope:   extensionsobj.NamespaceScoped,
			Names: extensionsobj.CustomResourceDefinitionNames{
				Plural: v1alpha1.BPFPlural,
				Kind:   v1alpha1.BPFKind,
			},
		},
	},
}
