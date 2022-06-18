package controllers

import (
	"context"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Modela system represent the model core system
type ModelaSystem struct {
	Namespace     string
	Version       string
	Url           string
	ReleaseName   string
	RepoUrl       string
	RepoName      string
	Name          string
	PodNamePrefix string
	Images        []string
	Dryrun        bool
}

func (m ModelaSystem) IsEnabled(modela managementv1.Modela) bool {
	return *modela.Spec.DefaultTenantChart.Installed
}

func NewModelaSystem(version string) *ModelaSystem {
	return &ModelaSystem{
		Namespace:   "modela-system",
		Version:     version,
		ReleaseName: "modela",
		RepoName:    "modela-charts",
		Name:        "modela",
		RepoUrl:     "https://metaprov.github.io/helm-charts",
		Dryrun:      false,
		Images: []string{
			"ghcr.io/metaprov/modela-supervisor:" + version,
			"ghcr.io/metaprov/modela-frontend:" + version,
			"ghcr.io/metaprov/modela-prediction-router:" + version,
			"ghcr.io/metaprov/modela-database-proxy:" + version,
			"ghcr.io/metaprov/modela-batch-predictor:" + version,
			"ghcr.io/metaprov/modela-trainer:" + version,
			"ghcr.io/metaprov/modela-publisher:" + version,
			"ghcr.io/metaprov/modela-workload:" + version,
			"ghcr.io/metaprov/modela-data-dock:" + version,
			"ghcr.io/metaprov/modela-data-plane:" + version,
			"ghcr.io/metaprov/modela-control-plane:" + version,
			"ghcr.io/metaprov/modela-cloud-proxy:" + version,
			"ghcr.io/metaprov/modela-api-gateway:" + version,
		},
	}
}

// Check if the database installed
func (ms ModelaSystem) Installed() (bool, error) {
	return IsChartInstalled(
		ms.RepoName,
		ms.RepoUrl,
		ms.Url,
		ms.Namespace,
		ms.ReleaseName,
		ms.Version,
	)
}

func (d ModelaSystem) Install(ctx context.Context, modela managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := CreateNamespace("modela-system"); err != nil {
		logger.Error(err, "failed to create modela-system namespace")
		return err
	}
	logger.Info("created namespace modela-system")

	// apply the crd
	if err := CreateNamespace("modela-catalog"); err != nil {
		logger.Error(err, "failed to create modela-catalog namespace")
		return err
	}
	logger.Info("created namespace modela-catalog")

	if err := CreateNamespace("default-tenant"); err != nil {
		logger.Error(err, "failed to create default-tenant namespace")
		return err
	}

	logger.Info("created namespace default-tenant")

	// pull the images.
	logger.Info("pulling modela images")

	dockerClient := RealDockerClient{}
	for _, v := range d.Images {
		logger.Info("pulling image " + v)
		err := dockerClient.Pull(v)
		if err != nil {
			logger.Error(err, "failed to pull image "+v)
		}
	}

	return InstallChart(
		d.RepoName,
		d.RepoUrl,
		d.Name,
		d.Namespace,
		d.ReleaseName,
		d.Version,
	)
}

// Check if we are still installing the database
func (ms ModelaSystem) Installing() (bool, error) {
	installed, err := ms.Installed()
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(ms.Namespace, ms.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

// Check if the database is ready
func (ds ModelaSystem) Ready() (bool, error) {
	installed, err := ds.Installed()
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(ds.Namespace, ds.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return running, nil
}

func (dm ModelaSystem) Uninstall() error {
	return UninstallChart(
		dm.RepoName,
		dm.RepoUrl,
		"",
		dm.Namespace,
		dm.ReleaseName,
		dm.Version,
	)
}
