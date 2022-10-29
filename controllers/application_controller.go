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
	"net/url"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	argov1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	opsv1alpha1 "github.com/getais/kupdater/api/v1alpha1"
)

// AppVersionReconciler reconciles a AppVersion object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	var log = ctrllog.Log.WithName("application.argoproj.Reconcile").WithValues("namespace", req.Namespace, "name", req.Name)

	// Lookup the Update instance for this reconcile request
	app := &argov1alpha1.Application{}
	err := r.Get(ctx, req.NamespacedName, app)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Applicaiton.argoproj not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Application.argoproj.")
		return ctrl.Result{}, err
	}

	log.Info("Reconciling")
	appver := &opsv1alpha1.AppVersion{}
	err = r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Spec.Destination.Namespace}, appver)
	if err != nil && errors.IsNotFound(err) {

		// If Helm is present in App
		if app.Spec.Source.Helm != nil {

			// If helm repo url is valid http url
			_, err := url.ParseRequestURI(app.Spec.Source.RepoURL)
			if err != nil {
				log.Info("Skipping. Helm RepoUrl is not http")
				return ctrl.Result{}, nil
			}
			appver = r.NewAppver(app)
			log.Info("Creating a new AppVersion")
			err = r.Create(ctx, appver)
			if err != nil {
				log.Error(err, "Failed to create new AppVersion")
				return ctrl.Result{RequeueAfter: 2 * time.Minute}, err
			}

		}
		// AppVersion created successfully - return and requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Application")
		return ctrl.Result{}, err
	}

	log.Info("Reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&argov1alpha1.Application{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *ApplicationReconciler) NewAppver(a *argov1alpha1.Application) *opsv1alpha1.AppVersion {

	AppVer := &opsv1alpha1.AppVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Name,
			Namespace: a.Spec.Destination.Namespace,
		},
		Spec: opsv1alpha1.UpdateSource{
			Name:    a.Name,
			Type:    "helm",
			Source:  a.Spec.Source.RepoURL,
			Version: a.Spec.Source.TargetRevision,
		},
		Status: opsv1alpha1.AppVersionStatus{},
	}
	// Set Application instance as the owner and controller
	ctrl.SetControllerReference(a, AppVer, r.Scheme)
	return AppVer

}
