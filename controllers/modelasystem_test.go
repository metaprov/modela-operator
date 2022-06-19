package controllers

import (
	"context"
	"testing"

	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestModela_Installed(t *testing.T) {
	t.Skip("Run only on empty cluster")
	modela := NewModelaSystem("v0.4.716")
	installed, err := modela.Installed(context.Background())
	assert.NoError(t, err)
	assert.False(t, installed)

}

// run on an empty system
func TestModela_Install(t *testing.T) {
	t.Skip("Run only on empty cluster")
	modela := NewModelaSystem("v0.4.716")

	err := modela.Install(context.Background(), v1alpha1.Modela{})
	assert.NoError(t, err)

}
