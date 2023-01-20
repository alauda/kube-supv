package controllers

import (
	"context"

	packagev1alpha1 "github.com/alauda/kube-supv/api/package/v1alpha1"
	"github.com/alauda/kube-supv/pkg/log"
	"github.com/alauda/kube-supv/pkg/utils/kubeclient"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

type InstallRequestReconciler struct {
	KubeClient *kubeclient.Client
}

func (r *InstallRequestReconciler) createOrUpdate(ireq *packagev1alpha1.InstallRequest) error {
	return nil
}

func (r *InstallRequestReconciler) delete(ireq *packagev1alpha1.InstallRequest) error {
	return nil
}

func (r *InstallRequestReconciler) Reconcile(ctx context.Context, req controllerruntime.Request) (rt controllerruntime.Result, err error) {
	ireq := packagev1alpha1.InstallRequest{}
	defer func() {
		if any := recover(); any != nil {
			log.DPanicf(`Package Reconcile "%s" recover from panic: %v`, req.Name, any)
		}
		if err != nil {
			log.Errorf(`Package Reconcile "%s" error: %v`, req.Name, err)
			err = nil
			rt.RequeueAfter = reconcileAfterDuration
		} else {
			rt.RequeueAfter = reconcileHealthCheckDuration
		}
	}()

	err = r.KubeClient.Get(req.NamespacedName, &ireq)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Infof(`Not fournd Package "%s"`, req.Name)
			err = nil
		}
		return
	}

	if ireq.DeletionTimestamp != nil {
		err = r.delete(&ireq)
		return
	}

	if err = r.KubeClient.EnsureFinalizer(&ireq, finalizer); err != nil {
		return
	}

	if err = r.createOrUpdate(&ireq); err != nil {
		return
	}

	return
}

func (c *InstallRequestReconciler) SetupWithManager(mgr controllerruntime.Manager) error {
	return controllerruntime.NewControllerManagedBy(mgr).
		For(&packagev1alpha1.InstallRequest{}).
		WithEventFilter(kubeclient.EventFilter).
		Complete(c)
}
