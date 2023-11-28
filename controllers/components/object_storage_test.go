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

const ObjectVersion = ""

var _ = Describe("Object storage installer", func() {
	objectStorage := NewObjectStorage()

	It("Should install minio", func() {
		if installed, err := objectStorage.Installed(context.Background()); err == v1alpha1.ComponentNotInstalledByModelaError || installed {
			Skip("Test should be run on an empty cluster")
			return
		}

		Expect(objectStorage.Install(context.Background(), &v1alpha1.Modela{
			ObjectMeta: v1.ObjectMeta{Name: "modela-test"},
		})).To(BeNil())

		By("Checking if it was installed")
		installed, err := objectStorage.Installed(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeTrue())

		changeDeploymentModelaOperatorLabel(false, "modela-system", "modela-storage-minio")
		installed, err = objectStorage.Installed(context.Background())
		Expect(err).To(Equal(v1alpha1.ComponentNotInstalledByModelaError))

		By("Uninstalling minio")
		Expect(objectStorage.Uninstall(context.Background(), &v1alpha1.Modela{})).To(BeNil())
		_, err = kube.IsDeploymentCreatedByModela("modela-system", "modela-storage-minio")
		Expect(k8serr.IsNotFound(err)).To(BeTrue())

		By("Checking if it was uninstalled")
		installed, err = objectStorage.Installed(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeFalse())
	})
})
