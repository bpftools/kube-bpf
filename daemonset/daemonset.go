package daemonset

import (
	"fmt"
	"path"

	resources "github.com/bpftools/kube-bpf/apis/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/utils/pointer"
)

const bpfProgramAbsolutePath = "/bpf"

type DaemonSet struct {
	resource *resources.BPF
	client   tappsv1.DaemonSetInterface
}

func New(res *resources.BPF, appsv1Client tappsv1.AppsV1Interface) (*DaemonSet, error) {
	if appsv1Client == nil {
		return nil, fmt.Errorf("missing AppsV1 client")
	}
	if res == nil {
		return nil, fmt.Errorf("missing BPF resource")
	}
	if res.Spec.Program.ValueFrom == nil {
		return nil, fmt.Errorf("missing BPF program in .Spec.Program.ValueFrom")
	}
	if res.Spec.Program.ValueFrom.ConfigMapKeyRef == nil {
		return nil, fmt.Errorf("missing BPF program in .Spec.Program.ValueFrom.ConfigMapKeyRef")
	}
	if res.Spec.Program.ValueFrom.ConfigMapKeyRef.Name == "" {
		return nil, fmt.Errorf("missing BPF program in.Spec.Program.ValueFrom.ConfigMapKeyRef.Name")
	}
	if res.Spec.Program.ValueFrom.ConfigMapKeyRef.Key == "" {
		return nil, fmt.Errorf("missing BPF program in.Spec.Program.ValueFrom.ConfigMapKeyRef.Key")
	}

	// Fallback to default namesace when it is missing
	if res.Namespace == "" {
		res.Namespace = "default"
	}

	s := &DaemonSet{
		resource: res,
		client:   appsv1Client.DaemonSets(res.Namespace),
	}
	return s, nil
}

func (s *DaemonSet) Create() (*appsv1.DaemonSet, error) {
	if s.resource.ObjectMeta.Labels == nil {
		s.resource.ObjectMeta.Labels = map[string]string{}
	}
	s.resource.ObjectMeta.Labels["bpf.sh/bpf-origin-uid"] = string(s.resource.ObjectMeta.UID)

	daemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("bpf-%s", s.resource.Name),
			Namespace:   s.resource.Namespace,
			Labels:      s.resource.ObjectMeta.Labels,
			Annotations: s.resource.ObjectMeta.Annotations,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "runbpf",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "runbpf",
					},
				},
				Spec: corev1.PodSpec{
					HostNetwork: true, // --net="host" // todos > this means two BPF resources cannot run together (since the port will be occupied)
					HostPID:     true, // --pid="host"
					// NodeSelector: map[string]string{}, // todos > node filtering/selection?
					Containers: []corev1.Container{
						{
							Name:  "runbpf",
							Image: "leodido/runbpf:latest",
							Args: []string{
								path.Join(bpfProgramAbsolutePath, s.resource.Spec.Program.ValueFrom.ConfigMapKeyRef.Key),
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9387,
								},
							},
							ImagePullPolicy: "IfNotPresent", // todos > use Always here?
							SecurityContext: &corev1.SecurityContext{
								Privileged: pointer.BoolPtr(true),
							},
							Env: []corev1.EnvVar{
								{
									Name: "NODENAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "sys",
									MountPath: "/sys",
									ReadOnly:  true,
								},
								corev1.VolumeMount{
									Name:      "program",
									MountPath: bpfProgramAbsolutePath,
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "sys",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/sys",
								},
							},
						},
						corev1.Volume{
							Name: "program",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: s.resource.Spec.Program.ValueFrom.ConfigMapKeyRef.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return s.client.Create(daemonSet)
}
