package controllers

import (
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestModela_Installed(t *testing.T) {
	modela := NewModelaSystem("v0.4.716")
	installed, err := modela.Installed(context.Background())
	assert.NoError(t, err)
	assert.False(t, installed)

}

// run on an empty system
func TestModela_InstallAPI(t *testing.T) {
	modela := NewModelaSystem("v0.4.716")
	err := modela.InstallCRD(context.Background(), &v1alpha1.Modela{ObjectMeta: v1.ObjectMeta{Name: "modela-test"}})
	assert.NoError(t, err)
}

// run on an empty system
func TestModela_Install(t *testing.T) {
	modela := NewModelaSystem("develop")

	err := modela.Install(context.Background(), &v1alpha1.Modela{ObjectMeta: v1.ObjectMeta{Name: "modela-test"}})
	assert.NoError(t, err)
}
