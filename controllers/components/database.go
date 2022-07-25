package components

import (
	"context"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/helm"
	"github.com/metaprov/modela-operator/pkg/kube"
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

func NewDatabase() *Database {
	return &Database{
		Namespace:     "modela-system",
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
	if belonging, err := kube.IsStatefulSetCreatedByModela(db.Namespace, "modela-postgresql"); err == nil && !belonging {
		return true, managementv1.ComponentNotInstalledByModelaError
	}
	if installed, err := helm.IsChartInstalled(ctx, db.Name, db.Namespace, db.ReleaseName); !installed {
		return false, err
	}
	return true, nil
}

func (db Database) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := helm.AddRepo(db.RepoName, db.RepoUrl, false); err != nil {
		logger.Error(err, "Failed to download Helm Repo", "repo", db.RepoUrl)
		return err
	}
	logger.Info("Added Helm Repo", "repo", db.RepoName)
	if err := kube.CreateNamespace(db.Namespace, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	return helm.InstallChart(ctx, db.Name, db.Namespace, db.ReleaseName, modela.Spec.SystemDatabase.Values.Object)
}

func (db Database) Installing(ctx context.Context) (bool, error) {
	installed, err := db.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := kube.IsPodRunning(db.Namespace, db.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

func (db Database) Ready(ctx context.Context) (bool, error) {
	installing, err := db.Installing(ctx)
	if err != nil && err != managementv1.ComponentNotInstalledByModelaError {
		return false, err
	}
	return !installing, nil
}

func (db Database) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	return helm.UninstallChart(ctx, db.Name, db.Namespace, db.ReleaseName, map[string]interface{}{})
}
