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
	"fmt"
	namespacelabelv1alpha1 "github.com/oshribelay/namespace-label/api/v1alpha1"
	"github.com/oshribelay/namespace-label/internal/controller/finalizer"
	"github.com/oshribelay/namespace-label/internal/controller/resources"
	"github.com/oshribelay/namespace-label/internal/controller/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
func (r *NamespaceLabelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling NamespaceLabel", "NamespaceLabel", req.NamespacedName)

	nsLabel := namespacelabelv1alpha1.NamespaceLabel{}
	if err := r.Get(ctx, req.NamespacedName, &nsLabel); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to fetch NamespaceLabel")
		return ctrl.Result{}, err
	}

	namespace := corev1.Namespace{}
	if err := r.Get(ctx, types.NamespacedName{Name: nsLabel.Namespace}, &namespace); err != nil {
		logger.Error(err, "Failed to fetch Namespace")
		return ctrl.Result{}, err
	}

	namespaceLabelList := namespacelabelv1alpha1.NamespaceLabelList{}
	if err := r.List(ctx, &namespaceLabelList, client.InNamespace(namespace.Name)); err != nil {
		return ctrl.Result{}, err
	}

	protectedLabelsConfigMap := corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-protected-labels-configmap", nsLabel.Name), Namespace: nsLabel.Namespace}, &protectedLabelsConfigMap); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("Creating ConfigMap %s-protected-labels-configmap", nsLabel.Name))
			protectedLabelsConfigMap, err = resources.CreateConfigMap(&nsLabel, r.Client, ctx)
			if err != nil {
				logger.Error(err, "Failed to create ConfigMap")
				return ctrl.Result{}, err
			}
		} else {
			logger.Error(err, "Failed to fetch ConfigMap")
			return ctrl.Result{}, err
		}
	}
	protectedPrefixes := protectedLabelsConfigMap.Data

	if err := resources.ValidateNamespaceLabel(nsLabel.Spec.Labels, protectedPrefixes); err != nil {
		return ctrl.Result{}, err
	}

	if !nsLabel.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, req, namespace, nsLabel, protectedPrefixes); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if err := finalizer.EnsureFinalizer(ctx, r.Client, &nsLabel); err != nil {
		logger.Error(err, "unable to add finalizer")
		return ctrl.Result{Requeue: true}, err
	}

	if err := r.updateNamespaceLabels(ctx, req, namespaceLabelList, namespace, protectedPrefixes); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// updateNamespaceLabels updates the labels of the namespace according to the NamespaceLabel resource.
func (r *NamespaceLabelReconciler) updateNamespaceLabels(ctx context.Context, req ctrl.Request, namespaceLabelList namespacelabelv1alpha1.NamespaceLabelList, namespace corev1.Namespace, protectedPrefixes map[string]string) error {
	logger := log.FromContext(ctx)
	desiredLabels := make(map[string]string)
	for _, label := range namespaceLabelList.Items {
		for key, value := range label.Spec.Labels {
			if !utils.IsReservedLabel(key, protectedPrefixes) {
				desiredLabels[key] = value
			}
		}
	}

	updatedLabels := make(map[string]string)
	for key, value := range namespace.GetLabels() {
		updatedLabels[key] = value
	}

	for key := range updatedLabels {
		if !utils.IsReservedLabel(key, protectedPrefixes) {
			if _, exists := desiredLabels[key]; !exists {
				delete(updatedLabels, key)
			}
		}
	}

	for k, v := range desiredLabels {
		if !utils.IsReservedLabel(k, protectedPrefixes) {
			updatedLabels[k] = v
		}
	}

	if !utils.EqualLabels(updatedLabels, namespace.Labels) {
		namespace.Labels = updatedLabels
		if err := r.Update(ctx, &namespace); err != nil {
			logger.Error(err, "Failed to update NamespaceLabel")
			return err
		}
		logger.Info("Updated Namespace Successfully", "namespace", req.Namespace)
	} else {
		logger.Info("Namespace label is already up to date no changes needed")
	}
	return nil
}

// handleDeletion handles the deletion of the NamespaceLabel object.
func (r *NamespaceLabelReconciler) handleDeletion(ctx context.Context, req ctrl.Request, namespace corev1.Namespace, namespaceLabel namespacelabelv1alpha1.NamespaceLabel, protectedPrefixes map[string]string) error {
	logger := log.FromContext(ctx)
	namespaceLabelList := namespacelabelv1alpha1.NamespaceLabelList{}
	if err := r.List(ctx, &namespaceLabelList, client.InNamespace(namespace.Name)); err != nil {
		logger.Error(err, "Failed to fetch NamespaceLabels")
		return err
	}

	for index, label := range namespaceLabelList.Items {
		if label.Name == namespaceLabel.Name && label.Namespace == namespaceLabel.Namespace {
			namespaceLabelList.Items[index] = namespaceLabelList.Items[len(namespaceLabelList.Items)-1]
			namespaceLabelList.Items = namespaceLabelList.Items[:len(namespaceLabelList.Items)-1]
		}
	}

	if err := r.updateNamespaceLabels(ctx, req, namespaceLabelList, namespace, protectedPrefixes); err != nil {
		logger.Error(err, "Failed to remove deleted namespaces from the namespace")
		return err
	}
	if err := finalizer.RemoveFinalizer(ctx, r.Client, &namespaceLabel); err != nil {
		logger.Error(err, "Failed to remove finalizer")
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceLabelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&namespacelabelv1alpha1.NamespaceLabel{}).
		Complete(r)
}
