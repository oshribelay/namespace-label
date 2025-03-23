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

package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	namespacelabelv1alpha1 "github.com/oshribelay/namespace-label/api/v1alpha1"
)

const namespaceLabelAlreadyExists = "a NamespaceLabel already exists in this namespace"

// nolint:unused
// log is for logging in this package.
var namespacelabellog = logf.Log.WithName("namespacelabel-resource")

// SetupNamespaceLabelWebhookWithManager registers the webhook for NamespaceLabel in the manager.
func SetupNamespaceLabelWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&namespacelabelv1alpha1.NamespaceLabel{}).
		WithValidator(&NamespaceLabelCustomValidator{
			Client: mgr.GetClient(),
		}).
		Complete()
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-namespacelabel-dana-io-v1alpha1-namespacelabel,mutating=false,failurePolicy=fail,sideEffects=None,groups=namespacelabel.dana.io,resources=namespacelabels,verbs=create;update,versions=v1alpha1,name=vnamespacelabel-v1alpha1.kb.io,admissionReviewVersions=v1

// NamespaceLabelCustomValidator struct is responsible for validating the NamespaceLabel resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type NamespaceLabelCustomValidator struct {
	Client client.Client
}

var _ webhook.CustomValidator = &NamespaceLabelCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type NamespaceLabel.
func (v *NamespaceLabelCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	namespacelabel, ok := obj.(*namespacelabelv1alpha1.NamespaceLabel)
	if !ok {
		return nil, fmt.Errorf("expected a NamespaceLabel object but got %T", obj)
	}
	namespacelabellog.Info("Validate creation of NamespaceLabel", "name", namespacelabel.GetName())
	exists, err := v.checkIfNamespaceLabelExistsInNamespace(ctx, namespacelabel)
	if err != nil {
		return nil, err
	}
	if exists {
		return admission.Warnings{namespaceLabelAlreadyExists}, errors.New(namespaceLabelAlreadyExists)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type NamespaceLabel.
func (v *NamespaceLabelCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	namespacelabel, ok := newObj.(*namespacelabelv1alpha1.NamespaceLabel)
	if !ok {
		return nil, fmt.Errorf("expected a NamespaceLabel object for the newObj but got %T", newObj)
	}
	namespacelabellog.Info("Validation for NamespaceLabel upon update", "name", namespacelabel.GetName())

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type NamespaceLabel.
func (v *NamespaceLabelCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	namespacelabel, ok := obj.(*namespacelabelv1alpha1.NamespaceLabel)
	if !ok {
		return nil, fmt.Errorf("expected a NamespaceLabel object but got %T", obj)
	}
	namespacelabellog.Info("Validation for NamespaceLabel upon deletion", "name", namespacelabel.GetName())

	return nil, nil
}

func (v *NamespaceLabelCustomValidator) checkIfNamespaceLabelExistsInNamespace(ctx context.Context, namespaceLabel *namespacelabelv1alpha1.NamespaceLabel) (bool, error) {
	existingNsLabels := namespacelabelv1alpha1.NamespaceLabelList{}
	if err := v.Client.List(ctx, &existingNsLabels, client.InNamespace(namespaceLabel.Namespace)); err != nil {
		namespacelabellog.Error(err, "Failed to list NamespaceLabels", "name", namespaceLabel.GetName())
		return true, err
	}

	if len(existingNsLabels.Items) > 0 {
		return true, nil
	}

	return false, nil
}
