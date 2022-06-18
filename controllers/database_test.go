package controllers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDatabase_Installed(t *testing.T) {
	t.Skip("Run only on empty cluster")
	database := NewDatabase()
	installed, err := database.Installed()
	assert.NoError(t, err)
	assert.True(t, installed)
}

// run on an empty system
func TestDatabase_Install(t *testing.T) {
	t.Skip("Run only on empty cluster")
	database := NewDatabase()

	err := database.Install()
	assert.NoError(t, err)

}
