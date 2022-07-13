package components

import (
	"context"
	"encoding/json"
	"fmt"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/kube"
	infra "github.com/metaprov/modelaapi/pkg/apis/infra/v1alpha1"
	"golang.org/x/mod/semver"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

// ModelaSystem represents an installation of the Modela core system (control plane, API gateway, etc.)
type ModelaSystem struct {
	ModelaVersion       string
	Namespace           string
	CatalogNamespace    string
	SystemManifestPath  string
	CatalogManifestPath string
	CrdUrl              string
	VersionMatrixUrl    string
	PodNamePrefix       string
}

func (m ModelaSystem) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingModela
}

func (m ModelaSystem) IsEnabled(_ managementv1.Modela) bool {
	return true
}

func NewModelaSystem(version string) *ModelaSystem {
	return &ModelaSystem{
		ModelaVersion:       version,
		Namespace:           "modela-system",
		CatalogNamespace:    "modela-catalog",
		SystemManifestPath:  "modela-system",
		CatalogManifestPath: "modela-catalog",
		CrdUrl:              "github.com/metaprov/modelaapi/manifests/%s/base/crd",
		VersionMatrixUrl:    "https://raw.githubusercontent.com/metaprov/modelaapi/main/version_matrix.json",
		PodNamePrefix:       "modela-control-plane",
	}
}

func (ms ModelaSystem) Installed(ctx context.Context) (bool, error) {
	if created, err := kube.IsNamespaceCreated("modela-system"); !created || err != nil {
		return created, err
	}
	if _, missing, err := kube.LoadResources(ms.SystemManifestPath, nil); missing > 0 {
		log.FromContext(ctx).Info("Resources detected as missing from the modela-system namespace", "count", missing)
		return false, managementv1.ComponentMissingResourcesError
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (ms ModelaSystem) CatalogInstalled(ctx context.Context) (bool, error) {
	if created, err := kube.IsNamespaceCreated("modela-catalog"); !created || err != nil {
		return created, err
	}
	if _, missing, err := kube.LoadResources(ms.CatalogManifestPath, nil); missing > 0 {
		log.FromContext(ctx).Info("Resources detected as missing from the modela-catalog namespace", "count", missing)
		return false, managementv1.ComponentMissingResourcesError
	} else if err != nil {
		return false, err
	}
	return true, nil
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
	if version := kube.GetCRDVersion("tenants.infra.modela.ai"); version == finalVersion {
		logger.Info(fmt.Sprintf("CRD version %s already installed; skipping CRD installation", finalVersion))
		return nil
	}

	// Install the determined CRD version using Kustomize
	logger.Info(fmt.Sprintf("Installing CRD version %s", finalVersion))
	return kube.ApplyUrlKustomize(fmt.Sprintf(ms.CrdUrl, finalVersion))
}

func (ms ModelaSystem) InstallManagedImages(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	yaml, _, err := kube.LoadResources(ms.CatalogManifestPath+"/managedimages", []kio.Filter{
		kube.LabelFilter{Labels: map[string]string{"management.modela.ai/operator": modela.Name}},
		kube.ManagedImageFilter{Version: ms.ModelaVersion},
	})
	if err != nil {
		return err
	}

	logger.Info("Applying modela-catalog ManagedImage resources", "length", len(yaml))
	if err := kube.ApplyYaml(string(yaml)); err != nil {
		return err
	}

	return nil
}

func (ms ModelaSystem) InstallLicense(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if modela.Spec.License.LinkLicense == nil {
		return nil
	}

	if err := kube.CreateOrUpdateSecret("modela-system", "license-secret", map[string]string{
		"token": *modela.Spec.License.LicenseKey,
	}); err != nil {
		logger.Error(err, "Failed to update license secret")
		return err
	}

	now := metav1.Now()
	if err := kube.CreateOrUpdateLicense("modela-system", "modela-license", &infra.License{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "modela-license",
			Namespace: "modela-system",
		},
		Spec: infra.LicenseSpec{
			SecretRef: v1.SecretReference{
				Namespace: "modela-system",
				Name:      "license-secret",
			},
		},
		Status: infra.LicenseStatus{
			LastUpdated: &now,
		},
	}); err != nil {
		logger.Error(err, "Failed to update license object")
		return err
	}

	return nil
}

func (ms ModelaSystem) InstallCatalog(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	yaml, _, err := kube.LoadResources(ms.CatalogManifestPath, []kio.Filter{
		kube.LabelFilter{Labels: map[string]string{"management.modela.ai/operator": modela.Name}},
	})
	if err != nil {
		return err
	}

	if err := kube.CreateNamespace(ms.CatalogNamespace, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	logger.Info("Applying modela-catalog resources", "length", len(yaml))
	if err := kube.ApplyYaml(string(yaml)); err != nil {
		return err
	}

	if err := ms.InstallLicense(ctx, modela); err != nil {
		return err
	}
	return ms.InstallManagedImages(ctx, modela)
}

func (ms ModelaSystem) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)
	if err := ms.InstallCRD(ctx, modela); err != nil {
		return err
	}

	if err := kube.CreateNamespace(ms.Namespace, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	yaml, _, err := kube.LoadResources(ms.SystemManifestPath, []kio.Filter{
		kube.ContainerVersionFilter{ms.ModelaVersion},
		kube.LabelFilter{Labels: map[string]string{"management.modela.ai/operator": modela.Name}},
	})
	if err != nil {
		return err
	}

	logger.Info("Applying modela-system resources", "length", len(yaml))
	if err := kube.ApplyYaml(string(yaml)); err != nil {
		return err
	}

	return nil
}

func (ms ModelaSystem) Installing(ctx context.Context) (bool, error) {
	installed, err := ms.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := kube.IsPodRunning(ms.Namespace, ms.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

func (ms ModelaSystem) Ready(ctx context.Context) (bool, error) {
	installing, err := ms.Installing(ctx)
	if err != nil {
		return false, err
	}
	return !installing, nil
}

func (ms ModelaSystem) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	if created, err := kube.IsNamespaceCreatedByOperator(ms.Namespace, modela.Name); !created {
		return managementv1.ComponentNotInstalledByModelaError
	} else if err != nil {
		return err
	}

	if err := kube.DeleteNamespace(ms.Namespace); err != nil {
		return err
	}

	return nil
}
