package controllers

import (
	"context"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Modela system represent the model core system
type Loki struct {
	Namespace     string
	Version       string
	ReleaseName   string
	RepoUrl       string
	RepoName      string
	Name          string
	PodNamePrefix string
	Dryrun        bool
}

func NewLoki(version string) *Loki {
	return &Loki{
		Namespace:     "loki",
		Version:       version,
		ReleaseName:   "loki",
		RepoName:      "grafana",
		Name:          "loki-stack",
		PodNamePrefix: "loki",
		RepoUrl:       "https://grafana.github.io/helm-charts",
		Dryrun:        false,
	}
}

func (m Loki) IsEnabled(modela managementv1.Modela) bool {
	return *modela.Spec.Observability.Loki
}

// Check if the database installed
func (m Loki) Installed() (bool, error) {
	return IsChartInstalled(
		m.RepoName,
		m.RepoUrl,
		m.ReleaseName,
		m.Namespace,
		m.ReleaseName,
		m.Version,
	)
}

func (m Loki) Install(ctx context.Context, modela managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := AddRepo(m.RepoName, m.RepoUrl, m.Dryrun); err != nil {
		logger.Error(err, "failed to add repo %s"+m.RepoName)
		return err
	}
	logger.Info("added repo " + m.RepoName)
	// install namespace modela-system
	if err := CreateNamespace(m.Namespace); err != nil {
		logger.Error(err, "failed to create namespace %s"+m.RepoName)
		return err
	}
	logger.Info("created namespace " + m.Namespace)

	return InstallChart(
		m.RepoName,
		m.RepoUrl,
		m.ReleaseName,
		m.Namespace,
		m.ReleaseName,
		m.Version,
	)
}

// Check if we are still installing the database
func (m Loki) Installing() (bool, error) {
	installed, err := m.Installed()
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(m.Namespace, m.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

// Check if the database is ready
func (m Loki) Ready() (bool, error) {
	installed, err := m.Installed()
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(m.Namespace, m.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return running, nil
}

func (m Loki) Uninstall() error {
	return UninstallChart(
		m.RepoName,
		m.RepoUrl,
		"",
		m.Namespace,
		m.ReleaseName,
		m.Version,
	)
}