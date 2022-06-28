package controllers

import (
	"context"
	k8serr "k8s.io/apimachinery/pkg/api/errors"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
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
	return *modela.Spec.SystemDatabase.Install

}

// Check if the database installed
func (db Database) Installed(ctx context.Context) (bool, error) {
	return IsChartInstalled(
		ctx,
		db.RepoName,
		db.RepoUrl,
		db.Url,
		db.Namespace,
		db.ReleaseName,
		db.Version,
	)
}

func (db Database) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := AddRepo(db.RepoName, db.RepoUrl, false); err != nil {
		logger.Error(err, "Failed to download Helm Repo")
		return err
	}
	logger.Info("Added Helm Repo", "repo", db.RepoName)
	if err := CreateNamespace(db.Namespace); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "Failed to create namespace")
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
func (d Database) Installing(ctx context.Context) (bool, error) {
	installed, err := d.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(d.Namespace, d.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

// Check if the database is ready
func (db Database) Ready(ctx context.Context) (bool, error) {
	installed, err := db.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(db.Namespace, db.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return running, nil
}

func (db Database) Uninstall(ctx context.Context) error {
	return UninstallChart(ctx,
		db.RepoName,
		db.RepoUrl,
		"",
		db.Namespace,
		db.ReleaseName,
		db.Version,
	)
}
