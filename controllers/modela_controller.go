/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/metaprov/modela-operator/controllers/common"
	"github.com/metaprov/modela-operator/controllers/components"
	"github.com/metaprov/modelaapi/pkg/util"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	managementv1alpha1 "github.com/metaprov/modela-operator/api/v1alpha1"
)

var (
	componentNotInstalled = errors.New("component not installed")
)

const (
	defaultApiUrl  = "http://localhost:8080"
	defaultDataUrl = "http://localhost:8095/upload"
)

// ModelaReconciler reconciles a Modela object
type ModelaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas/finalizers,verbs=update
//+kubebuilder:rbac:groups=catalog.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=team.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=data.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inference.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infra.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=training.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=*,verbs=*
//+kubebuilder:rbac:groups="extensions",resources=*,verbs=*
//+kubebuilder:rbac:groups="apps",resources=*,verbs=*
//+kubebuilder:rbac:groups="core",resources=*,verbs=*
//+kubebuilder:rbac:groups="batch",resources=*,verbs=*
//+kubebuilder:rbac:groups=cert-manager.io,resources=*,verbs=*
//+kubebuilder:rbac:groups=issuers.cert-manager.io,resources=*,verbs=*
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=*,verbs=get;list;watch;create;update;delete;patch
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=*,verbs=get;list;watch;create;update;delete;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;delete;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingressclasses,verbs=get;list;watch;create;update;delete;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=services,verbs=get;list;watch;create;update;delete;patch
//+kubebuilder:rbac:groups=policy,resources=podsecuritypolicies,verbs=get;list;watch;create;update;delete;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *ModelaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Requeue initiated")

	var modela = new(managementv1alpha1.Modela)
	if err := r.Get(ctx, req.NamespacedName, modela); err != nil {
		logger.Error(err, "unable to fetch Modela")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	oldStatus := *modela.Status.DeepCopy()

	result, err := r.Install(ctx, modela)
	if err != nil {
		modela.Status.FailureMessage = util.StrPtr(err.Error())
		modela.Status.Phase = managementv1alpha1.ModelaPhaseFailed
		logger.Error(err, "failed to install Modela")
		result = ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 10,
		}
		goto updateStatus
	} else {
		modela.Status.FailureMessage = nil
	}

	if result.Requeue || !r.isStateEqual(modela.Status, oldStatus) {
		goto updateStatus
	}

	result, err = r.reconcileIngress(ctx, modela)
	if err != nil || result.Requeue || !r.isStateEqual(modela.Status, oldStatus) {
		goto updateStatus
	}

	result, err = r.reconcileControlPlane(ctx, modela)
	if err != nil || result.Requeue || !r.isStateEqual(modela.Status, oldStatus) {
		goto updateStatus
	}

	result, err = r.reconcileDataPlane(ctx, modela)
	if err != nil || result.Requeue || !r.isStateEqual(modela.Status, oldStatus) {
		goto updateStatus
	}

	result, err = r.reconcileApiGateway(ctx, modela)
	if err != nil || result.Requeue || !r.isStateEqual(modela.Status, oldStatus) {
		goto updateStatus
	}

updateStatus:
	statusResult, statusErr := r.updateStatus(ctx, oldStatus, *modela)
	if statusResult.Requeue {
		return statusResult, statusErr
	}

	return result, err
}

