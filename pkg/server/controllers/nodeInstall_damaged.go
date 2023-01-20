package controllers
/*


i
	packagev1alpha1 "github.com/alaudutias/kubeclskntient"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

type NodeInstallReconciler struct {
	KubeClient *kubeclient.Client
}


func (r *NodeInstallReconciler) ensureOwnerReference(nist *packagev1alpha1.NodeInstall) error {

	node := corev1.Node{}
	if err := r.KubeClient.GetByName("", nist.Name, &node); err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.KubeClient.Delete(nist); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	for _, ref := range nist.OwnerReferences {
		//	if ref.Kind != node.Kind || ref.{

		//	}
	}
	return nil
}

func (r *NodeInstallReconciler) createOrUpdate(nist *packagev1alpha1.NodeInstall) error {
	return nil
}

func (r *NodeInstallReconciler) Reconcile(ctx context.Context, req controllerruntime.Request) (rt controllerruntime.Result, err error) {
	nist := packagev1alpha1.NodeInstall{}
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

	err = r.KubeClient.Get(req.NamespacedName, &nist)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Infof(`Not fournd Package "%s"`, req.Name)
			err = nil
		}
		return
	}

	if err = r.ensureOwnerReference(&nist); err != nil {
		return
	}

	if err = r.createOrUpdate(&nist); err != nil {
		return
	}

	return
}

func (c *NodeInstallReconciler) SetupWithManager(mgr controllerruntime.Manager) error {
	return controllerruntime.NewControllerManagedBy(mgr).
		For(&packagev1alpha1.NodeInstall{}).
		Complete(c)
}
*/
