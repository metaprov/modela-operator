package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/controllers/components"
	"github.com/metaprov/modela-operator/pkg/kube"
	"github.com/metaprov/modelaapi/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/networking/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	ModelaName        = "modela"
	ModelaNamespace   = "modela-system"
	TimeoutInterval   = 60 * time.Second
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
		Status: v1alpha1.ModelaStatus{},
		Spec: v1alpha1.ModelaSpec{
			Distribution: "develop",
			Observability: v1alpha1.ObservabilitySpec{
				Loki:       true,
				Prometheus: true,
				Grafana:    true,
			},
			Ingress: v1alpha1.ModelaAccessSpec{},
			License: v1alpha1.ModelaLicenseSpec{},
			Tenants: nil,
			CertManager: v1alpha1.CertManagerSpec{
				Install: true,
			},
			ObjectStore: v1alpha1.ObjectStorageSpec{
				Install: true,
			},
			SystemDatabase: v1alpha1.SystemDatabaseSpec{},
			ControlPlane:   v1alpha1.ControlPlaneSpec{},
			DataPlane:      v1alpha1.DataPlaneSpec{},
			ApiGateway:     v1alpha1.ApiGatewaySpec{},
		},
	}

	certManagerController := components.NewCertManager()
	minioController := components.NewObjectStorage()
	lokiController := components.NewObjectStorage()
	grafanaController := components.NewGrafana()
	prometheusController := components.NewObjectStorage()
	modelaSystemController := components.NewModelaSystem("develop")
	nginxController := components.NewNginx()

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
				databaseController := components.NewDatabase()

				By("Installing postgres and changing the status")
				Eventually(getModelaStatus(ctx), TimeoutInterval, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingDatabase))

				By("Checking if postgres was installed")
				Eventually(getComponentInstalled(ctx, databaseController), time.Minute*3, PollInterval).Should(BeNil())
			})
			It("Should install the Modela system", func() {
				Eventually(getModelaStatus(ctx), 2*time.Minute, PollInterval).Should(Equal(v1alpha1.ModelaPhaseInstallingModela))
				Eventually(getComponentInstalled(ctx, modelaSystemController), time.Minute*3, PollInterval).Should(BeNil())
				Eventually(getComponentReady(ctx, modelaSystemController), time.Minute*3, PollInterval).Should(BeNil())
			})
			It("Should install the Modela catalog", func() {
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
				createModelaResource(testModelaResource)
				Expect(updateObject(testModelaResource, func(object client.Object) error {
					modela := object.(*v1alpha1.Modela)
					modela.Spec.Observability.Grafana = true
					return nil
				})).To(Succeed())

				By("Checking if grafana was installed")
				Eventually(getComponentInstalled(ctx, grafanaController), time.Minute*3, PollInterval).Should(BeNil())

				Expect(updateObject(testModelaResource, func(object client.Object) error {
					modela := object.(*v1alpha1.Modela)
					modela.Spec.CertManager.Install = false
					modela.Spec.ObjectStore.Install = false
					modela.Spec.Observability.Loki = false
					modela.Spec.Observability.Grafana = false
					modela.Spec.Observability.Prometheus = false
					return nil
				})).To(Succeed())

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
				testModelaResource.Spec.CertManager.Install = true
				testModelaResource.Spec.ObjectStore.Install = true
				testModelaResource.Spec.Observability.Loki = true
				testModelaResource.Spec.Observability.Grafana = true
				testModelaResource.Spec.Observability.Prometheus = true
				testModelaResource.Status.InstalledVersion = "develop"
				//createModelaResource(testModelaResource)

				Eventually(getComponentReady(ctx, modelaSystemController), time.Minute*3, PollInterval).Should(BeNil())
				Eventually(getModelaVersion(ctx), TimeoutInterval, PollInterval).Should(Equal("develop"))

				Expect(expectDeploymentTagVersion("modela-system", "modela-control-plane", "develop")()).To(BeTrue())

				By("Changing the distribution and updating the resource")
				Expect(updateObject(testModelaResource, func(object client.Object) error {
					modela := object.(*v1alpha1.Modela)
					modela.Spec.Distribution = "stable"
					return nil
				})).To(Succeed())

				Eventually(expectDeploymentTagVersion("modela-system", "modela-control-plane", "stable"),
					time.Minute*3, PollInterval).Should(BeTrue())
			})
			It("Should install tenants added to the spec", func() {
				createModelaResource(testModelaResource)

				Eventually(getComponentReady(ctx, modelaSystemController), time.Minute*3, PollInterval).Should(BeNil())

				By("Adding a tenant and updating the resource")
				Expect(updateObject(testModelaResource, func(object client.Object) error {
					modela := object.(*v1alpha1.Modela)
					modela.Spec.Tenants = []*v1alpha1.TenantSpec{{
						Name:          "default-tenant",
						AdminPassword: util.StrPtr("test123"),
					}}
					return nil
				})).To(Succeed())

				tenantController := components.NewTenant("default-tenant")
				Eventually(func() error {
					ready, err := tenantController.Ready(context.Background())
					fmt.Println(ready, err)
					if err != nil {
						return err
					} else if !ready {
						return NotInstalledError
					}
					return nil
				}, time.Minute*3, PollInterval).Should(BeNil())
			})
			It("Should uninstall the tenant when removed from the spec", func() {
				createModelaResource(testModelaResource)

				Eventually(getComponentReady(ctx, modelaSystemController), time.Minute*3, PollInterval).Should(BeNil())
				tenantController := components.NewTenant("default-tenant")
				if installed, _ := tenantController.Installed(context.Background()); !installed {
					By("Adding the tenant and updating the resource")
					Expect(updateObject(testModelaResource, func(object client.Object) error {
						modela := object.(*v1alpha1.Modela)
						modela.Spec.Tenants = []*v1alpha1.TenantSpec{{
							Name:          "default-tenant",
							AdminPassword: util.StrPtr("test123"),
						}}
						return nil
					})).To(Succeed())
					Eventually(func() error {
						ready, err := tenantController.Ready(context.Background())
						if err != nil {
							return err
						} else if !ready {
							return NotInstalledError
						}
						return nil
					}, time.Minute*3, PollInterval).Should(BeNil())
				}

				By("Removing tenants and updating the resource")
				Expect(updateObject(testModelaResource, func(object client.Object) error {
					modela := object.(*v1alpha1.Modela)
					modela.Spec.Tenants = []*v1alpha1.TenantSpec{}
					return nil
				})).To(Succeed())

				Eventually(func() error {
					installed, err := tenantController.Installed(context.Background())
					if err != nil {
						return err
					} else if !installed {
						return NotInstalledError
					}
					return nil
				}, time.Minute*3, PollInterval).ShouldNot(BeNil())
			})
			It("Should install ingress", func() {
				createModelaResource(testModelaResource)

				Eventually(getComponentReady(ctx, modelaSystemController), time.Minute*3, PollInterval).Should(BeNil())

				By("Enabling ingress updating the resource")
				Expect(updateObject(testModelaResource, func(object client.Object) error {
					modela := object.(*v1alpha1.Modela)
					modela.Spec.Ingress.Hostname = util.StrPtr("localhost")
					modela.Spec.Ingress.Enabled = true
					modela.Spec.Ingress.InstallNginx = true
					modela.SetAnnotations(map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					})
					return nil
				})).To(Succeed())

				By("Checking if nginx was installed")
				Eventually(getComponentInstalled(ctx, nginxController), time.Minute*3, PollInterval).Should(BeNil())

				var ingress v1.Ingress
				By("Checking if the Ingress resource was created")
				Eventually(
					getResourceFunc(context.Background(), client.ObjectKey{Name: "modela-frontend-ingress", Namespace: "modela-system"}, &ingress),
					time.Minute*3, PollInterval).Should(BeNil())

			})
			It("Should modify the number of replicas when changing the deployment specs", func() {

			})
		})
		When("Removing resources belonging to the Modela Operator", func() {
			It("Should re-install the missing resources from the modela-system namespace", func() {
				createModelaResource(testModelaResource)

				By("Deleting the modela control plane")
				var controlPlane appsv1.Deployment
				Eventually(
					deleteResourceFunc(context.Background(), client.ObjectKey{Name: "modela-control-plane", Namespace: ModelaNamespace}, &controlPlane),
					time.Second*3, PollInterval).Should(BeNil())

				time.Sleep(1 * time.Second)

				By("Expecting it to be re-created")
				Eventually(
					getResourceFunc(context.Background(), client.ObjectKey{Name: "modela-control-plane", Namespace: ModelaNamespace}, &controlPlane),
					TimeoutInterval, PollInterval).Should(BeNil())
			})
		})
	})
})

