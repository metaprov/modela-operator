package controllers

import (
	"context"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Database struct {
	Namespace     string
	Version       string
	Url           string
	ReleaseName   string
	RepoUrl       string
	RepoName      string
	Name          string
	PodNamePrefix string
}

func NewDatabase(version string) *Database {
	return &Database{
		Namespace:     "modela-system",
		Version:       version,
		ReleaseName:   "modela-postgresql",
		RepoUrl:       "https://charts.bitnami.com/bitnami",
		Name:          "postgresql",
		Url:           "postgresql",
		RepoName:      "bitnami",
		PodNamePrefix: "cert-manager",
	}
}

func (db Database) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingDatabase
}

func (db Database) IsEnabled(modela managementv1.Modela) bool {
	return true
}

func (db Database) Installed(ctx context.Context) (bool, error) {
	if installed, err := IsChartInstalled(
		ctx,
		db.RepoName,
		db.RepoUrl,
		db.Url,
		db.Namespace,
		db.ReleaseName,
		db.Version,
	); !installed {
		return false, err
	}
	if belonging, _ := IsStatefulSetCreatedByModela(db.Namespace, "modela-postgresql"); !belonging {
		return true, ComponentNotInstalledByModelaError
	}
	return true, nil
}

func (db Database) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := AddRepo(db.RepoName, db.RepoUrl, false); err != nil {
		logger.Error(err, "Failed to download Helm Repo", "repo", db.RepoUrl)
		return err
	}
	logger.Info("Added Helm Repo", "repo", db.RepoName)
	if err := CreateNamespace(db.Namespace, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	return InstallChart(
		ctx,
		db.RepoName,
		db.RepoUrl,
		db.Name,
		db.Namespace,
		db.ReleaseName,
		db.Version,
	)
}

// Check if we are still installing the database
func (db Database) Installing(ctx context.Context) (bool, error) {
	installed, err := db.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(db.Namespace, db.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

// Check if the database is ready
func (db Database) Ready(ctx context.Context) (bool, error) {
	installing, err := db.Installed(ctx)
	if err != nil && err != ComponentNotInstalledByModelaError {
		return false, err
	}
	return !installing, nil
}

func (db Database) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	return UninstallChart(ctx,
		db.RepoName,
		db.RepoUrl,
		db.Name,
		db.Namespace,
		db.ReleaseName,
		db.Version,
	)
}
