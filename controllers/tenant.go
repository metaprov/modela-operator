package controllers

import (
	"context"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
)

type Tenant struct {
	Name string
}

func NewTenant(name string) *Tenant {
	return &Tenant{
		Name: name,
	}
}

func (t Tenant) IsEnabled(_ managementv1.Modela) bool {
	return true
}

func (t Tenant) Installed(ctx context.Context) (bool, error) {
	return IsNamespaceCreated(t.Name)
}

func (dt Tenant) Install(ctx context.Context, modela managementv1.Modela) error {
	// TODO
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
