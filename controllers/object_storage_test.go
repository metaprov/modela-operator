package controllers

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"

	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const ObjectVersion = ""

func TestObjectStorage_Installed(t *testing.T) {
	prem := NewObjectStorage(ObjectVersion)
	installed, err := prem.Installed(context.Background())
	assert.NoError(t, err)
	assert.False(t, installed)
}

// run on an empty system
func TestObjectStorage_Install(t *testing.T) {
	prem := NewObjectStorage(ObjectVersion)

	err := prem.Install(context.Background(), &v1alpha1.Modela{})
	assert.NoError(t, err)
}

func TestObjectStorage_Uninstall(t *testing.T) {
	prem := NewObjectStorage(ObjectVersion)

	err := prem.Uninstall(context.Background(), nil)
	assert.NoError(t, err)
	installed, err := prem.Installed(context.Background())
	assert.NoError(t, err)
	assert.False(t, installed)
}

func init() {
	log.SetLogger(zap.New(zap.ConsoleEncoder()))
}
