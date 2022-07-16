package components

import (
	"context"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/helm"
	"github.com/metaprov/modela-operator/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Grafana struct {
	Namespace     string
	Version       string
	ReleaseName   string
	RepoUrl       string
	RepoName      string
	Name          string
	PodNamePrefix string
	Dryrun        bool
}

func NewGrafana(version string) *Grafana {
	return &Grafana{
		Namespace:     "grafana",
		Version:       version,
		ReleaseName:   "grafana-stack",
		RepoName:      "grafana",
		Name:          "grafana",
		PodNamePrefix: "grafana-stack",
		RepoUrl:       "https://grafana.github.io/helm-charts",
		Dryrun:        false,
	}
}

func (m Grafana) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingGrafana
}

func (m Grafana) IsEnabled(modela managementv1.Modela) bool {
	return modela.Spec.Observability.Grafana
}

func (m Grafana) Installed(ctx context.Context) (bool, error) {
	if belonging, err := kube.IsDeploymentCreatedByModela(m.Namespace, "grafana-stack"); err == nil && !belonging {
		return true, managementv1.ComponentNotInstalledByModelaError
	}
	return helm.IsChartInstalled(ctx, m.Name, m.Namespace, m.ReleaseName)
}

func (m Grafana) Install(ctx context.Context, modela *managementv1.Modela) error {
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
	return helm.InstallChart(ctx, m.Name, m.Namespace, m.ReleaseName, map[string]interface{}{})
}

func (m Grafana) Installing(ctx context.Context) (bool, error) {
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

func (m Grafana) Ready(ctx context.Context) (bool, error) {
	installing, err := m.Installing(ctx)
	if err != nil && err != managementv1.ComponentNotInstalledByModelaError {
		return false, err
	}
	return !installing, nil
}

func (m Grafana) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)
	if err := helm.AddRepo(m.RepoName, m.RepoUrl, false); err != nil {
		logger.Error(err, "Failed to download Helm Repo")
		return err
	}

	logger.Info("Added Helm Repo", "repo", m.RepoName)
	return helm.UninstallChart(ctx, m.Name, m.Namespace, m.ReleaseName, map[string]interface{}{})
}
