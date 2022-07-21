package kube

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

var _ = Describe("Resource filter", func() {
	It("Should add a controller reference", func() {
		yaml, _, err := LoadResources("../../manifests/modela-system", []kio.Filter{
			OwnerReferenceFilter{
				Owner: "modela",
				UID:   "abc-123",
			},
		}, true)
		Expect(err).To(BeNil())
		fmt.Println(string(yaml))
	})
	It("Should add minio secret and access keys", func() {
		yaml, _, err := LoadResources("../../manifests/tenant", []kio.Filter{
			MinioSecretFilter{
				AccessKey: "test123",
				SecretKey: "testabc",
			},
		}, true)
		Expect(err).To(BeNil())
		fmt.Println(string(yaml))
	})
	It("Should change default tenant objects", func() {
		yaml, _, err := LoadResources("../../manifests/tenant", []kio.Filter{
			TenantFilter{TenantName: "test-tenant"},
		}, true)
		Expect(err).To(BeNil())
		fmt.Println(string(yaml))
	})
})
