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
	TimeoutInterval   = 30 * time.Second
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

	certManagerController := components.NewCertManager("")
	minioController := components.NewObjectStorage("")
	lokiController := components.NewObjectStorage("")
	grafanaController := components.NewGrafana("")
	prometheusController := components.NewObjectStorage("")

	Describe("Modela Operator Controller", func() {
		Context("Modela CRD", func() {
			It("Should create the Modela CR", func() {
				createModelaResource(testModelaResource)

				By("Deleting the created Modela CR")
				Eventually(
					deleteResourceFunc(ctx, client.ObjectKey{Name: ModelaName, Namespace: ModelaNamespace}, testModelaResource),
					time.Second*3, PollInterval).Should(BeNil())

				// Uninstall database, as it should have started installing it
				//_ = components.NewDatabase("").Uninstall(ctx, testModelaResource)
			})
		})
		Context("After creation", func() {

			/*After(func() {
				Eventually(
					deleteResourceFunc(ctx, client.ObjectKey{Name: ModelaName, Namespace: ModelaNamespace}, testModelaResource),
					time.Second*3, PollInterval).Should(BeNil())
			})*/

			It("Should install the enabled Helm Charts", func() {
				testModelaResource.Spec.CertManager.Install = true
				testModelaResource.Spec.ObjectStore.Install = true
				testModelaResource.Spec.Observability.Loki = true
				testModelaResource.Spec.Observability.Grafana = true
				testModelaResource.Spec.Observability.Prometheus = true
				createModelaResource(testModelaResource)

				By("Installing cert-manager and changing the status")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingCertManager))

				By("Checking if cert-manager was installed")
				Eventually(getComponentInstalled(ctx, certManagerController), time.Minute*3, PollInterval).Should(BeNil())

				By("Installing minio and changing the status")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingObjectStorage))

				By("Checking if minio was installed")
				Eventually(getComponentInstalled(ctx, minioController), time.Minute*3, PollInterval).Should(BeNil())

				By("Installing loki and changing the status")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingLoki))

				By("Checking if loki was installed")
				Eventually(getComponentInstalled(ctx, lokiController), time.Minute*3, PollInterval).Should(BeNil())

				By("Installing grafana and changing the status")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingGrafana))

				By("Checking if grafana was installed")
				Eventually(getComponentInstalled(ctx, grafanaController), time.Minute*3, PollInterval).Should(BeNil())

				By("Installing prometheus and changing the status")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingPrometheus))

				By("Checking if prometheus was installed")
				Eventually(getComponentInstalled(ctx, prometheusController), time.Minute*3, PollInterval).Should(BeNil())
			})
			It("Should install the system database", func() {
				databaseController := components.NewDatabase("")

				By("Installing postgres and changing the status")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingDatabase))

				By("Checking if postgres was installed")
				Eventually(getComponentInstalled(ctx, databaseController), time.Minute*3, PollInterval).Should(BeNil())
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
			It("Should uninstall components after changing the spec", func() {
				testModelaResource.Spec.CertManager.Install = false
				testModelaResource.Spec.ObjectStore.Install = false
				testModelaResource.Spec.Observability.Loki = false
				testModelaResource.Spec.Observability.Grafana = false
				testModelaResource.Spec.Observability.Prometheus = false
				createModelaResource(testModelaResource)

				By("Changing the status to uninstalling")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseUninstalling))

				By("Checking if cert-manager is installed")
				Eventually(getComponentInstalled(ctx, certManagerController), time.Minute*3, PollInterval).Should(Equal(NotInstalledError))

				By("Checking if minio is installed")
				Eventually(getComponentInstalled(ctx, minioController), time.Minute*3, PollInterval).Should(Equal(NotInstalledError))

				By("Checking if loki is installed")
				Eventually(getComponentInstalled(ctx, lokiController), time.Minute*3, PollInterval).Should(Equal(NotInstalledError))

				By("Checking if prometheus is installed")
				Eventually(getComponentInstalled(ctx, prometheusController), time.Minute*3, PollInterval).Should(Equal(NotInstalledError))

				By("Checking if grafana is installed")
				Eventually(getComponentInstalled(ctx, grafanaController), time.Minute*3, PollInterval).Should(Equal(NotInstalledError))

				By("Should return to a ready state")
				Eventually(getModelaStatus(ctx), time.Minute*3, PollInterval).Should(Equal(v1alpha1.ModelaPhaseReady))
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
		if err := k8sClient.Delete(context.Background(), obj); err != nil {
			return err
		} else {
			return createObject(obj)
		}
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

func getModelaStatus(ctx context.Context) func() string {
	return func() string {
		obj := &v1alpha1.Modela{}
		_ = k8sClient.Get(ctx, client.ObjectKey{Name: ModelaName, Namespace: ModelaNamespace}, obj)
		return string(obj.Status.Phase)
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
