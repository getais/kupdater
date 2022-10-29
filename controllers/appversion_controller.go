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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	opsv1alpha1 "github.com/getais/kupdater/api/v1alpha1"
)

// AppVersionReconciler reconciles a AppVersion object
type AppVersionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ops.getais.cloud,resources=appversions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ops.getais.cloud,resources=appversions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ops.getais.cloud,resources=appversions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AppVersionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	var log = ctrllog.Log.WithName("appversion.ops.getais.Reconcile").WithValues("namespace", req.Namespace, "name", req.Name)

	appver := &opsv1alpha1.AppVersion{}
	err := r.Get(ctx, req.NamespacedName, appver)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("ApppVersion resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get AppVersion.")
		return ctrl.Result{}, err
	}

	update := &opsv1alpha1.Update{}
	err = r.Get(ctx, types.NamespacedName{Name: appver.Name, Namespace: appver.Namespace}, update)
	if err != nil && errors.IsNotFound(err) {
		update = r.NewUpdate(appver)
		log.Info("Creating a new Update")
		err = r.Create(ctx, update)
		if err != nil {
			log.Error(err, "Failed to create new Update")
			return ctrl.Result{RequeueAfter: 2 * time.Minute}, err
		}
		// Update created successfully
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Update")
		return ctrl.Result{}, err
	}
	log.Info("Reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppVersionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1alpha1.AppVersion{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *AppVersionReconciler) NewUpdate(a *opsv1alpha1.AppVersion) *opsv1alpha1.Update {

	Update := &opsv1alpha1.Update{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Name,
			Namespace: a.Namespace,
		},
		Spec: opsv1alpha1.UpdateSpec{
			Versioning: opsv1alpha1.UpdateVersioning{
				Sources: []opsv1alpha1.UpdateSource{
					{
						Name:    a.Spec.Name,
						Type:    a.Spec.Type,
						Source:  a.Spec.Source,
						Version: a.Spec.Version,
					},
				},
			},
		},
	}
	// Set Application instance as the owner and controller
	ctrl.SetControllerReference(a, Update, r.Scheme)
	return Update
}
