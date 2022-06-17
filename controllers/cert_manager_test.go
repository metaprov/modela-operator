package controllers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCertManager_Install(t *testing.T) {
	database := NewCertManager()

	installed, err := database.Installed()
	assert.NoError(t, err)
	assert.True(t, installed)

}
