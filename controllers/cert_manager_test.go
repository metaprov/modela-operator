package controllers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCertManager_Installed(t *testing.T) {
	t.Skip("Run only on empty cluster")
	database := NewCertManager()
	installed, err := database.Installed()
	assert.NoError(t, err)
	assert.True(t, installed)
}

// run on an empty system
func TestCertManager_Install(t *testing.T) {
	t.Skip("Run only on empty cluster")
	database := NewCertManager()

	err := database.Install()
	assert.NoError(t, err)

}
