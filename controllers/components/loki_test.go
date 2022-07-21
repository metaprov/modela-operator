package components

import (
	"context"
	"github.com/metaprov/modela-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Loki installer", func() {
	loki := NewLoki()

	It("Should install Loki", func() {
		if installed, err := loki.Installed(context.Background()); err == v1alpha1.ComponentNotInstalledByModelaError || installed {
			Skip("Test should be run on an empty cluster")
			return
		}

		Expect(loki.Install(context.Background(), &v1alpha1.Modela{
			ObjectMeta: v1.ObjectMeta{Name: "modela-test"},
		})).To(BeNil())
		/*
			By("Checking if it was installed")
			installed, err := loki.Installed(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(installed).To(BeTrue())

			changeModelaOperatorLabel(false, "modela-system", "modela-postgresql")
			installed, err = loki.Installed(context.Background())
			Expect(err).To(Equal(controllers.ComponentNotInstalledByModelaError))

			By("Uninstalling loki")
			Expect(loki.Uninstall(context.Background(), &v1alpha1.Modela{})).To(BeNil())
			_, err = controllers.IsDeploymentCreatedByModela("modela-system", "modela-postgresql")
			Expect(k8serr.IsNotFound(err)).To(BeTrue())

			By("Checking if it was uninstalled")
			installed, err = loki.Installed(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(installed).To(BeFalse())
		*/
	})
})
