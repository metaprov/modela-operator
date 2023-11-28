package components

import (
	"context"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/helm"
	"github.com/metaprov/modela-operator/pkg/kube"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Postgres struct {
	Namespace     string
	Name          string
	ReleaseName   string
	PodNamePrefix string
	MongoMetadata *Postgres
}

func NewPostgresDatabase() *Postgres {
	return &Postgres{
		Namespace:     "modela-system",
		ReleaseName:   "modela-postgresql",
		Name:          "postgresql",
		PodNamePrefix: "modela-postgresql",
		MongoMetadata: &Postgres{
			Namespace:     "modela-system",
			ReleaseName:   "modela-mongodb",
			Name:          "mongodb",
			PodNamePrefix: "modela-mongodb",
		},
	}
}

func (db Postgres) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingDatabase
}

func (db Postgres) IsEnabled(modela managementv1.Modela) bool {
	return true
}

func (db Postgres) Installed(ctx context.Context) (bool, error) {
	if belonging, err := kube.IsStatefulSetCreatedByModela(db.Namespace, "modela-postgresql"); err == nil && !belonging {
		return true, managementv1.ComponentNotInstalledByModelaError
	}

	if installed, err := helm.IsChartInstalled(ctx, db.Name, db.Namespace, db.ReleaseName); !installed {
		return false, err
	}

	return true, nil
}

func (db Postgres) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := kube.CreateNamespace(db.Namespace, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	values := modela.Spec.Database.PostgresValues.Object
	if values == nil {
		values = make(map[string]interface{})
	}
	if modela.Spec.Database.InstallPgvector {
		imageValues := make(map[string]interface{})
		imageValues["registry"] = "docker.io"
		imageValues["repository"] = "ankane/pgvector"
		imageValues["tag"] = "v0.5.1"
		values["image"] = imageValues
	}

	return helm.InstallChart(ctx, db.Name, db.Namespace, db.ReleaseName, values)
}

func (db Postgres) Installing(ctx context.Context) (bool, error) {
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

func (db Postgres) Ready(ctx context.Context) (bool, error) {
	installing, err := db.Installing(ctx)
	if err != nil && err != managementv1.ComponentNotInstalledByModelaError {
		return false, err
	}
	return !installing, nil
}

func (db Postgres) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	return helm.UninstallChart(ctx, db.Name, db.Namespace, db.ReleaseName, map[string]interface{}{})
}
