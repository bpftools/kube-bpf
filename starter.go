package kubebpf

import (
	"log"
	"sync"
	"time"

	"github.com/bpftools/kube-bpf/apis/v1alpha1"
	"github.com/bpftools/kube-bpf/controller"
	"github.com/bpftools/kube-bpf/handlers"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

// Config is the configuration struct for the operator
type Config struct {
	KubeConfig string
	Labels     map[string]string
}

// Operator reacts to changes to the desired state on kubernetes
// and tries to reconcile those changes to the actual state by calling the controllers.
type Operator struct {
	config      Config
	kubeClient  *kubernetes.Clientset
	clientSet   clientset.Interface
	Controllers []controller.Controller
	logger      *zap.Logger
}

func (o *Operator) runInformers(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	for _, v := range o.Controllers {
		wg.Add(1)
		// run only one instance of for each controller for now
		// todos > make this parametric?
		go v.Run(1, stopCh, wg)
	}
}

// New creates a new operator
func New(options Config, logger *zap.Logger) *Operator {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	if options.KubeConfig != "" {
		rules.ExplicitPath = options.KubeConfig
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
	if err != nil {
		logger.Fatal("unable to load non interactive client config", zap.Error(err))
	}

	config.GroupVersion = &schema.GroupVersion{
		Group:   v1alpha1.Group,
		Version: v1alpha1.Version,
	}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	kubeClient := kubernetes.NewForConfigOrDie(config)
	cs := clientset.NewForConfigOrDie(config)
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Fatal("unable to create the dynamic client for the given config", zap.Error(err))
	}

	operator := &Operator{
		config:     options,
		kubeClient: kubeClient,
		clientSet:  cs,
		logger:     logger,
	}

	restclient, err := rest.RESTClientFor(config)
	if err != nil {
		logger.Fatal("unable to create the rest client for the given config", zap.Error(err))
	}

	// BPF controller
	bpfWq := workqueue.NewNamedRateLimitingQueue(exponentialRateLimiter(), "BPF")
	bpfClient := v1alpha1.NewBPFClient(dynamicClient, restclient)
	bpfInformer := handlers.NewBPFSharedInformer(bpfClient, bpfWq)
	bpfController := controller.NewBPF(bpfInformer, bpfWq, kubeClient)
	bpfController.WithLogger(logger)

	// Controllers registration
	operator.Controllers = []controller.Controller{bpfController}
	return operator
}

// Run executes the operator loop
func (o *Operator) Run(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	crdClient := o.clientSet.ApiextensionsV1beta1().CustomResourceDefinitions()
	for _, crd := range crds {
		if _, err := crdClient.Create(crd); err != nil && !apierrors.IsAlreadyExists(err) {
			log.Fatal("unable to create the crd", zap.String("name", crd.GetName()), zap.Error(err))
		}
	}

	go o.runInformers(stopCh, wg)
}

// exponentialRateLimiter rate limiter
// applies an individual and an overeall rate limit
func exponentialRateLimiter() workqueue.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		// Exponential failure.
		// For each item retry with an increasingly wait of one second and drop after 1000 seconds
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 1000*time.Second),
		// 1 operation per second, 100 operations bucket size.
		// This is only for retry speed and its only the overall factor (not per item)
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(1), 100)},
	)
}
