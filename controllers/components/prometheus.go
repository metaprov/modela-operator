package components

import (
	"context"
	"github.com/metaprov/modela-operator/pkg/helm"
	"github.com/metaprov/modela-operator/pkg/kube"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Modela system represent the model core system
type Prometheus struct {
	Namespace     string
	Version       string
	ReleaseName   string
	RepoUrl       string
	RepoName      string
	Url           string
	Name          string
	PodNamePrefix string
	Dryrun        bool
}

func NewPrometheus() *Prometheus {
	return &Prometheus{
		Namespace:     "prometheus-community",
		ReleaseName:   "kube-prometheus",
		RepoName:      "prometheus-community",
		Name:          "prometheus",
		Url:           "prometheus",
		PodNamePrefix: "kube-prometheus-server",
		RepoUrl:       "https://prometheus-community.github.io/helm-charts",
		Dryrun:        false,
	}
}

func (m Prometheus) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingPrometheus
}

func (m Prometheus) IsEnabled(modela managementv1.Modela) bool {
	return modela.Spec.Observability.Prometheus
}

func (m Prometheus) Installed(ctx context.Context) (bool, error) {
	if belonging, err := kube.IsDeploymentCreatedByModela(m.Namespace, "kube-prometheus-server"); err == nil && !belonging {
		return true, managementv1.ComponentNotInstalledByModelaError
	}
	return helm.IsChartInstalled(ctx, m.Name, m.Namespace, m.ReleaseName)
}

func (m Prometheus) Install(ctx context.Context, modela *managementv1.Modela) error {
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
	return helm.InstallChart(ctx, m.Name, m.Namespace, m.ReleaseName, modela.Spec.Observability.PrometheusValues.Object)
}

func (m Prometheus) Installing(ctx context.Context) (bool, error) {
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

func (m Prometheus) Ready(ctx context.Context) (bool, error) {
	installing, err := m.Installing(ctx)
	if err != nil && err != managementv1.ComponentNotInstalledByModelaError {
		return false, err
	}
	return !installing, nil
}

func (m Prometheus) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	return helm.UninstallChart(
		ctx,
		m.Name,
		m.Namespace,
		m.ReleaseName,
		map[string]interface{}{},
	)
}
