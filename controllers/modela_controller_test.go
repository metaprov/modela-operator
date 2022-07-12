package controllers

import (
	"context"
	"errors"
	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/controllers/components"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	ModelaName        = "modela"
	ModelaNamespace   = "default"
	TimeoutInterval   = 15 * time.Second
	PollInterval      = 500 * time.Millisecond
	NotInstalledError = errors.New("not installed")
)

var _ = Context("Inside the default namespace", func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	ctx := context.Background()

	testModelaResource := &v1alpha1.Modela{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ModelaName,
			Namespace: ModelaNamespace,
		},
		Spec: v1alpha1.ModelaSpec{
			Distribution:   "develop",
			Observability:  v1alpha1.ObservabilitySpec{},
			Ingress:        v1alpha1.ModelaAccessSpec{},
			License:        v1alpha1.ModelaLicenseSpec{},
			Tenants:        nil,
			CertManager:    v1alpha1.CertManagerSpec{},
			ObjectStore:    v1alpha1.ObjectStorageSpec{},
			SystemDatabase: v1alpha1.SystemDatabaseSpec{},
			ControlPlane:   v1alpha1.ControlPlaneSpec{},
			DataPlane:      v1alpha1.DataPlaneSpec{},
			ApiGateway:     v1alpha1.ApiGatewaySpec{},
		},
	}

	Describe("Modela Operator Controller", func() {
		Context("Modela CRD", func() {
			It("Should create the Modela CR", func() {
				createModelaResource(testModelaResource)

				By("Deleting the created Modela CR")
				Eventually(
					deleteResourceFunc(ctx, client.ObjectKey{Name: ModelaName, Namespace: ModelaNamespace}, testModelaResource),
					time.Second*3, PollInterval).Should(BeNil())

				// Uninstall database, as it should have started installing it
				_ = components.NewDatabase("").Uninstall(ctx, testModelaResource)
			})
		})
		Context("After creation", func() {
			It("Should install the enabled Helm Charts", func() {
				testModelaResource.Spec.CertManager.Install = true
				testModelaResource.Spec.ObjectStore.Install = true
				createModelaResource(testModelaResource)

				By("Installing cert-manager and changing the status")
				certManagerController := components.NewCertManager("")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingCertManager))

				By("Checking if cert-manager was installed")
				Eventually(getComponentInstalled(ctx, certManagerController), time.Minute*3, PollInterval).Should(BeNil())

				By("Installing minio and changing the status")
				minioController := components.NewObjectStorage("")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingObjectStorage))

				By("Checking if minio was installed")
				Eventually(getComponentInstalled(ctx, minioController), time.Minute*3, PollInterval).Should(BeNil())

			})
			It("Should install the system database", func() {
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingDatabase))

			})
			It("Should install the Modela system", func() {
				modelaSystemController := components.NewModelaSystem("")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingModela))
				Eventually(getComponentInstalled(ctx, modelaSystemController), time.Minute*3, PollInterval).Should(BeNil())
				Eventually(getComponentReady(ctx, modelaSystemController), time.Minute*3, PollInterval).Should(BeNil())
			})
			It("Should install the Modela catalog", func() {
				modelaSystemController := components.NewModelaSystem("")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingModela))
				Eventually(func() error {
					ready, err := modelaSystemController.CatalogInstalled(ctx)
					if err != nil {
						return err
					} else if !ready {
						return NotInstalledError
					}
					return nil
				}, time.Minute*3, PollInterval).Should(BeNil())
			})
		})
		When("Changing the spec", func() {
			It("Should not uninstall components not installed by Modela", func() {
				By("Removing the Modela Operator labels")
				certManagerController := components.NewCertManager("")
				changeModelaOperatorLabel(false, certManagerController.Namespace, "cert-manager")
				Expect(getComponentInstalled(ctx, certManagerController)).Should(Equal(v1alpha1.ComponentNotInstalledByModelaError))

				By("Disabling the component in the spec")

			})
			It("Should uninstall components after changing the spec", func() {
				By("Adding Modela Operator labels to component namespaces")

				By("Disabling Helm Chart components in the spec")

				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseUninstalling))

				By("Checking if cert-manager is installed")

				By("Checking if minio is installed")

				By("Checking if loki is installed")

				By("Checking if prometheus is installed")

				By("Checking if grafana is installed")

			})
			It("Should change the container tags of Modela pods when changing the distribution spec", func() {

			})
			It("Should install tenants added to the spec", func() {

			})
			It("Should uninstall the tenant when removed from the spec", func() {

			})
			It("Should uninstall tenants belonging to the CR before removal", func() {

			})
			It("Should modify the number of replicas when changing the deployment specs", func() {

			})
		})
		When("Removing resources belonging to the Modela Operator", func() {
			It("Should re-install the missing resources from the modela-system namespace", func() {

			})
			It("Should re-install the missing resources from the modela-catalog namespace", func() {

			})
		})
	})
})

func createObject(obj client.Object) error {
	err := k8sClient.Create(context.Background(), obj)
	if k8serr.IsAlreadyExists(err) {
		err = nil
	}

	return err
}

func getResourceFunc(ctx context.Context, key client.ObjectKey, obj client.Object) func() error {
	return func() error {
		return k8sClient.Get(ctx, key, obj)
	}
}

func deleteResourceFunc(ctx context.Context, key client.ObjectKey, obj client.Object) func() error {
	return func() error {
		if err := getResourceFunc(ctx, key, obj)(); err != nil {
			if k8serr.IsNotFound(err) {
				err = nil
			}

			return err
		}

		return k8sClient.Delete(ctx, obj)
	}
}

func getModelaStatus(ctx context.Context) func() v1alpha1.ModelaPhase {
	return func() v1alpha1.ModelaPhase {
		obj := &v1alpha1.Modela{}
		_ = k8sClient.Get(ctx, client.ObjectKey{Name: ModelaName, Namespace: ModelaNamespace}, obj)
		return obj.Status.Phase
	}
}

func getComponentInstalled(ctx context.Context, component ModelaComponent) func() error {
	return func() error {
		installed, err := component.Installed(ctx)
		if err != nil {
			return err
		} else if !installed {
			return NotInstalledError
		}
		return nil
	}
}

func getComponentReady(ctx context.Context, component ModelaComponent) func() error {
	return func() error {
		ready, err := component.Ready(ctx)
		if err != nil {
			return err
		} else if !ready {
			return NotInstalledError
		}
		return nil
	}
}

func changeModelaOperatorLabel(add bool, ns string, name string) {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	deployment, err := clientSet.AppsV1().Deployments(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return
	}
	if add {
		deployment.SetLabels(map[string]string{"app.kubernetes.io/created-by": "modela-operator"})
	} else {
		deployment.SetLabels(map[string]string{"app.kubernetes.io/created-by": ""})
	}

	_, err = clientSet.AppsV1().Deployments(ns).Update(context.Background(), deployment, metav1.UpdateOptions{})
	return
}

func createModelaResource(modela *v1alpha1.Modela) {
	By("Creating a new Modela resource")
	Expect(createObject(modela)).Should(Succeed())

	By("Checking if the Modela resource was created")
	Eventually(
		getResourceFunc(context.Background(), client.ObjectKey{Name: modela.Name, Namespace: modela.Namespace}, modela),
		time.Second*3, PollInterval).Should(BeNil(), "Modela resource %s", modela.Name)

}