func (r ModelaReconciler) updateStatus(ctx context.Context, oldStatus managementv1alpha1.ModelaStatus, modela managementv1alpha1.Modela) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	if !r.isStateEqual(modela.Status, oldStatus) {
		now := metav1.Now()
		modela.Status.LastUpdated = &now

		if err := r.Status().Update(ctx, &modela); err != nil {
			if k8serr.IsConflict(err) || k8serr.IsNotFound(err) {
				// Modela has been updated since we read it.
				// Requeue Modela to try to reconciliate again.
				return ctrl.Result{Requeue: true}, nil
			}
			logger.Error(err, "unable to update Modela status")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r ModelaReconciler) updatePhase(ctx context.Context, modela *managementv1alpha1.Modela, phase managementv1alpha1.ModelaPhase) (ctrl.Result, error) {
	if modela.Status.Phase == phase {
		return ctrl.Result{}, nil
	}
	now := metav1.Now()
	modela.Status.LastUpdated = &now
	modela.Status.Phase = phase
	if err := r.Status().Update(ctx, modela); err != nil {
		log.FromContext(ctx).Error(err, "failed to update status")
		return ctrl.Result{Requeue: true}, err
	}
	log.FromContext(context.Background()).Info("New phase", "phase", phase)
	return ctrl.Result{}, nil
}

func (r ModelaReconciler) isStateEqual(old managementv1alpha1.ModelaStatus, new managementv1alpha1.ModelaStatus) bool {
	return old.InstalledVersion == new.InstalledVersion &&
		old.Phase == new.Phase &&
		reflect.DeepEqual(old.FailureMessage, new.FailureMessage) &&
		reflect.DeepEqual(old.LicenseToken, new.LicenseToken) &&
		reflect.DeepEqual(old.Conditions, new.Conditions) &&
		reflect.DeepEqual(old.Tenants, new.Tenants)

}

func (r *ModelaReconciler) Install(ctx context.Context, modela *managementv1alpha1.Modela) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var wg sync.WaitGroup
	var componentsInstalled sync.Map
	var componentList = []ModelaComponent{
		components.NewCertManager(),
		components.NewObjectStorage(),
		components.NewLoki(),
		components.NewGrafana(),
		components.NewPrometheus(),
		components.NewPostgresDatabase(),
		components.NewMongoDatabase(),
		components.NewNginx(),
		components.NewOnlineStore(),
		components.NewVault(),
	}

	for _, component := range componentList {
		wg.Add(1)

		go func(component ModelaComponent) {
			defer wg.Done()
			installed, err := component.Installed(ctx)
			if err != nil && err != managementv1alpha1.ComponentNotInstalledByModelaError && err != managementv1alpha1.ComponentMissingResourcesError {
				logger.Error(err, "Failed to check if component is installed", "component", component.GetInstallPhase())
			}

			if !installed {
				componentsInstalled.Store(component, componentNotInstalled)
			} else {
				componentsInstalled.Store(component, err)
			}
		}(component)
	}

	wg.Wait()
	for _, component := range componentList {
		installed, _ := componentsInstalled.Load(component)
		if installed == managementv1alpha1.ComponentNotInstalledByModelaError {
			continue
		}

		if !component.IsEnabled(*modela) && installed != componentNotInstalled {
			if result, err := r.reconcileComponent(ctx, component, true, modela); err != nil || result.Requeue {
				return result, err
			}
		} else if component.IsEnabled(*modela) && installed != nil {
			if result, err := r.reconcileComponent(ctx, component, false, modela); err != nil || result.Requeue {
				return result, err
			}
		}
	}

	vault := components.NewVault()
	if err := vault.ConfigureVault(ctx, modela); err != nil {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, err
	}

	modelaSystem := components.NewModelaSystem(modela.Spec.Distribution)
	if installed, err := modelaSystem.Installed(ctx); !installed {
		if err := modelaSystem.Install(ctx, modela); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, err
		}
	} else if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
	}

	if ready, err := modelaSystem.Ready(ctx); err != nil || !ready {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, err
	}

	if modela.Status.InstalledVersion == "" {
		modela.Status.InstalledVersion = modela.Spec.Distribution
		return ctrl.Result{Requeue: true}, nil
	}

	if installed, err := modelaSystem.CatalogInstalled(ctx); !installed || err == managementv1alpha1.ComponentMissingResourcesError {
		if result, _ := r.updatePhase(ctx, modela, managementv1alpha1.ModelaPhaseInstallingModela); result.Requeue {
			return result, nil
		}
		err := modelaSystem.InstallCatalog(ctx, modela)
		if err != nil {
			logger.Error(err, "Failed to install modela catalog")
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 10 * time.Second,
			}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 10 * time.Second,
		}, err
	}

	result, err := r.reconcileTenants(ctx, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	if modela.Spec.Distribution != modela.Status.InstalledVersion {
		var err error
		logger.Info("Applying new distribution", "version", modelaSystem.ModelaVersion)
		err = modelaSystem.InstallNewVersion(ctx, modela)
		if err != nil && modela.Spec.OnlineStore.Install {
			onlineStore := components.NewOnlineStore()
			err = onlineStore.InstallNewVersion(ctx, modela)
		}

		if err != nil {
			return ctrl.Result{Requeue: true}, err
		} else {
			modela.Status.InstalledVersion = modela.Spec.Distribution
		}
	}

	modela.Status.Phase = managementv1alpha1.ModelaPhaseReady
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).Named("modela-controller").
		For(&managementv1alpha1.Modela{}).
		Owns(&appsv1.Deployment{}).
		Owns(&v1.Service{}).
		Owns(&v1.ServiceAccount{}).
		Owns(&v1.Secret{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}

func (r *ModelaReconciler) updateFrontendConfig(configMap v1.ConfigMap) error {
	var frontendDeployment appsv1.Deployment

	if err := r.Update(context.TODO(), &configMap); err != nil {
		return err
	}

	frontendIdentifier := types.NamespacedName{
		Name:      "modela-frontend",
		Namespace: "modela-system",
	}

	if err := r.Get(context.TODO(), frontendIdentifier, &frontendDeployment); err == nil {
		frontendDeployment.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().String()
		_ = r.Update(context.TODO(), &frontendDeployment)
	}

	return nil
}

func (r *ModelaReconciler) reconcileIngress(ctx context.Context, modela *managementv1.Modela) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	frontendConfigMap := v1.ConfigMap{}
	configMapIdentifier := types.NamespacedName{
		Name:      "frontend-config",
		Namespace: "modela-system",
	}
	if err := r.Get(ctx, configMapIdentifier, &frontendConfigMap); err != nil {
		logger.Error(err, "error fetching frontend config")
		return ctrl.Result{Requeue: true}, nil
	}

	apiUrl, _ := frontendConfigMap.Data["apiUrl"]
	dataUrl, _ := frontendConfigMap.Data["dataUrl"]

	if modela.Spec.Network.Ingress == nil || !modela.Spec.Network.Ingress.Enabled {
		return ctrl.Result{}, nil
	}

	var hostname string
	if modela.Spec.Network.Ingress.Hostname == nil {
		hostname = "localhost"
	} else {
		hostname = *modela.Spec.Network.Ingress.Hostname
	}

	desiredApiUrl := fmt.Sprintf("http://modela-api.%s", hostname)
	desiredDataUrl := fmt.Sprintf("http://modela-api.%s/upload", hostname)

	if apiUrl != desiredApiUrl || dataUrl != desiredDataUrl {
		frontendConfigMap.Data["apiUrl"] = desiredApiUrl
		frontendConfigMap.Data["dataUrl"] = desiredDataUrl
		if err := r.updateFrontendConfig(frontendConfigMap); err != nil {
			logger.Error(err, "error updating frontend config")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	frontendIngress, err := common.BuildFrontendIngress(hostname, *modela)
	if err != nil {
		logger.Error(err, "unable to generate ingress")
		return ctrl.Result{}, err
	}

	var liveIngress networkingv1.Ingress
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: frontendIngress.GetNamespace(),
		Name:      frontendIngress.GetName(),
	}, &liveIngress); err != nil {
		if err := r.createIngress(frontendIngress, modela); err != nil {
			logger.Error(err, "failed to create ingress")
			return ctrl.Result{Requeue: true}, nil
		}
	} else {
		if liveIngress.Spec.Rules[0].Host != frontendIngress.Spec.Rules[0].Host ||
			liveIngress.Spec.Rules[1].Host != frontendIngress.Spec.Rules[1].Host ||
			!reflect.DeepEqual(liveIngress.Annotations, frontendIngress.Annotations) {
			liveIngress.Spec.Rules[0].Host = frontendIngress.Spec.Rules[0].Host
			liveIngress.Spec.Rules[1].Host = frontendIngress.Spec.Rules[1].Host
			liveIngress.Annotations = frontendIngress.Annotations
			if err := r.Update(ctx, &liveIngress); err != nil {
				logger.Error(err, "unable to update ingress")
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModelaReconciler) createIngress(ingress *networkingv1.Ingress, modela *managementv1alpha1.Modela) error {
	if err := controllerutil.SetControllerReference(modela, ingress, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(context.TODO(), ingress); err != nil {
		if k8serr.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}

func (r *ModelaReconciler) reconcileTenants(ctx context.Context, modela *managementv1.Modela) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var installedTenants = make(map[string]bool)
	for _, tenant := range modela.Status.Tenants {
		installedTenants[tenant] = true
	}

	var tenants = make(map[string]bool)
	for _, tenantSpec := range modela.Spec.Tenants {
		tenants[tenantSpec.Name] = true
		tenant := components.NewTenant(tenantSpec.Name)
		if _, installed := installedTenants[tenantSpec.Name]; installed {
			continue
		} else if !installed {
			if result, _ := r.updatePhase(ctx, modela, managementv1alpha1.ModelaPhaseInstallingTenant); result.Requeue {
				return result, nil
			}
			if err := tenant.Install(ctx, modela, tenantSpec); err != nil {
				logger.Error(err, "Failed to install tenant", "name", tenant.Name)
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 5 * time.Second,
				}, err
			}
			modela.Status.Tenants = append(modela.Status.Tenants, tenant.Name)
		}
	}

	// Uninstall inactive tenants
	for index, tenant := range modela.Status.Tenants {
		if _, ok := tenants[tenant]; !ok {
			// The tenant no longer exists in the spec, uninstall
			tenant := components.NewTenant(tenant)
			if result, _ := r.updatePhase(ctx, modela, managementv1alpha1.ModelaPhaseUninstalling); result.Requeue {
				return result, nil
			}
			err := tenant.Uninstall(ctx, modela)
			if err != nil {
				logger.Error(err, "Failed to uninstall tenant", "name", tenant.Name)
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 5 * time.Minute,
				}, err
			}
			// Remove the tenant from the status
			modela.Status.Tenants = append(modela.Status.Tenants[:index], modela.Status.Tenants[index+1:]...)
		}
	}

	return ctrl.Result{}, nil
}

// ModelaComponent defines the interface for system components that can be reconciled
type ModelaComponent interface {
	IsEnabled(modela managementv1.Modela) bool
	Installed(ctx context.Context) (bool, error)
	Install(ctx context.Context, modela *managementv1.Modela) error
	Installing(ctx context.Context) (bool, error)
	Ready(ctx context.Context) (bool, error)
	Uninstall(ctx context.Context, modela *managementv1.Modela) error
	GetInstallPhase() managementv1alpha1.ModelaPhase
}

func (r *ModelaReconciler) reconcileComponent(ctx context.Context, component ModelaComponent, installed bool, modela *managementv1.Modela) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if !component.IsEnabled(*modela) && installed {
		if result, _ := r.updatePhase(ctx, modela, managementv1alpha1.ModelaPhaseUninstalling); result.Requeue {
			return result, nil
		}
		err := component.Uninstall(ctx, modela)
		if err != nil {
			logger.Error(err, "Failed to uninstall component", "component", reflect.TypeOf(component).Name())
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, nil
	}

	installing, err := component.Installing(ctx)
	if err != nil && err != managementv1alpha1.ComponentMissingResourcesError {
		logger.Error(err, "Failed to check if component is installing", "component", reflect.TypeOf(component).Name())
		return ctrl.Result{Requeue: true}, err
	}

	if installing && err != managementv1alpha1.ComponentMissingResourcesError {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 10 * time.Second,
		}, nil
	} else {
		if result, _ := r.updatePhase(ctx, modela, component.GetInstallPhase()); result.Requeue {
			return result, nil
		}
		if err := component.Install(ctx, modela); err != nil {
			logger.Error(err, "Failed to install component", "component", reflect.TypeOf(component).Name())
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Minute,
			}, err
		}
		return ctrl.Result{}, nil
	}
}

