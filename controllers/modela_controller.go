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
	"github.com/metaprov/modelaapi/pkg/util"
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

var ComponentNotInstalledByModelaError = errors.New("component not installed by Modela")

// ModelaReconciler reconciles a Modela object
type ModelaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=*,verbs=*
//+kubebuilder:rbac:groups="batch",resources=*,verbs=*
//+kubebuilder:rbac:groups="extensions",resources=*,verbs=*
//+kubebuilder:rbac:groups="apps",resources=*,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Modela object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *ModelaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var modela managementv1alpha1.Modela
	if err := r.Get(ctx, req.NamespacedName, &modela); err != nil {
		logger.Error(err, "unable to fetch Modela")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	oldStatus := *modela.Status.DeepCopy()

	result, err := r.Install(ctx, &modela)
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
	if err != nil || result.Requeue {
		goto updateStatus
	}

	result, err = r.reconcileControlPlane(ctx, &modela)
	if err != nil || result.Requeue {
		goto updateStatus
	}

	result, err = r.reconcileDataPlane(ctx, &modela)
	if err != nil || result.Requeue {
		goto updateStatus
	}

	result, err = r.reconcileApiGateway(ctx, &modela)
	if err != nil || result.Requeue {
		goto updateStatus
	}

updateStatus:
	statusResult, statusErr := r.UpdateStatus(ctx, oldStatus, modela)
	if statusResult.Requeue {
		return statusResult, statusErr
	}
	return result, err
}

func (r ModelaReconciler) UpdateStatus(ctx context.Context, oldStatus managementv1alpha1.ModelaStatus, modela managementv1alpha1.Modela) (ctrl.Result, error) {
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
		logger.Info("Updated Modela status")
	}
	return ctrl.Result{}, nil
}

func (r ModelaReconciler) UpdatePhase(ctx context.Context, modela *managementv1alpha1.Modela, phase managementv1alpha1.ModelaPhase) {
	now := metav1.Now()
	modela.Status.LastUpdated = &now
	modela.Status.Phase = phase
	_ = r.Status().Update(ctx, modela)
}

func (r ModelaReconciler) isStateEqual(old managementv1alpha1.ModelaStatus, new managementv1alpha1.ModelaStatus) bool {
	return old.InstalledVersion == new.InstalledVersion &&
		old.ModelaSystemInstalled == new.ModelaSystemInstalled &&
		reflect.DeepEqual(old.LicenseToken, new.LicenseToken) &&
		reflect.DeepEqual(old.Conditions, new.Conditions) &&
		reflect.DeepEqual(old.Tenants, new.Tenants)

}

func (r *ModelaReconciler) Install(ctx context.Context, modela *managementv1alpha1.Modela) (ctrl.Result, error) {
	// Cert Manager
	certManager := NewCertManager(*modela.Spec.CertManager.CertManagerChartVersion)
	result, err := r.reconcileComponent(ctx, certManager, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	// Object Storage (Minio)
	objectStore := NewObjectStorage(*modela.Spec.ObjectStore.MinioChartVersion)
	result, err = r.reconcileComponent(ctx, objectStore, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	// PostgreSQL
	database := NewDatabase(*modela.Spec.SystemDatabase.PostgresChartVersion)
	result, err = r.reconcileComponent(ctx, database, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	// Loki
	loki := NewLoki(*modela.Spec.Observability.LokiVersion)
	result, err = r.reconcileComponent(ctx, loki, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	// Prometheus
	prom := NewPrometheus(*modela.Spec.Observability.PrometheusVersion)
	result, err = r.reconcileComponent(ctx, prom, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	// Modela System
	modelaSystem := NewModelaSystem(*modela.Spec.ModelaChart.ChartVersion)
	result, err = r.reconcileComponent(ctx, modelaSystem, modela)
	if err != nil || result.Requeue {
		return result, err
	}

	/*defaultTenant := NewDefaultTenant(*modela.Spec.DefaultTenantChart.ChartVersion)
	installed, err = defaultTenant.Installed(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !installed {
		// reconcile default tenant, make sure that all the items are as defined.
		result, err := r.reconcileComponent(ctx, defaultTenant, *modela)
		if err != nil || result.Requeue {
			return result, err
		}
	}*/

	modela.Status.Phase = managementv1alpha1.ModelaPhaseReady
	return ctrl.Result{}, err

}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&managementv1alpha1.Modela{}).
		Complete(r)
}

// Define the interface for modela components that can be reconciled
type ModelaComponent interface {
	IsEnabled(modela managementv1.Modela) bool
	Installed(ctx context.Context) (bool, error)
	Install(ctx context.Context, modela *managementv1.Modela) error
	Installing(ctx context.Context) (bool, error)
	Ready(ctx context.Context) (bool, error)
	Uninstall(ctx context.Context) error
	GetInstallPhase() managementv1alpha1.ModelaPhase
}

func (r *ModelaReconciler) reconcileComponent(ctx context.Context, component ModelaComponent, modela *managementv1.Modela) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	installed, err := component.Installed(ctx)
	if err != nil && err != ComponentNotInstalledByModelaError {
		logger.Error(err, "Failed to check if component is installed", "component", reflect.TypeOf(component).Name())
		return ctrl.Result{}, err
	}

	if !component.IsEnabled(*modela) {
		if err != ComponentNotInstalledByModelaError {
			r.UpdatePhase(ctx, modela, managementv1alpha1.ModelaPhaseUninstalling)
			err := component.Uninstall(ctx)
			if err != nil {
				logger.Error(err, "Failed to install component", "component", reflect.TypeOf(component).Name())
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if installed {
		return ctrl.Result{}, nil
	}

	installing, err := component.Installing(ctx)
	if err != nil {
		logger.Error(err, "Failed to check if component is installing", "component", reflect.TypeOf(component).Name())
		return ctrl.Result{}, err
	}

	if installing {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 0,
		}, nil
	} else {
		r.UpdatePhase(ctx, modela, component.GetInstallPhase())
		err := component.Install(ctx, modela)
		if err != nil {
			logger.Error(err, "Failed to install component", "component", reflect.TypeOf(component).Name())
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
}

func (r *ModelaReconciler) reconcileApiGateway(ctx context.Context, modela *managementv1alpha1.Modela) (ctrl.Result, error) {
	// get api gateway deployment
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
