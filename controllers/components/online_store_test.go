package components

import (
	"context"
	"github.com/metaprov/modela-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Online store installer", func() {
	onlineStore := NewOnlineStore()

	It("Should install redis", func() {
		if installed, err := onlineStore.Installed(context.Background()); err == v1alpha1.ComponentNotInstalledByModelaError || installed {
			Skip("Test should be run on an empty cluster")
			return
		}

		Expect(onlineStore.Install(context.Background(), &v1alpha1.Modela{
			ObjectMeta: v1.ObjectMeta{Name: "modela-test"},
		})).To(BeNil())

		By("Checking if it was installed")
		installed, err := onlineStore.Installed(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeTrue())
	})
})
