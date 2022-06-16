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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	managementv1alpha1 "github.com/metaprov/modela-operator/api/v1alpha1"
)

// ModelaReconciler reconciles a Modela object
type ModelaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=management.modela.ai,resources=modelas/finalizers,verbs=update
//+kubebuilder:rbac:groups=app,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app,resources=deployment/status,verbs=get

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

	certManager := NewCertManager()
	result, err := r.reconcileComponment(ctx, certManager)
	if err != nil || result.Requeue {
		return result, err
	}

	objectStore := NewObjectStorage()
	result, err = r.reconcileComponment(ctx, objectStore)
	if err != nil || result.Requeue {
		return result, err
	}

	database := NewDatabase()
	result, err = r.reconcileComponment(ctx, database)
	if err != nil || result.Requeue {
		return result, err
	}

	monitoring := NewMonitoring()
	result, err = r.reconcileComponment(ctx, monitoring)
	if err != nil || result.Requeue {
		return result, err
	}

	modelaSystem := NewModelaSystem()
	// reconcile modela system, make sure that all the items are as defined
	result, err = r.reconcileComponment(ctx, modelaSystem)
	if err != nil || result.Requeue {
		return result, err
	}

	defaultTenant := NewDefaultTenant()
	// reconcile default tenant, make sure that all the items are as defined.
	result, err = r.reconcileComponment(ctx, defaultTenant)
	if err != nil || result.Requeue {
		return result, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&managementv1alpha1.Modela{}).
		Complete(r)
}

// Define the interface for modela components that can be reconciled
type ModelaComponent interface {
	Installed() (bool, error)
	Install() error
	Installing() (bool, error)
	Ready() (bool, error)
	Uninstall() error
}

func (r *ModelaReconciler) reconcileComponment(ctx context.Context, component ModelaComponent) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	installed, err := component.Installed()
	if err != nil {
		logger.Error(err, "failed to check if database installed")
		return ctrl.Result{}, err
	}
	if !installed {
		installing, err := component.Installing()
		if err != nil {
			logger.Error(err, "failed to check if database installed")
			return ctrl.Result{}, err
		}
		if !installing {
			err := component.Install()
			if err != nil {
				logger.Error(err, "failed to check if database installed")
				return ctrl.Result{}, err
			}
		} else {
			// installing, dequeue the request
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 0,
			}, nil
		}
	}
	return ctrl.Result{}, nil
}

func (r *ModelaReconciler) updateSystemStatus(ctx context.Context, modela *managementv1alpha1.Modela) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
