package controllers

import (
	"context"
	"testing"

	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestPrometheus_Installed(t *testing.T) {
	t.Skip("Run only on empty cluster")
	prem := NewPrometheus("")
	installed, err := prem.Installed()
	assert.NoError(t, err)
	assert.False(t, installed)
}

// run on an empty system
func TestPrometheus_Install(t *testing.T) {
	t.Skip("Run only on empty cluster")
	prem := NewPrometheus("")

	err := prem.Install(context.Background(), v1alpha1.Modela{})
	assert.NoError(t, err)

}
