package components

import (
	"context"
	"github.com/metaprov/modela-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/exec"
	"time"
)

var _ = Describe("Vault installer", func() {
	It("Should install vault", func() {
		vault := NewVault()
		if installed, err := vault.Installed(context.Background()); err == v1alpha1.ComponentNotInstalledByModelaError || installed {
			Skip("Test should be run on an empty cluster")
			return
		}

		Expect(vault.Install(context.Background(), &v1alpha1.Modela{
			ObjectMeta: v1.ObjectMeta{Name: "modela-test"},
		})).To(BeNil())

		By("Checking if it was installed")
		installed, err := vault.Installed(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeTrue())
		Eventually(func() bool { t, _ := vault.Ready(context.Background()); return t }, time.Second*60, time.Second*1).Should(BeTrue())
	})

	It("Should configure vault", func() {
		// Port forward the vault server
		port_forward := exec.Command("kubectl", "port-forward", "-n", "modela-system", "svc/modela-vault", "8200:8200")
		Expect(port_forward.Start()).To(Succeed())

		vault := NewVault()
		Expect(vault.ConfigureVault(context.Background(), &v1alpha1.Modela{})).To(Succeed())

		_ = port_forward.Process.Kill()

	})
})
