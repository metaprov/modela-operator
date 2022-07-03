package controllers

import (
	"context"
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
	return IsNamespaceCreated(t.Name)
}

func (t Tenant) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := CreateNamespace(t.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	yaml, err := LoadResources(t.ManifestPath, []kio.Filter{
		LabelFilter{Labels: map[string]string{"management.modela.ai/operator": modela.Name}},
		NamespaceFilter{Namespace: t.Name},
	})
	if err != nil {
		return err
	}

	logger.Info("Applying tenant resources", "tenant", t.Name, "length", len(yaml))
	if err := ApplyYaml(string(yaml)); err != nil {
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

// Check if the default tenant is installed and ready
func (d Tenant) Ready(ctx context.Context) (bool, error) {
	return d.Installed(ctx)
}

func (d Tenant) Uninstall(ctx context.Context) error {
	return nil
}
