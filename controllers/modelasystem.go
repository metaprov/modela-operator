package controllers

import (
	"fmt"
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

func (d ModelaSystem) Install() error {
	if err := CreateNamespace("modela-system"); err != nil {
		return err
	}
	fmt.Println("\u2713 created namespace modela-system")

	// apply the crd
	if err := CreateNamespace("modela-catalog"); err != nil {
		return err
	}
	fmt.Println("\u2713 created namespace modela-catalog")

	if err := CreateNamespace("default-tenant"); err != nil {
		return err
	}

	fmt.Println("\u2713 created namespace default-tenant")

	// pull the images.
	fmt.Println("\u2713 pulling modela images")

	dockerClient := RealDockerClient{}
	for _, v := range d.Images {
		fmt.Println("\u2713 pulling image " + v)
		err := dockerClient.Pull(v)
		if err != nil {
			fmt.Println("\u2713 failed to pull image " + v)
			return err
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
