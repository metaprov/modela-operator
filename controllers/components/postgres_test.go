package components

import (
	"context"
	"github.com/metaprov/modela-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Postgres installer", func() {
	database := NewPostgresDatabase()

	It("Should install postgres", func() {
		if installed, err := database.Installed(context.Background()); err == v1alpha1.ComponentNotInstalledByModelaError || installed {
			Skip("Test should be run on an empty cluster")
			return
		}

		Expect(database.Install(context.Background(), &v1alpha1.Modela{
			ObjectMeta: v1.ObjectMeta{Name: "modela-test"},
			Spec: v1alpha1.ModelaSpec{
				Database: v1alpha1.DatabaseSpec{
					InstallPgvector: true,
				},
			},
		})).To(BeNil())

		By("Checking if it was installed")
		installed, err := database.Installed(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeTrue())

		changeStatefulSetModelaOperatorLabel(false, "modela-system", "modela-postgresql")
		installed, err = database.Installed(context.Background())
		Expect(err).To(Equal(v1alpha1.ComponentNotInstalledByModelaError))
		/*
			By("Uninstalling postgres")
			Expect(database.Uninstall(context.Background(), &v1alpha1.Modela{})).To(BeNil())
			_, err = kube.IsStatefulSetCreatedByModela("modela-system", "modela-postgresql")
			Expect(k8serr.IsNotFound(err)).To(BeTrue())

			By("Checking if it was uninstalled")
			installed, err = database.Installed(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(installed).To(BeFalse())

		*/
	})
})
