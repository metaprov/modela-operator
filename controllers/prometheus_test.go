package controllers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrometheus_Installed(t *testing.T) {
	prem := NewPrometheus()
	installed, err := prem.Installed()
	assert.NoError(t, err)
	assert.False(t, installed)
}

// run on an empty system
func TestPrometheus_Install(t *testing.T) {
	prem := NewPrometheus()

	err := prem.Install()
	assert.NoError(t, err)

}
