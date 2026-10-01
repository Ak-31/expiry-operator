package controller

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	expiryv1alpha1 "github.com/Ak-31/expiry-operator/api/v1alpha1"
)

const maxRequeue = 10 * time.Minute

// +kubebuilder:rbac:groups=expiry.example.io,resources=expirypolicies,verbs=get;list;watch
// +kubebuilder:rbac:groups=expiry.example.io,resources=expirypolicies/status,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps;persistentvolumeclaims;secrets,verbs=list;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=list;delete

type ExpiryPolicyReconciler struct {
	client.Client
}

func (r *ExpiryPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var policy expiryv1alpha1.ExpiryPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	selector, err := metav1.LabelSelectorAsSelector(&policy.Spec.Selector)
	if err != nil {
		return ctrl.Result{}, err
	}

	matched, deleted := 0, 0
	next := maxRequeue

	for _, target := range policy.Spec.TargetKinds {
		gv, err := schema.ParseGroupVersion(target.APIVersion)
		if err != nil {
			return ctrl.Result{}, err
		}

		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(gv.WithKind(target.Kind + "List"))
		err = r.List(ctx, list,
			client.InNamespace(policy.Namespace),
			client.MatchingLabelsSelector{Selector: selector})
		if err != nil {
			return ctrl.Result{}, err
		}

		for i := range list.Items {
			obj := &list.Items[i]
			created := obj.GetCreationTimestamp().Time
			remaining := time.Until(created.Add(policy.Spec.TTL.Duration))

			if remaining > 0 {
				matched++
				if remaining < next {
					next = remaining
				}
				continue
			}

			if policy.Spec.DryRun {
				logger.Info("dry run, would delete", "kind", target.Kind, "name", obj.GetName())
				matched++
				continue
			}

			err := r.Delete(ctx, obj, client.PropagationPolicy(metav1.DeletePropagationBackground))
			if client.IgnoreNotFound(err) != nil {
				return ctrl.Result{}, err
			}
			logger.Info("deleted expired object", "kind", target.Kind, "name", obj.GetName())
			deleted++
		}
	}

	policy.Status.MatchedCount = matched
	policy.Status.DeletedCount += deleted
	policy.Status.LastSweepTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, &policy); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: next}, nil
}

func (r *ExpiryPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Without this, our own status update would trigger another reconcile forever.
		For(&expiryv1alpha1.ExpiryPolicy{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}