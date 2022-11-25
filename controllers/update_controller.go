package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Masterminds/semver"
	"github.com/getais/kupdater/api/v1alpha1"
	opsv1alpha1 "github.com/getais/kupdater/api/v1alpha1"
	"github.com/getais/kupdater/pkg/libs/helm"
	"github.com/google/go-github/github"
)

// UpdateReconciler reconciles a Update object
type UpdateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const reconcilePeriod string = "2m"

//+kubebuilder:rbac:groups=ops.getais.cloud,resources=updates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ops.getais.cloud,resources=updates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ops.getais.cloud,resources=updates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *UpdateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	var log = ctrllog.Log.WithName("update.ops.getais.Reconcile").WithValues("namespace", req.Namespace, "name", req.Name)

	// Lookup the Update instance for this reconcile request
	update := &opsv1alpha1.Update{}
	err := r.Get(ctx, req.NamespacedName, update)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Update resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Update.")
		return ctrl.Result{}, err
	}

	log.Info("Reconciling")

	update.Status.Phase = "UpToDate"
	update.Status.SyncTimestamp = time.Now().Format(time.RFC3339)
	meta.SetStatusCondition(&update.Status.Conditions, metav1.Condition{Type: "UpToDate", Status: metav1.ConditionFalse, Reason: "CheckingForUpdates", Message: "Currently checking for new updates"})

	// Check for updates
	update = checkUpdatesHelm(ctx, update)
	update, err = checkUpdatesGithub(ctx, update)
	if err != nil {
		log.Error(err, "Failed calling Update services")
		return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
	}

	// Update CRD status
	err = r.Status().Update(ctx, update)
	if err != nil {
		log.Error(err, "Failed to update status")
	}

	log.Info("Reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UpdateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1alpha1.Update{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func checkUpdatesGithub(ctx context.Context, Update *v1alpha1.Update) (*v1alpha1.Update, error) {
	if len(Update.Spec.Versioning.Sources) > 0 {
		for _, s := range Update.Spec.Versioning.Sources {

			if s.Type == "github" || s.Type == "Github" {
				client := github.NewClient(nil)

				// https://github.com/argoproj/argo-cd
				var owner = strings.Split(s.Source, "/")[3]
				var repo = strings.Split(s.Source, "/")[4]

				// Fetch latest Github release
				release, _, err := client.Repositories.GetLatestRelease(context.Background(), owner, repo)
				if err != nil {
					return nil, err
				}

				LatestVersion := *release.TagName

				meta.SetStatusCondition(&Update.Status.Conditions, metav1.Condition{Type: "UpToDate", Status: metav1.ConditionTrue, Reason: "NoUpdatesAvailable", Message: "No updates were found"})
				if s.Version != LatestVersion {
					Update.Status.Phase = fmt.Sprintf("Outdated (%s available)", LatestVersion)
					s.Version = LatestVersion
					meta.SetStatusCondition(&Update.Status.Conditions, metav1.Condition{Type: "Outdated", Status: metav1.ConditionTrue, Reason: "UpdatesAvailable", Message: fmt.Sprintf("New release available: %s", LatestVersion)})
				}
			}
		}
	}
	return Update, nil
}

func checkUpdatesHelm(ctx context.Context, Update *v1alpha1.Update) *v1alpha1.Update {
	if len(Update.Spec.Versioning.Sources) > 0 {
		for _, s := range Update.Spec.Versioning.Sources {

			if s.Type == "helm" || s.Type == "Helm" {
				var Helm helm.Helm

				Repo, err := Helm.GetReleases(s.Source)
				if err != nil {
					Update.Status.Phase = err.Error()
					return Update
				}
				Releases := Repo.Entries[s.Name]

				if len(Releases) > 0 {
					Versions := make([]*semver.Version, len(Releases))
					for i, r := range Releases {
						v, _ := semver.NewVersion(r.Version)
						Versions[i] = v
					}
					LatestVersion := Versions[0].String()

					meta.SetStatusCondition(&Update.Status.Conditions, metav1.Condition{Type: "UpToDate", Status: metav1.ConditionTrue, Reason: "NoUpdatesAvailable", Message: "No updates were found"})
					if s.Version != LatestVersion {
						Update.Status.Phase = fmt.Sprintf("Outdated (%s available)", LatestVersion)
						s.Version = LatestVersion
						meta.SetStatusCondition(&Update.Status.Conditions, metav1.Condition{Type: "Outdated", Status: metav1.ConditionTrue, Reason: "UpdatesAvailable", Message: fmt.Sprintf("New release available: %s", LatestVersion)})
					}
				}
			}
		}
	}
	return Update
}
