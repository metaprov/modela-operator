package controllers

import (
	"context"
	"testing"

	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestDefaultTenant_Installed(t *testing.T) {
	t.Skip("Run only on empty cluster")
	tenant := NewDefaultTenant("v0.4.716")
	installed, err := tenant.Installed()
	assert.NoError(t, err)
	assert.False(t, installed)

}

// run on an empty system
func TestDefaultTenant_Install(t *testing.T) {
	t.Skip("Run only on empty cluster")
	tenant := NewDefaultTenant("v0.4.716")

	err := tenant.Install(context.Background(), v1alpha1.Modela{})
	assert.NoError(t, err)

}
