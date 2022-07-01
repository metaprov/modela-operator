package controllers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"strings"
	"testing"
)

const TEST_YAML = `
apiVersion: v1
kind: Namespace
metadata:
  name: test-yaml2
`

func TestResources_LoadModelaSystem(t *testing.T) {
	output, err := LoadResources("modela-system", []kio.Filter{ContainerVersionFilter{"1.0.0"}})
	assert.Nil(t, err)
	fmt.Println(string(output))
}

func TestResources_LoadTenant(t *testing.T) {
	output, err := LoadResources("tenant", []kio.Filter{NamespaceFilter{"tenant-test"}})
	assert.Nil(t, err)
	assert.False(t, strings.Contains(string(output), "default-tenant"))
	fmt.Println(string(output))
}

func TestResources_Apply(t *testing.T) {
	err := ApplyYaml(TEST_YAML)
	assert.Nil(t, err)
}