func (r *ModelaReconciler) reconcileApiGateway(ctx context.Context, modela *managementv1alpha1.Modela) (ctrl.Result, error) {
	if modela.Spec.ApiGateway.Replicas == nil && modela.Spec.ApiGateway.Resources == nil {
		return ctrl.Result{}, nil
	}
	var deployment appsv1.Deployment

	name := types.NamespacedName{
		Namespace: "modela-system",
		Name:      "modela-api-gateway",
	}

	if err := r.Get(ctx, name, &deployment); err != nil {
		return ctrl.Result{}, err
	}

	var updateDeployment bool
	if modela.Spec.ApiGateway.Replicas != nil {
		if *deployment.Spec.Replicas != *modela.Spec.ApiGateway.Replicas && *modela.Spec.ApiGateway.Replicas > 0 {
			deployment.Spec.Replicas = modela.Spec.ApiGateway.Replicas
			updateDeployment = true
		}
	}

	if modela.Spec.ApiGateway.Resources != nil {
		if !reflect.DeepEqual(*modela.Spec.ApiGateway.Resources, deployment.Spec.Template.Spec.Containers[0].Resources) {
			deployment.Spec.Template.Spec.Containers[0].Resources = *modela.Spec.ApiGateway.Resources
			updateDeployment = true
		}
	}

	if updateDeployment {
		if err := r.Update(ctx, &deployment); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModelaReconciler) reconcileControlPlane(ctx context.Context, modela *managementv1alpha1.Modela) (ctrl.Result, error) {
	if modela.Spec.ControlPlane.Replicas == nil && modela.Spec.ControlPlane.Resources == nil {
		return ctrl.Result{}, nil
	}
	var deployment appsv1.Deployment

	name := types.NamespacedName{
		Namespace: "modela-system",
		Name:      "modela-control-plane",
	}

	if err := r.Get(ctx, name, &deployment); err != nil {
		return ctrl.Result{}, err
	}

	var updateDeployment bool
	if modela.Spec.ControlPlane.Replicas != nil {
		if *deployment.Spec.Replicas != *modela.Spec.ControlPlane.Replicas && *modela.Spec.ControlPlane.Replicas > 0 {
			deployment.Spec.Replicas = modela.Spec.ControlPlane.Replicas
			updateDeployment = true
		}
	}

	if modela.Spec.ControlPlane.Resources != nil {
		if !reflect.DeepEqual(*modela.Spec.ControlPlane.Resources, deployment.Spec.Template.Spec.Containers[0].Resources) {
			deployment.Spec.Template.Spec.Containers[0].Resources = *modela.Spec.ControlPlane.Resources
			updateDeployment = true
		}
	}

	if updateDeployment {
		if err := r.Update(ctx, &deployment); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModelaReconciler) reconcileDataPlane(ctx context.Context, modela *managementv1alpha1.Modela) (ctrl.Result, error) {
	if modela.Spec.DataPlane.Replicas == nil && modela.Spec.DataPlane.Resources == nil {
		return ctrl.Result{}, nil
	}
	var deployment appsv1.Deployment

	name := types.NamespacedName{
		Namespace: "modela-system",
		Name:      "modela-data-plane",
	}

	if err := r.Get(ctx, name, &deployment); err != nil {
		return ctrl.Result{}, err
	}

	var updateDeployment bool
	if modela.Spec.DataPlane.Replicas != nil {
		if *deployment.Spec.Replicas != *modela.Spec.DataPlane.Replicas && *modela.Spec.DataPlane.Replicas > 0 {
			deployment.Spec.Replicas = modela.Spec.DataPlane.Replicas
			updateDeployment = true
		}
	}

	if modela.Spec.DataPlane.Resources != nil {
		if !reflect.DeepEqual(*modela.Spec.DataPlane.Resources, deployment.Spec.Template.Spec.Containers[0].Resources) {
			deployment.Spec.Template.Spec.Containers[0].Resources = *modela.Spec.DataPlane.Resources
			updateDeployment = true
		}
	}

	if updateDeployment {
		if err := r.Update(ctx, &deployment); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{}, nil
}
