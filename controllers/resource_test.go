package controllers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const TEST_YAML = `
apiVersion: v1
kind: Namespace
metadata:
  name: test-yaml2
`

func TestResources_LoadResources(t *testing.T) {
	output, err := LoadResources("modela-system", nil)
	assert.Nil(t, err)
	fmt.Println(string(output))
}

func TestResources_Apply(t *testing.T) {
	err := ApplyYaml(TEST_YAML)
	assert.Nil(t, err)
}
