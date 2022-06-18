package controllers

import (
	"context"
	"testing"

	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

const PostgresVersion = "1.1.1"

func TestDatabase_Installed(t *testing.T) {
	t.Skip("Run only on empty cluster")
	database := NewDatabase(PostgresVersion)
	installed, err := database.Installed()
	assert.NoError(t, err)
	assert.True(t, installed)
}

// run on an empty system
func TestDatabase_Install(t *testing.T) {
	t.Skip("Run only on empty cluster")
	database := NewDatabase(PostgresVersion)

	err := database.Install(context.Background(), v1alpha1.Modela{})
	assert.NoError(t, err)

}
