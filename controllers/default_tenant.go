package controllers

import (
	"context"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
)

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

func (m DefaultTenant) IsEnabled(modela managementv1.Modela) bool {
	return *modela.Spec.DefaultTenantChart.Installed
}

// Check if the database installed
func (dt DefaultTenant) Installed(ctx context.Context) (bool, error) {
	return IsChartInstalled(
		ctx,
		dt.RepoName,
		dt.RepoUrl,
		dt.Url,
		dt.Namespace,
		dt.ReleaseName,
		dt.Version,
	)

}

func (dt DefaultTenant) Install(ctx context.Context, modela managementv1.Modela) error {
	return InstallChart(
		ctx,
		dt.RepoName,
		dt.RepoUrl,
		dt.Name,
		dt.Namespace,
		dt.ReleaseName,
		dt.Version,
	)
}

// Check if we are still installing the default tenant
func (dt DefaultTenant) Installing(ctx context.Context) (bool, error) {
	installed, err := dt.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(dt.Namespace, dt.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

// Check if the default tenant is installed and ready
func (d DefaultTenant) Ready(ctx context.Context) (bool, error) {
	return d.Installed(ctx)
}

func (d DefaultTenant) Uninstall(ctx context.Context) error {
	return nil
}
