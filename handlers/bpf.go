package handlers

import (
	"github.com/leodido/bpf-operator/apis/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func NewBPFSharedInformer(bpfClient dynamic.ResourceInterface, queue workqueue.RateLimitingInterface) cache.SharedInformer {
	si := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				res, err := bpfClient.List(options)
				if err != nil {
					return nil, err
				}

				ro := &v1alpha1.BPFList{}
				ro.Items = make([]*v1alpha1.BPF, len(res.Items))
				runtime.DefaultUnstructuredConverter.FromUnstructured(res.UnstructuredContent(), ro)
				return ro, nil
			},
			WatchFunc: bpfClient.Watch,
		},
		&v1alpha1.BPF{},
		0,
	)

	si.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			queue.Add(obj)
		},
		UpdateFunc: func(oldobj, newobj interface{}) {
			queue.Add(newobj)
		},
		DeleteFunc: func(obj interface{}) {
			queue.Add(obj)
		},
	})
	return si
}
