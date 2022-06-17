package controllers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModela_Installed(t *testing.T) {
	modela := NewModelaSystem("v0.4.716")
	installed, err := modela.Installed()
	assert.NoError(t, err)
	assert.False(t, installed)

}

// run on an empty system
func TestModela_Install(t *testing.T) {
	modela := NewModelaSystem("v0.4.716")

	err := modela.Install()
	assert.NoError(t, err)

}
