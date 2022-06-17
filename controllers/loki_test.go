package controllers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoki_Installed(t *testing.T) {
	loki := NewLoki()
	installed, err := loki.Installed()
	assert.Error(t, err)
	assert.False(t, installed)

}

// run on an empty system
func TestLoki_Install(t *testing.T) {
	loki := NewLoki()

	err := loki.Install()
	assert.NoError(t, err)

}
