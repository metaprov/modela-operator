package controllers

import (
	"context"
	"testing"

	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestTenant_Installed(t *testing.T) {
	tenant := NewTenant("default-tenant")
	installed, err := tenant.Installed(context.Background())
	assert.NoError(t, err)
	assert.False(t, installed)
}

func TestDefaultTenant_Install(t *testing.T) {
	tenant := NewTenant("default-tenant")
	err := tenant.Install(context.Background(), &v1alpha1.Modela{})
	assert.NoError(t, err)

}
