package controllers

import "github.com/metaprov/modela-operator/internal/pkg/util"

type DefaultTenant struct {
	Namespace     string
	Version       string
	Url           string
	ReleaseName   string
	RepoUrl       string
	RepoName      string
	Name          string
	PodNamePrefix string
	Dryrun        bool
}

func NewDefaultTenant(version string) *DefaultTenant {
	return &DefaultTenant{
		Namespace:   "modela-system",
		Version:     version,
		ReleaseName: "modela-default-tenant",
		RepoName:    "modela-charts",
		Name:        "modela-default-tenant",
		RepoUrl:     "https://metaprov.github.io/helm-charts/",
		Dryrun:      false,
	}
}

// Check if the database installed
func (dt DefaultTenant) Installed() (bool, error) {
	return util.IsChartInstalled(
		dt.RepoName,
		dt.RepoUrl,
		dt.Url,
		dt.Namespace,
		dt.ReleaseName,
		dt.Version,
	)

}

func (dt DefaultTenant) Install() error {
	return util.InstallChart(
		dt.RepoName,
		dt.RepoUrl,
		dt.Name,
		dt.Namespace,
		dt.ReleaseName,
		dt.Version,
	)
}

// Check if we are still installing the default tenant
func (dt DefaultTenant) Installing() (bool, error) {
	installed, err := dt.Installed()
	if !installed {
		return installed, err
	}
	running, err := util.IsPodRunning(dt.Namespace, dt.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

// Check if the default tenant is installed and ready
func (d DefaultTenant) Ready() (bool, error) {
	return false, nil
}

func (d DefaultTenant) Uninstall() error {
	return nil
}
