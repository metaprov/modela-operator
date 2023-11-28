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

var _ = Describe("Prometheus installer", func() {
	prometheus := NewPrometheus()

	It("Should install Prometheus", func() {
		if installed, err := prometheus.Installed(context.Background()); err == v1alpha1.ComponentNotInstalledByModelaError || installed {
			Skip("Test should be run on an empty cluster")
			return
		}

		Expect(prometheus.Install(context.Background(), &v1alpha1.Modela{
			ObjectMeta: v1.ObjectMeta{Name: "modela-test"},
		})).To(BeNil())

		By("Checking if it was installed")
		installed, err := prometheus.Installed(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeTrue())

		changeDeploymentModelaOperatorLabel(false, "prometheus-community", "kube-prometheus-server")
		installed, err = prometheus.Installed(context.Background())
		Expect(err).To(Equal(v1alpha1.ComponentNotInstalledByModelaError))

		By("Uninstalling prometheus")
		Expect(prometheus.Uninstall(context.Background(), &v1alpha1.Modela{})).To(BeNil())
		_, err = kube.IsDeploymentCreatedByModela("prometheus-community", "kube-prometheus-server")
		Expect(k8serr.IsNotFound(err)).To(BeTrue())

		By("Checking if it was uninstalled")
		installed, err = prometheus.Installed(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeFalse())
	})
})
