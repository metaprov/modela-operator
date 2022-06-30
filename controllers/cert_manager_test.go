package controllers

import (
	"context"
	"testing"

	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestCertManager_Installed(t *testing.T) {
	certmanager := NewCertManager("v1.7.1")
	installed, err := certmanager.Installed(context.Background())
	assert.NoError(t, err)
	assert.True(t, installed)
}

func TestCertManager_Install(t *testing.T) {
	certmanager := NewCertManager("v1.7.1")

	err := certmanager.Install(context.Background(), &v1alpha1.Modela{})
	assert.NoError(t, err)
}

func TestCertManager_Uninstall(t *testing.T) {
	certmanager := NewCertManager("v1.7.1")

	err := certmanager.Uninstall(context.Background())
	assert.NoError(t, err)
	installed, err := certmanager.Installed(context.Background())
	assert.NoError(t, err)
	assert.False(t, installed)
}
