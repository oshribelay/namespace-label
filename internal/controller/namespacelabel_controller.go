/*
Copyright 2025.

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

package controller

import (
	"context"
	namespacelabelv1alpha1 "github.com/oshribelay/namespace-label/api/v1alpha1"
	"github.com/oshribelay/namespace-label/internal/controller/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NamespaceLabelReconciler reconciles a NamespaceLabel object
type NamespaceLabelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=namespacelabel.dana.io,resources=namespacelabels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=namespacelabel.dana.io,resources=namespacelabels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=namespacelabel.dana.io,resources=namespacelabels/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NamespaceLabel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *NamespaceLabelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling NamespaceLabel", "NamespaceLabel", req.NamespacedName)

	var nsLabel namespacelabelv1alpha1.NamespaceLabel
	if err := r.Get(ctx, req.NamespacedName, &nsLabel); err != nil {
		if errors.IsNotFound(err) {
			// NamespaceLabel was deleted, clean up if needed
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to fetch NamespaceLabel")
		return ctrl.Result{}, err
	}

	var namespace corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: nsLabel.Namespace}, &namespace); err != nil {
		logger.Error(err, "Failed to fetch Namespace")
		return ctrl.Result{}, err
	}

	// Fetch all NamespaceLabel objects in the namespace to merge labels from multiple NamespaceLabel CRs
	var namespaceLabelList namespacelabelv1alpha1.NamespaceLabelList
	if err := r.List(ctx, &namespaceLabelList, client.InNamespace(namespace.Name)); err != nil {
		return ctrl.Result{}, err
	}

	var protectedPrefixes = map[string]struct{}{
		"k8s.io/":        {},
		"kubernetes.io/": {},
		"openshift.io/":  {},
	}

	desiredLabels := make(map[string]string)
	for _, label := range namespaceLabelList.Items {
		for key, value := range label.Spec.Labels {
			desiredLabels[key] = value
		}
	}

	updatedLabels := make(map[string]string)
	// copy current namespace labels
	for k, v := range namespace.GetLabels() {
		updatedLabels[k] = v
	}

	// apply desired labels but skip protected ones
	for k, v := range desiredLabels {
		if !utils.IsReservedLabel(k, protectedPrefixes) {
			updatedLabels[k] = v
		}
	}

	if !utils.EqualLabels(updatedLabels, nsLabel.GetLabels()) {
		namespace.Labels = updatedLabels
		if err := r.Update(ctx, &namespace); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		logger.Info("Namespace label is already up to date no changes needed")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceLabelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&namespacelabelv1alpha1.NamespaceLabel{}).
		Complete(r)
}
