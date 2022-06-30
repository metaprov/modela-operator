package controllers

import (
	"context"
	"testing"

	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

const PostgresVersion = ""

func TestDatabase_Installed(t *testing.T) {
	database := NewDatabase(PostgresVersion)
	installed, err := database.Installed(context.Background())
	assert.NoError(t, err)
	assert.False(t, installed)
}

func TestDatabase_Install(t *testing.T) {
	database := NewDatabase(PostgresVersion)

	err := database.Install(context.Background(), &v1alpha1.Modela{})
	assert.NoError(t, err)
}

func TestDatabase_Uninstall(t *testing.T) {
	database := NewDatabase(PostgresVersion)

	err := database.Uninstall(context.Background())
	assert.NoError(t, err)
	installed, err := database.Installed(context.Background())
	assert.NoError(t, err)
	assert.False(t, installed)
}
