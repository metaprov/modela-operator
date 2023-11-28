package components

import (
	"context"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/helm"
	"github.com/metaprov/modela-operator/pkg/kube"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Nginx struct {
	Namespace     string
	Name          string
	ReleaseName   string
	PodNamePrefix string
	Dryrun        bool
}

func NewNginx() *Nginx {
	return &Nginx{
		Namespace:     "nginx",
		ReleaseName:   "ingress-nginx",
		Name:          "ingress-nginx",
		PodNamePrefix: "ingress-nginx-controller",
		Dryrun:        false,
	}
}

func (n Nginx) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingNginx
}

func (n Nginx) IsEnabled(modela managementv1.Modela) bool {
	if modela.Spec.Network.Nginx == nil {
		return false
	}
	return modela.Spec.Network.Nginx.Install
}

func (n Nginx) Installed(ctx context.Context) (bool, error) {
	if belonging, err := kube.IsDeploymentCreatedByModela(n.Namespace, "ingress-nginx-controller"); err == nil && !belonging {
		return true, managementv1.ComponentNotInstalledByModelaError
	}
	if installed, err := helm.IsChartInstalled(ctx, n.Name, n.Namespace, n.ReleaseName); !installed {
		return false, err
	}

	return true, nil
}

func (n Nginx) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := kube.CreateNamespace(n.Namespace, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	return helm.InstallChart(ctx, n.Name, n.Namespace, n.ReleaseName, modela.Spec.Network.Nginx.Values.Object)
}

// Check if we are still installing the database
func (n Nginx) Installing(ctx context.Context) (bool, error) {
	installed, err := n.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := kube.IsPodRunning(n.Namespace, n.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

func (n Nginx) Ready(ctx context.Context) (bool, error) {
	installing, err := n.Installing(ctx)
	if err != nil && err != managementv1.ComponentNotInstalledByModelaError {
		return false, err
	}
	return !installing, nil
}

func (n Nginx) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	return helm.UninstallChart(ctx, n.Name, n.Namespace, n.ReleaseName, map[string]interface{}{})
}
