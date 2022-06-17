package controllers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultTenant_Installed(t *testing.T) {
	tenant := NewDefaultTenant("v0.4.716")
	installed, err := tenant.Installed()
	assert.NoError(t, err)
	assert.False(t, installed)

}

// run on an empty system
func TestDefaultTenant_Install(t *testing.T) {
	tenant := NewDefaultTenant("v0.4.716")

	err := tenant.Install()
	assert.NoError(t, err)

}
