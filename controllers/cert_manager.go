package controllers

import (
	"fmt"
	"github.com/metaprov/modela-operator/internal/pkg/util"
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

func NewCertManager() *CertManager {
	return &CertManager{
		Namespace:     "cert-manager",
		Version:       "v1.6.1",
		ReleaseName:   "cert-manager",
		Url:           "jetstack/cert-manager",
		RepoUrl:       "https://charts.jetstack.io",
		RepoName:      "jetstack",
		Name:          "cert-manager",
		PodNamePrefix: "cert-manager",
	}
}

func (cm CertManager) Installed() (bool, error) {
	return util.IsPodRunning(cm.Namespace, cm.PodNamePrefix)
}

func (cm CertManager) Install() error {
	if err := util.AddRepo(cm.RepoName, cm.RepoUrl, false); err != nil {
		return err
	}
	fmt.Println("\u2713 added repo " + cm.RepoName)
	// install namespace modela-system
	if err := util.CreateNamespace(cm.Namespace); err != nil {
		return err
	}
	fmt.Println("\u2713 created namespace " + cm.Namespace)
	return util.InstallCrd("https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml")

}

func (cm CertManager) Installing() (bool, error) {
	installed, err := cm.Installed()
	if !installed {
		return installed, err
	}
	running, err := util.IsPodRunning(cm.Namespace, cm.PodNamePrefix)
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
	running, err := util.IsPodRunning(cm.Namespace, cm.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return running, nil
}

func (cm CertManager) Uninstall() error {
	return util.UninstallChart(
		cm.RepoName,
		cm.RepoUrl,
		"",
		cm.Namespace,
		cm.ReleaseName,
		cm.Version,
	)
}