func createObject(obj client.Object) error {
	err := k8sClient.Create(context.Background(), obj)
	obj.SetResourceVersion("")
	if k8serr.IsAlreadyExists(err) {
		err = nil
	}

	return err
}

func updateObject(obj client.Object, mutate func(client.Object) error) error {
	key := client.ObjectKeyFromObject(obj)
	if err := k8sClient.Get(context.Background(), key, obj); err != nil {
		return err
	}

	if err := mutate(obj); err != nil {
		return err
	}

	return k8sClient.Update(context.Background(), obj)
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

func getModelaVersion(ctx context.Context) func() string {
	return func() string {
		obj := &v1alpha1.Modela{}
		_ = k8sClient.Get(ctx, client.ObjectKey{Name: ModelaName, Namespace: ModelaNamespace}, obj)
		return string(obj.Status.InstalledVersion)
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

func expectDeploymentTagVersion(ns, name, version string) func() bool {
	return func() bool {
		var deployment appsv1.Deployment
		err := k8sClient.Get(context.Background(), client.ObjectKey{Name: name, Namespace: ns}, &deployment)
		if err != nil {
			fmt.Printf("Error fetching deployment: %v (name=%s, ns=%s)\n", err, name, ns)
			return false
		}
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if strings.Split(container.Image, ":")[1] != version {
				return false
			}
		}
		return true
	}
}

func createModelaResource(modela *v1alpha1.Modela) {
	_ = kube.CreateNamespace("modela-system", "modela")
	By("Creating a new Modela resource")
	Expect(createObject(modela)).Should(Succeed())

	By("Checking if the Modela resource was created")
	Eventually(
		getResourceFunc(context.Background(), client.ObjectKey{Name: modela.Name, Namespace: modela.Namespace}, modela),
		time.Second*3, PollInterval).Should(BeNil(), "Modela resource %s", modela.Name)

}
