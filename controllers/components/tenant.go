package components

import (
	"context"
	"github.com/metaprov/modela-operator/pkg/kube"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kustomize/kyaml/kio"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
)

type Tenant struct {
	Name         string
	ManifestPath string
}

func NewTenant(name string) *Tenant {
	return &Tenant{
		Name:         name,
		ManifestPath: "tenant",
	}
}

func (t Tenant) IsEnabled(_ managementv1.Modela) bool {
	return true
}

func (t Tenant) Installed(ctx context.Context) (bool, error) {
	return kube.IsNamespaceCreated(t.Name)
}

func (t Tenant) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := kube.CreateNamespace(t.Name, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	yaml, _, err := kube.LoadResources(t.ManifestPath, []kio.Filter{
		kube.LabelFilter{Labels: map[string]string{"management.modela.ai/operator": modela.Name}},
		kube.NamespaceFilter{Namespace: t.Name},
	})
	if err != nil {
		return err
	}

	logger.Info("Applying tenant resources", "tenant", t.Name, "length", len(yaml))
	if err := kube.ApplyYaml(string(yaml)); err != nil {
		return err
	}

	return nil
}

func (dt Tenant) Installing(ctx context.Context) (bool, error) {
	installed, err := dt.Installed(ctx)
	if !installed {
		return installed, err
	}
	return false, nil
}

func (d Tenant) Ready(ctx context.Context) (bool, error) {
	return d.Installed(ctx)
}

func (d Tenant) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	if created, err := kube.IsNamespaceCreatedByOperator(d.Name, modela.Name); !created {
		return managementv1.ComponentNotInstalledByModelaError
	} else if err != nil {
		return err
	}

	if err := kube.DeleteNamespace(d.Name); err != nil {
		return err
	}

	return nil
}
