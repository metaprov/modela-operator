package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/mod/semver"
	"io/ioutil"
	"net/http"
	"sigs.k8s.io/kustomize/kyaml/kio"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ModelaSystem represents an installation of the Modela core system (control plane, API gateway, etc.)
type ModelaSystem struct {
	ModelaVersion    string
	Namespace        string
	ManifestPath     string
	CrdUrl           string
	VersionMatrixUrl string
	PodNamePrefix    string
}

func (m ModelaSystem) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingModela
}

func (m ModelaSystem) IsEnabled(_ managementv1.Modela) bool {
	return true
}

func NewModelaSystem(version string) *ModelaSystem {
	return &ModelaSystem{
		ModelaVersion:    version,
		Namespace:        "modela-system",
		ManifestPath:     "manifests/modela-system",
		CrdUrl:           "github.com/metaprov/modelaapi/manifests/%s/base/crd",
		VersionMatrixUrl: "https://raw.githubusercontent.com/metaprov/modelaapi/main/version_matrix.json",
		PodNamePrefix:    "modela-control-plane",
	}
}

// Check if the database installed
func (ms ModelaSystem) Installed(ctx context.Context) (bool, error) {
	return false, nil
}

func (ms ModelaSystem) InstallCRD(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	// Download the version matrix, which associates a minimum Modela version for each API version
	resp, err := http.Get(ms.VersionMatrixUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return err
	}

	// Determine the required version based on the version closest to the Modela version
	versionData := jsonData.(map[string]interface{})
	var finalVersion string
	var versions []string
	for version, _ := range versionData {
		versions = append(versions, version)
	}
	semver.Sort(versions)

	for _, version := range versions {
		if semver.Compare(ms.ModelaVersion, version) >= 0 {
			finalVersion = versionData[version].(string)
		}
	}

	if ms.ModelaVersion == "develop" {
		finalVersion = "v1alpha1"
	}

	// Check if the version is already installed
	if version := GetCRDVersion("tenants.infra.modela.ai"); version == finalVersion {
		logger.Info(fmt.Sprintf("CRD version %s already installed; skipping CRD installation", finalVersion))
		return nil
	}

	// Install the determined CRD version using Kustomize
	logger.Info(fmt.Sprintf("Installing CRD version %s", finalVersion))
	return ApplyUrlKustomize(fmt.Sprintf(ms.CrdUrl, finalVersion))
}

func (ms ModelaSystem) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)
	if err := ms.InstallCRD(ctx, modela); err != nil {
		return err
	}

	yaml, err := LoadResources("modela-system", []kio.Filter{
		ContainerVersionFilter{ms.ModelaVersion},
		LabelFilter{Labels: map[string]string{"management.modela.ai/operator": modela.Name}},
	})
	if err != nil {
		return err
	}

	logger.Info("Applying modela-system resources", "length", len(yaml))
	if err := ApplyYaml(string(yaml)); err != nil {
		return err
	}

	return nil
}

func (ms ModelaSystem) Installing(ctx context.Context) (bool, error) {
	installed, err := ms.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(ms.Namespace, ms.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

func (ds ModelaSystem) Ready(ctx context.Context) (bool, error) {
	installed, err := ds.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := IsPodRunning(ds.Namespace, ds.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return running, nil
}

func (dm ModelaSystem) Uninstall(ctx context.Context) error {
	return nil
}
