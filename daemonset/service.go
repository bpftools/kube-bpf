package daemonset

import (
	"fmt"

	resources "github.com/bpftools/kube-bpf/apis/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Service struct {
	resource *resources.BPF
	client   tcorev1.ServiceInterface
}

func NewService(res *resources.BPF, corev1Client tcorev1.CoreV1Interface) (*Service, error) {
	if corev1Client == nil {
		return nil, fmt.Errorf("missing CoreV1 client")
	}

	// Fallback to default namesace when it is missing
	if res.Namespace == "" {
		res.Namespace = "default"
	}

	s := &Service{
		resource: res,
		client:   corev1Client.Services(res.Namespace),
	}
	return s, nil
}

func (s *Service) Create() (*corev1.Service, error) {
	if s.resource.ObjectMeta.Labels == nil {
		s.resource.ObjectMeta.Labels = map[string]string{}
	}
	s.resource.ObjectMeta.Labels["bpf.sh/bpf-origin-uid"] = string(s.resource.ObjectMeta.UID)

	appName := fmt.Sprintf("bpf-%s", s.resource.Name)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("bpf-%s", s.resource.Name),
			Namespace:   s.resource.Namespace,
			Labels:      s.resource.ObjectMeta.Labels,
			Annotations: s.resource.ObjectMeta.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port: 9387,
				},
			},
			Selector: map[string]string{
				"app": appName,
			},
			Type: "ClusterIP",
		},
	}

	return s.client.Create(service)
}
