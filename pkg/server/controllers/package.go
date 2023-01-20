package controllers

import (
	"context"

	packagev1alpha1 "github.com/alauda/kube-supv/api/package/v1alpha1"
	"github.com/alauda/kube-supv/pkg/log"
	"github.com/alauda/kube-supv/pkg/utils/kubeclient"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

type PackageReconciler struct {
	KubeClient *kubeclient.Client
}

func (r *PackageReconciler) Reconcile(ctx context.Context, req controllerruntime.Request) (rt controllerruntime.Result, err error) {
	pakg := packagev1alpha1.Package{}
	defer func() {
		if any := recover(); any != nil {
			log.DPanicf(`Package reconcile "%s" recover from panic: %v`, req.Name, any)
		}
		if err != nil {
			log.Errorf(`Package reconcile "%s" error: %v`, req.Name, err)
			err = nil
			rt.RequeueAfter = reconcileAfterDuration
		} else {
			rt.RequeueAfter = reconcileHealthCheckDuration
		}
	}()

	err = r.KubeClient.Get(req.NamespacedName, &pakg)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Infof(`Not fournd Package "%s"`, req.Name)
			err = nil
		}
		return
	}

	if err = r.KubeClient.EnsureFinalizer(&pakg, finalizer); err != nil {
		return
	}

	return
}

func (c *PackageReconciler) SetupWithManager(mgr controllerruntime.Manager) error {
	return controllerruntime.NewControllerManagedBy(mgr).
		For(&packagev1alpha1.Package{}).
		WithEventFilter(kubeclient.EventFilter).
		Complete(c)
}
