package controllers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestObjectStorage_Installed(t *testing.T) {
	t.Skip("Run only on empty cluster")
	prem := NewObjectStorage()
	installed, err := prem.Installed()
	assert.NoError(t, err)
	assert.False(t, installed)
}

// run on an empty system
func TestObjectStorage_Install(t *testing.T) {
	t.Skip("Run only on empty cluster")
	prem := NewObjectStorage()

	err := prem.Install()
	assert.NoError(t, err)

}
