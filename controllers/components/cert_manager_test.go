package components

import (
	"context"
	"fmt"
	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/kube"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("Cert manager installer", func() {
	certmanager := NewCertManager()

	It("Should install cert-manager", func() {
		if installed, err := certmanager.Installed(context.Background()); err == v1alpha1.ComponentNotInstalledByModelaError || installed {
			Skip("Test should be run on an empty cluster")
			return
		}

		Expect(certmanager.Install(context.Background(), &v1alpha1.Modela{
			ObjectMeta: v1.ObjectMeta{Name: "modela-test"},
		})).To(BeNil())

		By("Checking if it was installed")
		installed, err := certmanager.Installed(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeTrue())

		changeModelaOperatorLabel(false, "cert-manager", "cert-manager")
		installed, err = certmanager.Installed(context.Background())
		Expect(err).To(Equal(v1alpha1.ComponentNotInstalledByModelaError))

		By("Uninstalling cert-manager")
		Expect(certmanager.Uninstall(context.Background(), &v1alpha1.Modela{})).To(BeNil())
		_, err = kube.IsDeploymentCreatedByModela("cert-manager", "cert-manager")
		Expect(k8serr.IsNotFound(err)).To(BeTrue())

		By("Checking if it was uninstalled")
		installed, err = certmanager.Installed(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeFalse())
	})

	It("Should uninstall cert-manager", func() {
		fmt.Println(certmanager.Installed(context.Background()))
		By("Uninstalling cert-manager")
		Expect(certmanager.Uninstall(context.Background(), &v1alpha1.Modela{})).To(BeNil())
		_, err := kube.IsDeploymentCreatedByModela("cert-manager", "cert-manager")
		Expect(k8serr.IsNotFound(err)).To(BeTrue())
	})
})

func changeModelaOperatorLabel(add bool, ns string, name string) {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	deployment, err := clientSet.AppsV1().Deployments(ns).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return
	}
	if add {
		deployment.SetLabels(map[string]string{"app.kubernetes.io/created-by": "modela-operator"})
	} else {
		deployment.SetLabels(map[string]string{"app.kubernetes.io/created-by": ""})
	}

	_, err = clientSet.AppsV1().Deployments(ns).Update(context.Background(), deployment, v1.UpdateOptions{})
	return
}
