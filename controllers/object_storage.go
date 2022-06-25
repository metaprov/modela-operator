package controllers

import (
	"context"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/pkg/errors"
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

func NewObjectStorage(version string) *ObjectStorage {
	return &ObjectStorage{
		Namespace: "modela-system",
		//Version:     "9.2.9",
		Version:     version,
		ReleaseName: "modela-storage",
		RepoName:    "bitnami",
		Name:        "minio",
		RepoUrl:     "https://charts.bitnami.com/bitnami",
		Dryrun:      false,
	}
}

func (os ObjectStorage) IsEnabled(modela managementv1.Modela) bool {
	return *modela.Spec.ObjectStore.Installed

}

// Check if the database installed
func (os ObjectStorage) Installed(ctx context.Context) (bool, error) {
	//repoName, repoUrl string, url string, ns string, releaseName string, versionName string
	return IsChartInstalled(
		ctx,
		os.RepoName,
		os.RepoUrl,
		os.Name,
		os.Namespace,
		os.ReleaseName,
		os.Version,
	)
}

func (os ObjectStorage) Install(ctx context.Context, modela managementv1.Modela) error {
	logger := log.FromContext(ctx)
	if err := AddRepo(os.RepoName, os.RepoUrl, os.Dryrun); err != nil {
		logger.Error(err, "Failed to add repo "+os.RepoName)
		return err
	}

	logger.Info("added repo " + os.RepoName)
	// install namespace modela-system
	if err := CreateNamespace(os.Namespace); err != nil {
		logger.Error(err, "Failed to create cert-manager namespace "+os.Namespace)
		return err
	}
	logger.Info("created namespace " + os.Namespace)

	return InstallChart(
		ctx,
		os.RepoName,
		os.RepoUrl,
		os.Name,
		os.Namespace,
		os.ReleaseName,
		os.Version,
	)
}

// Check if we are still installing the database
func (os ObjectStorage) Installing(ctx context.Context) (bool, error) {
	installed, err := os.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(os.Namespace, os.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

// Check if the database is ready
func (os ObjectStorage) Ready(ctx context.Context) (bool, error) {
	installed, err := os.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(os.Namespace, os.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return running, nil
}

func (os ObjectStorage) Uninstall(ctx context.Context) error {
	return UninstallChart(
		ctx,
		os.RepoName,
		os.RepoUrl,
		"",
		os.Namespace,
		os.ReleaseName,
		os.Version,
	)
}

func (os ObjectStorage) PostInstall() error {

	values, err := GetSecretValuesAsString("modela-system", "modela-storage-minio")

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
	connection, err := GetConnection("default-tenant", "default-minio")
	if err != nil {
		return err
	}
	host := "modela-storage-minio.modela-system.svc.cluster.local:9000"
	connection.Spec.Minio.Host = &host
	// save the connection
	err = CreateOrUpdateConnection("default-tenant", connection.Name, connection)
	if err != nil {
		return err
	}
	defaultSecret, err := GetSecret("default-tenant", "default-minio-secret")
	if err != nil {
		return err
	}
	values = make(map[string]string)
	values["accessKey"] = accessKey
	values["secretKey"] = secertKey
	err = CreateOrUpdateSecret(defaultSecret.Namespace, defaultSecret.Name, values)
	return err
}
