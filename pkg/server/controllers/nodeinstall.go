package controllers

import (
	"context"
	"fmt"
	"time"

	packagev1alpha1 "github.com/alauda/kube-supv/api/package/v1alpha1"
	"github.com/alauda/kube-supv/pkg/log"
	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/alauda/kube-supv/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type NodeInstallReconciler struct {
	KubeClient *kubeclient.Client
	Image      string
}

func (r *NodeInstallReconciler) update(nist *packagev1alpha1.NodeInstall, node *corev1.Node) error {
	if err := r.KubeClient.EnsureFinalizer(nist, finalizer); err != nil {
		return err
	}
	if err := r.KubeClient.AddOwnerReference(nist, node); err != nil {
		return err
	}

	if err := r.recycleChecker(nist); err != nil {
		return errors.Wrapf(err, `recycle checker for NodeInstall "%s"`, nist.Name)
	}

	if nist.Status.NextCheckTime.After(time.Now()) {
		if err := r.createChecker(nist.Name); err != nil {
			return errors.Wrapf(err, `create NodeInstall checker for "%s"`)
		}
	}
	return nil
}

func (r *NodeInstallReconciler) cleanChecker(name string) error {
	podName := fmt.Sprintf(checkerNameTemplate, name)
	pod := corev1.Pod{}
	if err := r.KubeClient.GetByName(metav1.NamespaceSystem, podName, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := r.KubeClient.Delete(&pod); err != nil {
		return err
	}
	return nil
}

func (r *NodeInstallReconciler) recycleChecker(nist *packagev1alpha1.NodeInstall) error {
	podName := fmt.Sprintf(checkerNameTemplate, nist.Name)
	pod := corev1.Pod{}
	if err := r.KubeClient.GetByName(metav1.NamespaceSystem, podName, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	message := ""
	phase := nist.Status.Phase
	switch pod.Status.Phase {
	case corev1.PodFailed:
		if len(pod.Status.ContainerStatuses) > 0 {
			if terminated := pod.Status.ContainerStatuses[0].LastTerminationState.Terminated; terminated != nil {
				message = fmt.Sprintf("%s: %s", terminated.Reason, terminated.Message)
			}
		}
		phase = packagev1alpha1.NodeInstallFailed

	case corev1.PodSucceeded:
		phase = packagev1alpha1.NodeInstallUpdated
		nist.Status.Message = message
	case corev1.PodRunning:
		phase = packagev1alpha1.NodeInstallUpdating
	case corev1.PodPending:
		phase = packagev1alpha1.NodeInstallUpdating
	case corev1.PodUnknown:
		phase = packagev1alpha1.NodeInstallUnknown
	}

	changed := false

	if nist.Status.Phase != phase {
		nist.Status.Phase = phase
		changed = true
	}
	if nist.Status.Message != message {
		nist.Status.Message = message
		changed = true
	}
	if changed {
		if err := r.KubeClient.UpdateStatus(nist); err != nil {
			return err
		}
	}
	if phase != packagev1alpha1.NodeInstallUpdating {
		if err := r.KubeClient.Delete(&pod); err != nil {
			return err
		}
	}
	return nil
}

func (r *NodeInstallReconciler) createChecker(name string) error {
	podName := fmt.Sprintf(checkerNameTemplate, name)
	pod := corev1.Pod{}

	err := r.KubeClient.GetByName(metav1.NamespaceSystem, podName, &pod)
	if err != nil && (!apierrors.IsNotFound(err)) {
		return err
	}

	if err == nil {
		if err := r.KubeClient.Delete(&pod); err != nil {
			return err
		}
	}

	if err := utils.RenderObject(checkerPodTemplate, struct {
		NodeName string
		Image    string
	}{
		NodeName: name,
		Image:    r.Image,
	}, &pod); err != nil {
		return errors.Wrapf(err, `render checker for NodeInstall "%s"`, name)
	}

	if err := r.KubeClient.Create(&pod); err != nil {
		return errors.Wrapf(err, `create checker for NodeInstall "%s"`, name)
	}

	return nil
}

func (r *NodeInstallReconciler) Reconcile(ctx context.Context, req controllerruntime.Request) (rt controllerruntime.Result, err error) {

	defer func() {
		if any := recover(); any != nil {
			log.DPanicf(`NodeInstall reconcile "%s" recover from panic: %v`, req.Name, any)
		}
		if err != nil {
			log.Errorf(`NodeInstall reconcile "%s" error: %v`, req.Name, err)
			err = nil
			rt.RequeueAfter = reconcileAfterDuration
		} else {
			rt.RequeueAfter = reconcileHealthCheckDuration
		}
	}()

	node := corev1.Node{}
	nodeExist := true
	if err = r.KubeClient.GetByName("", req.Name, &node); err != nil {
		if apierrors.IsNotFound(err) {
			nodeExist = false
			err = nil
		} else {
			return
		}
	}

	nist := packagev1alpha1.NodeInstall{}
	nistExist := true
	if err = r.KubeClient.Get(req.NamespacedName, &nist); err != nil {
		if apierrors.IsNotFound(err) {
			nistExist = false
			err = nil
		} else {
			return
		}
	}

	if !nodeExist {
		if nistExist {
			err = r.KubeClient.Delete(&nist)
		}
		return
	}

	if !nistExist {
		err = r.createChecker(req.Name)
		return
	}

	if nist.DeletionTimestamp != nil {
		if err = r.cleanChecker(req.Name); err != nil {
			return
		}
		if err = r.KubeClient.FinishFinalizer(&nist, finalizer); err != nil {
			return
		}
		return
	}

	if err = r.update(&nist, &node); err != nil {
		return
	}
	return
}

func (c *NodeInstallReconciler) SetupWithManager(mgr controllerruntime.Manager) error {
	return controllerruntime.NewControllerManagedBy(mgr).
		For(&packagev1alpha1.NodeInstall{}).
		WithEventFilter(kubeclient.EventFilter).
		Watches(&source.Kind{Type: &corev1.Node{}}, &handler.Funcs{
			CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
				q.Add(controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name: e.Object.GetName(),
					},
				})
			},
			UpdateFunc: func(ue event.UpdateEvent, q workqueue.RateLimitingInterface) {
				q.Add(controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name: ue.ObjectNew.GetName(),
					},
				})
			},
			DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
				q.Add(controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name: e.Object.GetName(),
					},
				})
			},
		}).
		Complete(c)
}
