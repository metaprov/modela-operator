package controllers

import (
	"context"

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
	Dryrun        bool
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
		Dryrun:        false,
	}
}

func (db Database) IsEnabled(modela managementv1.Modela) bool {
	return *modela.Spec.SystemDatabase.Installed

}

// Check if the database installed
func (db Database) Installed() (bool, error) {
	return IsChartInstalled(
		db.RepoName,
		db.RepoUrl,
		db.Url,
		db.Namespace,
		db.ReleaseName,
		db.Version,
	)
}

func (db Database) Install(ctx context.Context, modela managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := AddRepo(db.RepoName, db.RepoUrl, db.Dryrun); err != nil {
		return err
	}
	logger.Info("added repo " + db.RepoName)
	// install namespace modela-system
	if err := CreateNamespace(db.Namespace); err != nil {
		logger.Error(err, "failed to create namespace")
		return err
	}
	logger.Info("created namespace " + db.Namespace)

	return InstallChart(
		db.RepoName,
		db.RepoUrl,
		db.Name,
		db.Namespace,
		db.ReleaseName,
		db.Version,
	)
}

// Check if we are still installing the database
func (d Database) Installing() (bool, error) {
	installed, err := d.Installed()
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
func (db Database) Ready() (bool, error) {
	installed, err := db.Installed()
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(db.Namespace, db.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return running, nil
}

func (db Database) Uninstall() error {
	return UninstallChart(
		db.RepoName,
		db.RepoUrl,
		"",
		db.Namespace,
		db.ReleaseName,
		db.Version,
	)
}
