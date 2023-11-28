package components

import (
	"context"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/helm"
	"github.com/metaprov/modela-operator/pkg/kube"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

// Modela system represent the model core system
type OnlineStore struct {
	Namespace     string
	Version       string
	ReleaseName   string
	Name          string
	PodNamePrefix string
	ManifestPath  string
	Dryrun        bool
}

func NewOnlineStore() *OnlineStore {
	return &OnlineStore{
		Namespace:     "modela-system",
		ManifestPath:  "online-store",
		PodNamePrefix: "modela-online-store",
		ReleaseName:   "modela-redis",
		Name:          "redis",
	}
}

func (os OnlineStore) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingOnlineStore
}

func (os OnlineStore) IsEnabled(modela managementv1.Modela) bool {
	return modela.Spec.OnlineStore.Install
}

// Check if the database installed
func (os OnlineStore) Installed(ctx context.Context) (bool, error) {
	if created, err := kube.IsNamespaceCreated("modela-system"); !created || err != nil {
		return created, err
	}
	if _, missing, err := kube.LoadResources(os.ManifestPath, nil, false); missing > 0 {
		log.FromContext(ctx).Info("Resources detected as missing from the modela-system namespace", "count", missing)
		return false, managementv1.ComponentMissingResourcesError
	} else if err != nil {
		return false, err
	}

	if installed, err := helm.IsChartInstalled(ctx, os.Name, os.Namespace, os.ReleaseName); !installed {
		return false, err
	}

	return true, nil
}

func (os OnlineStore) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := kube.CreateNamespace(os.Namespace, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	logger.Info("Applying Helm Chart", "version", os.Version)
	if installed, err := helm.IsChartInstalled(ctx, os.Name, os.Namespace, os.ReleaseName); !installed {
		if err = helm.InstallChart(ctx, os.Name, os.Namespace, os.ReleaseName, modela.Spec.OnlineStore.Values.Object); err != nil {
			return err
		}
	}

	var password string
	if values, err := kube.GetSecretValuesAsString("modela-system", "redis"); err == nil {
		password, _ = values["redis-password"]
	}

	yaml, _, err := kube.LoadResources(os.ManifestPath, []kio.Filter{
		kube.LabelFilter{Labels: map[string]string{"management.modela.ai/operator": modela.Name}},
		kube.NamespaceFilter{Namespace: os.Namespace},
		kube.RedisSecretFilter{Password: password},
		kube.ContainerVersionFilter{Version: modela.Spec.Distribution},
		kube.OwnerReferenceFilter{Owner: modela.GetName(), OwnerNamespace: modela.GetNamespace(), UID: string(modela.GetUID())},
	}, false)
	if err != nil {
		return err
	}

	logger.Info("Applying online store resources", "length", len(yaml))
	if err := kube.ApplyYaml(string(yaml)); err != nil {
		return err
	}

	return nil
}

func (os OnlineStore) InstallNewVersion(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	var password string
	if values, err := kube.GetSecretValuesAsString("modela-system", "redis"); err == nil {
		password, _ = values["redis-password"]
	}

	yaml, _, err := kube.LoadResources(os.ManifestPath, []kio.Filter{
		kube.LabelFilter{Labels: map[string]string{"management.modela.ai/operator": modela.Name}},
		kube.NamespaceFilter{Namespace: os.Namespace},
		kube.RedisSecretFilter{Password: password},
		kube.ContainerVersionFilter{Version: modela.Spec.Distribution},
		kube.OwnerReferenceFilter{Owner: modela.GetName(), OwnerNamespace: modela.GetNamespace(), UID: string(modela.GetUID())},
	}, false)
	if err != nil {
		return err
	}

	logger.Info("Applying online store resources", "length", len(yaml), "ns", os.Namespace)
	if err := kube.ApplyYaml(string(yaml)); err != nil {
		return err
	}

	return nil
}

func (os OnlineStore) Installing(ctx context.Context) (bool, error) {
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

func (os OnlineStore) Ready(ctx context.Context) (bool, error) {
	installing, err := os.Installing(ctx)
	if err != nil && err != managementv1.ComponentNotInstalledByModelaError {
		return false, err
	}
	return !installing, nil
}

func (os OnlineStore) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	return helm.UninstallChart(ctx, os.Name, os.Namespace, os.ReleaseName, map[string]interface{}{})
}
