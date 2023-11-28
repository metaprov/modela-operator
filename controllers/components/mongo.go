package components

import (
	"context"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/helm"
	"github.com/metaprov/modela-operator/pkg/kube"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Mongo struct {
	Namespace     string
	Name          string
	ReleaseName   string
	PodNamePrefix string
	MongoMetadata *Mongo
}

func NewMongoDatabase() *Mongo {
	return &Mongo{
		Namespace:     "modela-system",
		ReleaseName:   "modela-mongodb",
		Name:          "mongodb",
		PodNamePrefix: "modela-mongodb",
	}
}

func (db Mongo) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingDatabase
}

func (db Mongo) IsEnabled(modela managementv1.Modela) bool {
	return modela.Spec.Database.InstallMongoDB
}

func (db Mongo) Installed(ctx context.Context) (bool, error) {
	if belonging, err := kube.IsStatefulSetCreatedByModela(db.Namespace, "modela-mongodb"); err == nil && !belonging {
		return true, managementv1.ComponentNotInstalledByModelaError
	}

	if installed, err := helm.IsChartInstalled(ctx, db.Name, db.Namespace, db.ReleaseName); !installed {
		return false, err
	}

	return true, nil
}

func (db Mongo) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := kube.CreateNamespace(db.Namespace, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	values := modela.Spec.Database.MongoDBValues.Object
	if values == nil {
		values = make(map[string]interface{})
	}
	values["useStatefulSet"] = true
	return helm.InstallChart(ctx, db.Name, db.Namespace, db.ReleaseName, values)
}

func (db Mongo) Installing(ctx context.Context) (bool, error) {
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

func (db Mongo) Ready(ctx context.Context) (bool, error) {
	installing, err := db.Installing(ctx)
	if err != nil && err != managementv1.ComponentNotInstalledByModelaError {
		return false, err
	}
	return !installing, nil
}

func (db Mongo) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	return helm.UninstallChart(ctx, db.Name, db.Namespace, db.ReleaseName, map[string]interface{}{})
}
