package kube

const TEST_YAML = `
apiVersion: v1
kind: Namespace
metadata:
  name: test-yaml2
`

/*
func TestResources_LoadModelaSystem(t *testing.T) {
	output, _, err := LoadResources("modela-system", []kio.Filter{ContainerVersionFilter{"1.0.0"}})
	assert.Nil(t, err)
	fmt.Println(string(output))
}

func TestResources_LoadTenant(t *testing.T) {
	output, _, err := LoadResources("tenant", []kio.Filter{NamespaceFilter{"tenant-test"}})
	assert.Nil(t, err)
	assert.False(t, strings.Contains(string(output), "default-tenant"))
	fmt.Println(string(output))
}

func TestResources_LoadCatalog(t *testing.T) {
	output, _, err := LoadResources("modela-catalog", []kio.Filter{ManagedImageFilter{"1.0.0"}})
	assert.Nil(t, err)
	assert.False(t, strings.Contains(string(output), "tag: latest"))
	fmt.Println(string(output))
}

func TestResources_Apply(t *testing.T) {
	err := ApplyYaml(TEST_YAML)
	assert.Nil(t, err)
}*/
