package controllers

import (
	"context"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Modela system represent the model core system
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
		Url:           "jetstack/cert-manager",
		RepoUrl:       "https://charts.jetstack.io",
		RepoName:      "jetstack",
		Name:          "cert-manager",
		PodNamePrefix: "cert-manager",
	}
}

func (cm CertManager) IsEnabled(modela managementv1.Modela) bool {
	return *modela.Spec.CertManager.Install
}

func (cm CertManager) Installed() (bool, error) {
	return IsPodRunning(cm.Namespace, cm.PodNamePrefix)
}

func (cm CertManager) Install(ctx context.Context, modela managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := AddRepo(cm.RepoName, cm.RepoUrl, false); err != nil {
		logger.Error(err, "Failed to download Helm Repo")
		return err
	}
	logger.Info("Added Helm Repo", "repo", cm.RepoName)
	if err := CreateNamespace(cm.Namespace); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	logger.Info("Applying Helm Chart", "version", cm.Version)
	return InstallChartWithValues(
		ctx,
		cm.RepoName,
		cm.RepoUrl,
		cm.Name,
		cm.Namespace,
		cm.ReleaseName,
		cm.Version,
		map[string]interface{}{
			"installCRDs": "true",
		},
	)

}

func (cm CertManager) Installing() (bool, error) {
	installed, err := cm.Installed()
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(cm.Namespace, cm.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil

}

func (cm CertManager) Ready() (bool, error) {
	installed, err := cm.Installed()
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(cm.Namespace, cm.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return running, nil
}

func (cm CertManager) Uninstall(ctx context.Context) error {
	logger := log.FromContext(ctx)
	if err := AddRepo(cm.RepoName, cm.RepoUrl, false); err != nil {
		logger.Error(err, "Failed to download Helm Repo")
		return err
	}

	logger.Info("Added Helm Repo", "repo", cm.RepoName)
	return UninstallChart(
		ctx,
		cm.RepoName,
		cm.RepoUrl,
		cm.Name,
		cm.Namespace,
		cm.ReleaseName,
		cm.Version,
	)
}
