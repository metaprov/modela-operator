package components

import (
	"context"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/helm"
	"github.com/metaprov/modela-operator/pkg/kube"
	"github.com/pkg/errors"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Modela system represent the model core system
type ObjectStorage struct {
	Namespace     string
	Version       string
	ReleaseName   string
	RepoUrl       string
	RepoName      string
	Name          string
	PodNamePrefix string
	Dryrun        bool
}

func NewObjectStorage() *ObjectStorage {
	return &ObjectStorage{
		Namespace:   "modela-system",
		ReleaseName: "modela-storage",
		RepoName:    "bitnami",
		Name:        "minio",
		RepoUrl:     "https://charts.bitnami.com/bitnami",
		Dryrun:      false,
	}
}

func (os ObjectStorage) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingObjectStorage
}

func (os ObjectStorage) IsEnabled(modela managementv1.Modela) bool {
	return modela.Spec.ObjectStore.Install
}

// Check if the database installed
func (os ObjectStorage) Installed(ctx context.Context) (bool, error) {
	if belonging, err := kube.IsDeploymentCreatedByModela(os.Namespace, "modela-storage-minio"); err == nil && !belonging {
		return true, managementv1.ComponentNotInstalledByModelaError
	}
	if installed, err := helm.IsChartInstalled(ctx, os.Name, os.Namespace, os.ReleaseName); !installed {
		return false, err
	}

	return true, nil
}

func (os ObjectStorage) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := helm.AddRepo(os.RepoName, os.RepoUrl, false); err != nil {
		logger.Error(err, "Failed to download Helm Repo", "repo", os.RepoUrl)
		return err
	}
	logger.Info("Added Helm Repo", "repo", os.RepoName)
	if err := kube.CreateNamespace(os.Namespace, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	logger.Info("Applying Helm Chart", "version", os.Version)
	return helm.InstallChart(ctx, os.Name, os.Namespace, os.ReleaseName, map[string]interface{}{})
}

// Check if we are still installing the database
func (os ObjectStorage) Installing(ctx context.Context) (bool, error) {
	installed, err := os.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := kube.IsPodRunning(os.Namespace, os.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

func (os ObjectStorage) Ready(ctx context.Context) (bool, error) {
	installing, err := os.Installing(ctx)
	if err != nil && err != managementv1.ComponentNotInstalledByModelaError {
		return false, err
	}
	return !installing, nil
}

func (os ObjectStorage) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	return helm.UninstallChart(ctx, os.Name, os.Namespace, os.ReleaseName, map[string]interface{}{})
}

func (os ObjectStorage) PostInstall() error {

	values, err := kube.GetSecretValuesAsString("modela-system", "modela-storage-minio")

	// build the minio url
	accessKey, ok := values["root-user"]
	if !ok {
		return errors.New("key root-user is missing in the minio secret")
	}
	secertKey, ok := values["root-password"]
	if !ok {
		return errors.New("key root-password is missing in the minio secret")
	}
	// create a connection and update the fields.
	connection, err := kube.GetConnection("default-tenant", "default-minio")
	if err != nil {
		return err
	}
	host := "modela-storage-minio.modela-system.svc.cluster.local:9000"
	connection.Spec.Minio.Host = &host
	// save the connection
	err = kube.CreateOrUpdateConnection("default-tenant", connection.Name, connection)
	if err != nil {
		return err
	}
	defaultSecret, err := kube.GetSecret("default-tenant", "default-minio-secret")
	if err != nil {
		return err
	}
	values = make(map[string]string)
	values["accessKey"] = accessKey
	values["secretKey"] = secertKey
	err = kube.CreateOrUpdateSecret(defaultSecret.Namespace, defaultSecret.Name, values)
	return err
}
