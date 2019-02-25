package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/util/strings"
)

type BPFClient struct {
	restClient rest.Interface
	client     dynamic.ResourceInterface
}

func (i *BPFClient) List(opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	list, err := i.client.List(opts)
	return list, err
}

func (i *BPFClient) Get(name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	namespace, resname := strings.SplitQualifiedName(name)
	BPF := &unstructured.Unstructured{}
	err := i.restClient.Get().Namespace(namespace).Name(resname).Resource(BPFPlural).Do().Into(BPF)
	if err != nil {
		return nil, err
	}
	return BPF, nil
}

func (i *BPFClient) Delete(name string, opts *metav1.DeleteOptions, subresources ...string) error {
	return nil
}

func (i *BPFClient) DeleteCollection(deleteOptions *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}

func (i *BPFClient) Create(obj *unstructured.Unstructured, subresources ...string) (*unstructured.Unstructured, error) {
	// todos > create daemonset here?
	panic("not implemented")
}

func (i *BPFClient) Update(obj *unstructured.Unstructured, subresources ...string) (*unstructured.Unstructured, error) {
	snap := &unstructured.Unstructured{}
	err := i.restClient.Put().Namespace(obj.GetNamespace()).Resource(BPFPlural).Name(obj.GetName()).Body(obj).Do().Into(snap)
	return snap, err
}

func (i *BPFClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	res, err := i.restClient.Get().Prefix("watch").Resource(BPFPlural).Stream()
	if err != nil {
		return nil, err
	}

	return watch.NewStreamWatcher(&BPFDecoder{
		dec:   json.NewDecoder(res),
		close: res.Close,
	}), nil
}

func (i *BPFClient) UpdateStatus(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	panic("not implemented")
}

func (i *BPFClient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*unstructured.Unstructured, error) {
	panic("not implemented")
}

func NewBPFClient(dynamicClient dynamic.Interface, restclient rest.Interface) dynamic.ResourceInterface {
	resource := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: BPFPlural,
	})
	return &BPFClient{
		restClient: restclient,
		client:     resource,
	}
}

type BPFDecoder struct {
	dec   *json.Decoder
	close func() error
}

func (j *BPFDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object BPF
	}
	if err := j.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}

func (j *BPFDecoder) Close() {
	j.close()
}
