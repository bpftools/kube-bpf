package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/bpftools/kube-bpf/apis/v1alpha1"
	"github.com/bpftools/kube-bpf/daemonset"
	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	tappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	tcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type BPF struct {
	informer     cache.SharedInformer
	workqueue    workqueue.RateLimitingInterface
	corev1Client tcorev1.CoreV1Interface
	appsv1Client tappsv1.AppsV1Interface
	logger       *zap.Logger
	validator    *validator.Validate
}

func NewBPF(
	i cache.SharedInformer,
	wq workqueue.RateLimitingInterface,
	kc *kubernetes.Clientset) *BPF {
	return &BPF{
		informer:     i,
		workqueue:    wq,
		corev1Client: kc.CoreV1(),
		appsv1Client: kc.AppsV1(),
		logger:       zap.NewNop(),
		validator:    validator.New(),
	}
}

func (o *BPF) WithLogger(logger *zap.Logger) {
	o.logger = logger
}

func (s *BPF) Run(threadiness int, stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer utilruntime.HandleCrash()
	defer s.workqueue.ShutDown()
	s.logger.Info("starting BPF controller")

	go s.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, s.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(s.runWorker, time.Second, stopCh)
	}

	<-stopCh
	s.logger.Info("stopping BPF controller")
}

func (s *BPF) runWorker() {
	for s.processNextItem() {
	}
}

func (s *BPF) processNextItem() bool {
	keysnap, quit := s.workqueue.Get()
	if quit {
		return false
	}
	defer s.workqueue.Done(keysnap)

	err := s.syncToStdout(keysnap)

	s.handleErr(err, keysnap)
	return true
}

func (s *BPF) syncToStdout(keysnap interface{}) error {
	k, err := cache.MetaNamespaceKeyFunc(keysnap)
	if err != nil {
		s.logger.Error("extracting key from BPF failed", zap.Error(err))
		return err
	}
	obj, exists, err := s.informer.GetStore().GetByKey(k)
	if err != nil {
		s.logger.Error("fetching BPF from queue failed", zap.String("key", k), zap.Error(err))
		return err
	}

	if exists {
		bp := obj.(*v1alpha1.BPF)
		ds, err := daemonset.New(bp, s.appsv1Client)
		if err != nil {
			return err
		}
		_, err = ds.Create()
		if err != nil {
			if errors.IsAlreadyExists(err) {
				return nil
			}
			return err
		}

		svc, err := daemonset.NewService(bp, s.corev1Client)
		if err != nil {
			return err
		}
		_, err = svc.Create()
		if err != nil {
			if errors.IsAlreadyExists(err) {
				return nil
			}
			return err
		}

		return nil
	}

	//bp := keysnap.(*v1alpha1.BPF)

	//s.logger.Info("BPF resource does not exist, removing stuff", zap.String("bpf", bp.GetName()))
	//if err := daemonset.Delete(bp, s.appsv1Client); err != nil {
	//s.logger.Error("daemonset removal failed", zap.Error(err), zap.String("bpf", bp.GetName()))
	//}

	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (s *BPF) handleErr(err error, keysnap interface{}) {
	snap := keysnap.(*v1alpha1.BPF)
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		s.workqueue.Forget(keysnap)
		return
	}

	s.reportErrored(snap, err.Error())
	k, errmeta := cache.MetaNamespaceKeyFunc(keysnap)
	if errmeta != nil {
		s.logger.Error("Extracting key from BPFs failed while handling error", zap.Error(errmeta))
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if s.workqueue.NumRequeues(keysnap) < 5 {
		s.logger.Info("Error syncing BPF", zap.String("key", k), zap.Error(err))
		s.reportNewRetry(snap)
		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		s.workqueue.AddRateLimited(keysnap)
		return
	}

	s.workqueue.Forget(keysnap)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	utilruntime.HandleError(err)
	s.logger.Warn("Dropping BPF", zap.String("key", k), zap.Error(err))
	s.reportDropped(snap)
}

func (s *BPF) reportDropped(sn *v1alpha1.BPF) {
	_, err := s.corev1Client.Events(sn.Namespace).Create(newBPFEvent(sn, "Dropped", "BPF dropped", "BPF dropped after several retries, removal of the BPF resource recommended"))
	if err != nil {
		s.logger.Error("error creating event for BPF dropped", zap.Error(err))
	}
}

func (s *BPF) reportNewRetry(sn *v1alpha1.BPF) {
	_, err := s.corev1Client.Events(sn.Namespace).Create(newBPFEvent(sn, "Retry", "Retry planned", "BPF errored but will be retried"))
	if err != nil {
		s.logger.Error("error creating event for BPF to be retried", zap.Error(err))
	}
}

func (s *BPF) reportInvalid(sn *v1alpha1.BPF, message string) {
	_, err := s.corev1Client.Events(sn.Namespace).Create(newBPFEvent(sn, "Invalid", "BPF not valid", message))
	if err != nil {
		s.logger.Error("error creating event for BPF invalid", zap.Error(err))
	}
}

func (s *BPF) reportErrored(sn *v1alpha1.BPF, message string) {
	_, err := s.corev1Client.Events(sn.Namespace).Create(newBPFEvent(sn, "Error", "BPF errored", message))
	if err != nil {
		s.logger.Error("error creating event for BPF errored", zap.Error(err))
	}
}

func newBPFEvent(snap *v1alpha1.BPF, eventType string, reason string, message string) *v1.Event {
	t := time.Now()
	return &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: snap.Name + "-",
			Namespace:    snap.Namespace,
		},
		InvolvedObject: v1.ObjectReference{
			APIVersion:      snap.APIVersion,
			Kind:            snap.Kind,
			Name:            snap.Name,
			Namespace:       snap.Namespace,
			UID:             snap.UID,
			ResourceVersion: snap.ResourceVersion,
		},
		// Treat each event as unique
		FirstTimestamp: metav1.Time{Time: t},
		LastTimestamp:  metav1.Time{Time: t},
		Count:          int32(1),
		Reason:         reason,
		Message:        message,
		Type:           eventType,
	}
}
