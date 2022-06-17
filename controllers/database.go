package controllers

import (
	"fmt"
	"github.com/metaprov/modela-operator/internal/pkg/util"
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

func NewDatabase() *Database {
	return &Database{
		Namespace:     "modela-system",
		Version:       "10.9.2",
		ReleaseName:   "modela-postgresql",
		RepoUrl:       "https://charts.bitnami.com/bitnami",
		Name:          "postgresql",
		Url:           "postgresql",
		RepoName:      "bitnami",
		PodNamePrefix: "cert-manager",
		Dryrun:        false,
	}
}

// Check if the database installed
func (db Database) Installed() (bool, error) {
	return util.IsChartInstalled(
		db.RepoName,
		db.RepoUrl,
		db.Url,
		db.Namespace,
		db.ReleaseName,
		db.Version,
	)
}

func (db Database) Install() error {

	if err := util.AddRepo(db.RepoName, db.RepoUrl, db.Dryrun); err != nil {
		return err
	}
	fmt.Println("\u2713 added repo " + db.RepoName)
	// install namespace modela-system
	if err := util.CreateNamespace(db.Namespace); err != nil {
		return err
	}
	fmt.Println("\u2713 created namespace " + db.Namespace)

	return util.InstallChart(
		db.RepoName,
		db.RepoUrl,
		"",
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
	running, err := util.IsPodRunning(d.Namespace, d.PodNamePrefix)
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
	running, err := util.IsPodRunning(db.Namespace, db.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return running, nil
}

func (db Database) Uninstall() error {
	return util.UninstallChart(
		db.RepoName,
		db.RepoUrl,
		"",
		db.Namespace,
		db.ReleaseName,
		db.Version,
	)
}
