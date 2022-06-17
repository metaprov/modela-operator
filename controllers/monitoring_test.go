package controllers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMonitoring_Install(t *testing.T) {
	monitoring := NewMonitoring()
	installed, err := monitoring.Installed()
	assert.Error(t, err)
	assert.False(t, installed)

	if !installed {
		err = monitoring.Install()
	}
}
