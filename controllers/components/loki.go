package components

import (
	"context"
	"github.com/metaprov/modela-operator/pkg/helm"
	"github.com/metaprov/modela-operator/pkg/kube"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

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
		Name:          "loki",
		PodNamePrefix: "grafana",
		RepoUrl:       "https://grafana.github.io/helm-charts",
		Dryrun:        false,
	}
}

func (m Loki) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingLoki
}

func (m Loki) IsEnabled(modela managementv1.Modela) bool {
	return modela.Spec.Observability.Loki
}

func (m Loki) Installed(ctx context.Context) (bool, error) {
	return helm.IsChartInstalled(ctx, m.Name, m.Namespace, m.ReleaseName)
}

func (m Loki) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := helm.AddRepo(m.RepoName, m.RepoUrl, m.Dryrun); err != nil {
		logger.Error(err, "Failed to download Helm Repo", "repo", m.RepoUrl)
		return err
	}

	logger.Info("Added Helm Repo", "repo", m.RepoName)
	if err := kube.CreateNamespace(m.Namespace, modela.Name); err != nil {
		logger.Error(err, "failed to create namespace")
		return err
	}

	logger.Info("Applying Helm Chart", "version", m.Version)
	return helm.InstallChart(
		ctx,
		m.Name,
		m.Namespace,
		m.ReleaseName,
		map[string]interface{}{},
	)
}

func (m Loki) Installing(ctx context.Context) (bool, error) {
	installed, err := m.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := kube.IsPodRunning(m.Namespace, m.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

func (m Loki) Ready(ctx context.Context) (bool, error) {
	installing, err := m.Installing(ctx)
	if err != nil && err != managementv1.ComponentNotInstalledByModelaError {
		return false, err
	}
	return !installing, nil
}

func (m Loki) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)
	if err := helm.AddRepo(m.RepoName, m.RepoUrl, false); err != nil {
		logger.Error(err, "Failed to download Helm Repo")
		return err
	}

	logger.Info("Added Helm Repo", "repo", m.RepoName)
	return helm.UninstallChart(ctx, m.Name, m.Namespace, m.ReleaseName, map[string]interface{}{})
}
