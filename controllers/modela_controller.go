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
	"github.com/metaprov/modela-operator/controllers/components"
	"github.com/metaprov/modelaapi/pkg/util"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
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

// ModelaReconciler reconciles a Modela object
type ModelaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=catalog.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=team.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=data.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inference.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infra.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=training.modela.ai,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas/finalizers,verbs=update
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=*,verbs=*
//+kubebuilder:rbac:groups="extensions",resources=*,verbs=*
//+kubebuilder:rbac:groups="apps",resources=*,verbs=*
//+kubebuilder:rbac:groups="core",resources=*,verbs=*
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;delete;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=services,verbs=get;list;watch;create;update;delete;patch
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=*,verbs=get;list;watch;create;update;delete;patch
//+kubebuilder:rbac:groups=cert-manager.io/v1,resources=*,verbs=get;list;watch;create;update;delete;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *ModelaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting requeue again")

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
	}
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
	if !r.isStateEqual(modela.Status, oldStatus) || modela.Generation > modela.Status.ObservedGeneration {
		modela.Status.ObservedGeneration = modela.Generation
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
		old.FailureMessage == new.FailureMessage &&
		old.Phase == new.Phase &&
		reflect.DeepEqual(old.LicenseToken, new.LicenseToken) &&
		reflect.DeepEqual(old.Conditions, new.Conditions) &&
		reflect.DeepEqual(old.Tenants, new.Tenants)

}

func (r *ModelaReconciler) Install(ctx context.Context, modela *managementv1alpha1.Modela) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Cert Manager
	certManager := components.NewCertManager(modela.Spec.CertManager.CertManagerChartVersion)
	result, err := r.reconcileComponent(ctx, certManager, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	// Object Storage (Minio)
	objectStore := components.NewObjectStorage(modela.Spec.ObjectStore.MinioChartVersion)
	result, err = r.reconcileComponent(ctx, objectStore, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	// Loki
	loki := components.NewLoki(modela.Spec.Observability.LokiVersion)
	result, err = r.reconcileComponent(ctx, loki, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	// Grafana
	grafana := components.NewGrafana(modela.Spec.Observability.GrafanaVersion)
	result, err = r.reconcileComponent(ctx, grafana, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	// Prometheus
	prom := components.NewPrometheus(modela.Spec.Observability.PrometheusVersion)
	result, err = r.reconcileComponent(ctx, prom, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	// PostgreSQL
	database := components.NewDatabase(modela.Spec.SystemDatabase.PostgresChartVersion)
	result, err = r.reconcileComponent(ctx, database, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	// Modela System
	modelaSystem := components.NewModelaSystem(modela.Spec.Distribution)
	result, err = r.reconcileComponent(ctx, modelaSystem, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	if ready, err := modelaSystem.Ready(ctx); err != nil || !ready {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 10 * time.Second,
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
				RequeueAfter: 5 * time.Minute,
			}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 10 * time.Second,
		}, err
	}

	result, err = r.reconcileTenants(ctx, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	if modela.Spec.Distribution != modela.Status.InstalledVersion {
		logger.Info("Applying new distribution", "version", modelaSystem.ModelaVersion)
		if err := modelaSystem.InstallNewVersion(ctx, modela); err != nil {
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&managementv1alpha1.Modela{}).
		Owns(&appsv1.Deployment{}).
		Owns(&v1.Service{}).
		Owns(&v1.ServiceAccount{}).
		Owns(&v1.Secret{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}

func (r *ModelaReconciler) reconcileTenants(ctx context.Context, modela *managementv1.Modela) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var tenants = make(map[string]bool)
	for _, tenantSpec := range modela.Spec.Tenants {
		tenants[tenantSpec.Name] = true
		tenant := components.NewTenant(tenantSpec.Name)
		if installed, err := tenant.Installed(ctx); err != nil {
			return ctrl.Result{}, err
		} else if !installed {
			if result, _ := r.updatePhase(ctx, modela, managementv1alpha1.ModelaPhaseInstallingTenant); result.Requeue {
				return result, nil
			}
			if err := tenant.Install(ctx, modela, tenantSpec); err != nil {
				logger.Error(err, "Failed to install tenant", "name", tenant.Name)
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 5 * time.Minute,
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

func (r *ModelaReconciler) reconcileComponent(ctx context.Context, component ModelaComponent, modela *managementv1.Modela) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	installed, err := component.Installed(ctx)
	// fmt.Println(installed, err, component.GetInstallPhase())
	if err != nil && err != managementv1alpha1.ComponentNotInstalledByModelaError && err != managementv1alpha1.ComponentMissingResourcesError {
		logger.Error(err, "Failed to check if component is installed", "component", reflect.TypeOf(component).Name())
		return ctrl.Result{}, err
	}

	if !component.IsEnabled(*modela) {
		if err != managementv1alpha1.ComponentNotInstalledByModelaError && installed {
			if result, _ := r.updatePhase(ctx, modela, managementv1alpha1.ModelaPhaseUninstalling); result.Requeue {
				return result, nil
			}
			err := component.Uninstall(ctx, modela)
			if err != nil {
				logger.Error(err, "Failed to uninstall component", "component", reflect.TypeOf(component).Name())
				return ctrl.Result{Requeue: true}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	if installed {
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
		return ctrl.Result{Requeue: true}, nil
	}
}

func (r *ModelaReconciler) reconcileApiGateway(ctx context.Context, modela *managementv1alpha1.Modela) (ctrl.Result, error) {
	var deployment appsv1.Deployment

	name := types.NamespacedName{
		Namespace: "modela-system",
		Name:      "modela-api-gateway",
	}

	if err := r.Get(ctx, name, &deployment); err != nil {
		return ctrl.Result{}, err
	}

	if *deployment.Spec.Replicas != *modela.Spec.ApiGateway.Replicas {
		err := r.Update(ctx, &deployment)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModelaReconciler) reconcileControlPlane(ctx context.Context, modela *managementv1alpha1.Modela) (ctrl.Result, error) {
	var deployment appsv1.Deployment

	name := types.NamespacedName{
		Namespace: "modela-system",
		Name:      "modela-control-plane",
	}

	if err := r.Get(ctx, name, &deployment); err != nil {
		return ctrl.Result{}, err
	}

	if *deployment.Spec.Replicas != *modela.Spec.ControlPlane.Replicas {
		err := r.Update(ctx, &deployment)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModelaReconciler) reconcileDataPlane(ctx context.Context, modela *managementv1alpha1.Modela) (ctrl.Result, error) {
	var deployment appsv1.Deployment

	name := types.NamespacedName{
		Namespace: "modela-system",
		Name:      "modela-data-plane",
	}

	if err := r.Get(ctx, name, &deployment); err != nil {
		return ctrl.Result{}, err
	}

	if *deployment.Spec.Replicas != *modela.Spec.ControlPlane.Replicas {
		err := r.Update(ctx, &deployment)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
