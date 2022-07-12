package components

import (
	"context"
	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/kube"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Postgres installer", func() {
	database := NewDatabase("")

	It("Should install postgres", func() {
		if installed, err := database.Installed(context.Background()); err == v1alpha1.ComponentNotInstalledByModelaError || installed {
			Skip("Test should be run on an empty cluster")
			return
		}

		Expect(database.Install(context.Background(), &v1alpha1.Modela{
			ObjectMeta: v1.ObjectMeta{Name: "modela-test"},
		})).To(BeNil())

		By("Checking if it was installed")
		installed, err := database.Installed(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeTrue())

		changeModelaOperatorLabel(false, "modela-system", "modela-postgresql")
		installed, err = database.Installed(context.Background())
		Expect(err).To(Equal(v1alpha1.ComponentNotInstalledByModelaError))

		By("Uninstalling postgres")
		Expect(database.Uninstall(context.Background(), &v1alpha1.Modela{})).To(BeNil())
		_, err = kube.IsDeploymentCreatedByModela("modela-system", "modela-postgresql")
		Expect(k8serr.IsNotFound(err)).To(BeTrue())

		By("Checking if it was uninstalled")
		installed, err = database.Installed(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeFalse())
	})
})
