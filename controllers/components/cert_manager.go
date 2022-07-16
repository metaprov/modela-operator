package components

import (
	"context"
	"fmt"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/helm"
	"github.com/metaprov/modela-operator/pkg/kube"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type CertManager struct {
	Namespace     string
	Version       string
	ReleaseName   string
	Url           string
	RepoUrl       string
	RepoName      string
	Name          string
	PodNamePrefix string
}

func NewCertManager(version string) *CertManager {
	return &CertManager{
		Namespace:     "cert-manager",
		Version:       version,
		ReleaseName:   "cert-manager",
		Url:           "cert-manager",
		RepoName:      "jetstack",
		Name:          "cert-manager",
		PodNamePrefix: "cert-manager",
	}
}

func (cm CertManager) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingCertManager
}

func (cm CertManager) IsEnabled(modela managementv1.Modela) bool {
	return modela.Spec.CertManager.Install
}

func (cm CertManager) Installed(ctx context.Context) (bool, error) {
	fmt.Println("Checking deployment")
	if belonging, err := kube.IsDeploymentCreatedByModela(cm.Namespace, "cert-manager"); err == nil && !belonging {
		return true, managementv1.ComponentNotInstalledByModelaError
	}
	fmt.Println("Checking chart install")
	if installed, err := helm.IsChartInstalled(ctx, cm.Name, cm.Namespace, cm.ReleaseName); !installed {
		return false, err
	}
	return true, nil
}

func (cm CertManager) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := helm.AddRepo(cm.RepoName, cm.RepoUrl, false); err != nil {
		logger.Error(err, "Failed to download Helm Repo", "repo", cm.RepoUrl)
		return err
	}
	logger.Info("Added Helm Repo", "repo", cm.RepoName)
	if err := kube.CreateNamespace(cm.Namespace, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	logger.Info("Applying Helm Chart", "version", cm.Version)
	return helm.InstallChart(ctx, cm.Name, cm.Namespace, cm.ReleaseName, map[string]interface{}{
		"installCRDs": "true",
	})

}

func (cm CertManager) Installing(ctx context.Context) (bool, error) {
	installed, err := cm.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := kube.IsPodRunning(cm.Namespace, cm.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil

}

func (cm CertManager) Ready(ctx context.Context) (bool, error) {
	installing, err := cm.Installing(ctx)
	if err != nil && err != managementv1.ComponentNotInstalledByModelaError {
		return false, err
	}
	return !installing, nil
}

func (cm CertManager) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)
	if err := helm.AddRepo(cm.RepoName, cm.RepoUrl, false); err != nil {
		logger.Error(err, "Failed to download Helm Repo")
		return err
	}

	logger.Info("Added Helm Repo", "repo", cm.RepoName)
	return helm.UninstallChart(ctx, cm.Name, cm.Namespace, cm.ReleaseName, map[string]interface{}{})
}
