package controllers

import (
	"fmt"
	"github.com/metaprov/modela-operator/internal/pkg/util"
	"github.com/pkg/errors"
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
		Version:     "9.2.9",
		ReleaseName: "modela-storage",
		RepoName:    "bitnami",
		Name:        "minio",
		RepoUrl:     "https://charts.bitnami.com/bitnami",
		Dryrun:      false,
	}
}

// Check if the database installed
func (os ObjectStorage) Installed() (bool, error) {
	//repoName, repoUrl string, url string, ns string, releaseName string, versionName string
	return util.IsChartInstalled(
		os.RepoName,
		os.RepoUrl,
		os.Name,
		os.Namespace,
		os.ReleaseName,
		os.Version,
	)
}

func (os ObjectStorage) Install() error {

	if err := util.AddRepo(os.RepoName, os.RepoUrl, os.Dryrun); err != nil {
		return err
	}
	fmt.Println("\u2713 added repo " + os.RepoName)
	// install namespace modela-system
	if err := util.CreateNamespace(os.Namespace); err != nil {
		return err
	}
	fmt.Println("\u2713 created namespace " + os.Namespace)

	return util.InstallChart(
		os.RepoName,
		os.RepoUrl,
		os.Name,
		os.Namespace,
		os.ReleaseName,
		os.Version,
	)
}

// Check if we are still installing the database
func (os ObjectStorage) Installing() (bool, error) {
	installed, err := os.Installed()
	if !installed {
		return installed, err
	}
	running, err := util.IsPodRunning(os.Namespace, os.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

// Check if the database is ready
func (os ObjectStorage) Ready() (bool, error) {
	installed, err := os.Installed()
	if !installed {
		return installed, err
	}
	running, err := util.IsPodRunning(os.Namespace, os.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return running, nil
}

func (os ObjectStorage) Uninstall() error {
	return util.UninstallChart(
		os.RepoName,
		os.RepoUrl,
		"",
		os.Namespace,
		os.ReleaseName,
		os.Version,
	)
}

func (os ObjectStorage) PostInstall() error {

	values, err := util.GetSecretValuesAsString("modela-system", "modela-storage-minio")

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
	connection, err := util.GetConnection("default-tenant", "default-minio")
	if err != nil {
		return err
	}
	host := "modela-storage-minio.modela-system.svc.cluster.local:9000"
	connection.Spec.Minio.Host = &host
	// save the connection
	err = util.CreateOrUpdateConnection("default-tenant", connection.Name, connection)
	if err != nil {
		return err
	}
	defaultSecret, err := util.GetSecret("default-tenant", "default-minio-secret")
	if err != nil {
		return err
	}
	values = make(map[string]string)
	values["accessKey"] = accessKey
	values["secretKey"] = secertKey
	err = util.CreateOrUpdateSecret(defaultSecret.Namespace, defaultSecret.Name, values)
	return err
}
