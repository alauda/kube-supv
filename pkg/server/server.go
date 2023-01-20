package server

import (
	"time"

	packagev1alpha1 "github.com/alauda/kube-supv/api/package/v1alpha1"
	"github.com/alauda/kube-supv/pkg/log"
	"github.com/alauda/kube-supv/pkg/server/controllers"
	"github.com/alauda/kube-supv/pkg/utils/kubeclient"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

const (
	leaderElectionID = "kubesupv-lock"
)

var (
	scheme        = runtime.NewScheme()
	renewDeadline = time.Second * 40
	leaseDuration = time.Second * 60
	retryPeriod   = time.Second * 12
)

func init() {
	for _, addScheme := range []func(*runtime.Scheme) error{
		corev1.AddToScheme,
		packagev1alpha1.AddToScheme,
	} {
		if err := addScheme(scheme); err != nil {
			panic(err)
		}
	}

}

func Start() error {
	controllerruntime.SetLogger(zapr.NewLogger(log.Logger()))
	manager, err := controllerruntime.NewManager(controllerruntime.GetConfigOrDie(), controllerruntime.Options{
		Scheme:                     scheme,
		LeaderElection:             true,
		LeaderElectionID:           leaderElectionID,
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
		LeaderElectionNamespace:    metav1.NamespaceSystem,
		RenewDeadline:              &renewDeadline,
		LeaseDuration:              &leaseDuration,
		RetryPeriod:                &retryPeriod,
	})
	if err != nil {
		return errors.Wrap(err, "create new manager")
	}

	if err = (&controllers.PackageReconciler{
		KubeClient: kubeclient.Wrap(manager.GetClient()),
	}).SetupWithManager(manager); err != nil {
		return errors.Wrap(err, "create Package controller")
	}

	if err = (&controllers.NodeInstallReconciler{
		KubeClient: kubeclient.Wrap(manager.GetClient()),
	}).SetupWithManager(manager); err != nil {
		return errors.Wrap(err, "create NodeInstall controller")
	}

	if err = (&controllers.InstallRequestReconciler{
		KubeClient: kubeclient.Wrap(manager.GetClient()),
	}).SetupWithManager(manager); err != nil {
		return errors.Wrap(err, "create InstallRequest controller")
	}

	log.Infof("starting manager")
	if err := manager.Start(controllerruntime.SetupSignalHandler()); err != nil {
		return errors.Wrapf(err, "server stopped")
	}
	return nil
}
