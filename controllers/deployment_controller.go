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
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
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
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	var log = ctrllog.Log.WithName("deployment.apps.Reconcile").WithValues("namespace", req.Namespace, "name", req.Name)

	// Lookup the Deployment instance for this reconcile request
	dep := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Deployment.apps not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Deployment.apps.")
		return ctrl.Result{}, err
	}

	// log.Info("Reconciling")
	appver := &opsv1alpha1.AppVersion{}
	err = r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, appver)
	if err != nil && errors.IsNotFound(err) {

		// If Pod has annotations
		if _, enabled := dep.Annotations["kupdater.ops.getais.cloud/enabled"]; enabled {

			// If Source repo annotation is present
			if _, source := dep.Annotations["kupdater.ops.getais.cloud/source"]; source {

				// Create AppVersion to track updates on
				appversions, err := r.NewAppver(dep)
				if err != nil {
					log.Error(err, "Invalid Deployment")
					return ctrl.Result{RequeueAfter: 2 * time.Minute}, err
				}

				for _, appver := range appversions {
					// Create Appversion{
					log.Info("Creating a new AppVersion")
					err = r.Create(ctx, appver)
					if err != nil {
						log.Error(err, "Failed to create new AppVersion")
						return ctrl.Result{RequeueAfter: 2 * time.Minute}, err
					}
				}
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
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *DeploymentReconciler) NewAppver(a *appsv1.Deployment) (AppVersions []*opsv1alpha1.AppVersion, err error) {

	Source, success := a.Annotations["kupdater.ops.getais.cloud/source"]
	if !success {
		return []*opsv1alpha1.AppVersion{}, fmt.Errorf("Missing AppVersion source")
	}

	Type, success := a.Annotations["kupdater.ops.getais.cloud/type"]
	if !success {
		return []*opsv1alpha1.AppVersion{}, fmt.Errorf("Missing AppVersion type")
	}

	var Version string
	Version, _ = a.Annotations["kupdater.ops.getais.cloud/version"]

	for _, Container := range a.Spec.Template.Spec.Containers {
		AppVer := new(opsv1alpha1.AppVersion)

		if Type == "github" || Type == "Github" {
			Version = strings.Split(Container.Image, ":")[1]
		}

		AppVer = &opsv1alpha1.AppVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      a.Name,
				Namespace: a.Namespace,
			},
			Spec: opsv1alpha1.UpdateSource{
				Name:    a.Name,
				Type:    Type,
				Source:  Source,
				Version: Version,
			},
			Status: opsv1alpha1.AppVersionStatus{},
		}
		// Set Application instance as the owner and controller
		ctrl.SetControllerReference(a, AppVer, r.Scheme)
		AppVersions = append(AppVersions, AppVer)
	}
	return AppVersions, nil

}
