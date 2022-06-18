package controllers

import (
	"context"
	"fmt"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
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
	return *modela.Spec.CertManager.Installed
}

func (cm CertManager) Installed() (bool, error) {
	return IsPodRunning(cm.Namespace, cm.PodNamePrefix)
}

func (cm CertManager) Install(ctx context.Context, modela managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := AddRepo(cm.RepoName, cm.RepoUrl, false); err != nil {
		logger.Error(err, "failed to add repo "+cm.RepoUrl)
		return err
	}
	logger.Info("added repo " + cm.RepoName)
	// install namespace modela-system
	if err := CreateNamespace(cm.Namespace); err != nil {
		logger.Error(err, "failed to create namespace "+cm.Namespace)
		return err
	}
	logger.Info("created namespace " + cm.Namespace)
	versionUrl := fmt.Sprintf("https://github.com/cert-manager/cert-manager/releases/download/%s/cert-manager.yaml", *modela.Spec.CertManager.ChartVersion)
	return InstallCrd(versionUrl)

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

func (cm CertManager) Uninstall() error {
	return UninstallChart(
		cm.RepoName,
		cm.RepoUrl,
		"",
		cm.Namespace,
		cm.ReleaseName,
		cm.Version,
	)
}
