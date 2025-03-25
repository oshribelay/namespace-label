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
	namespacelabelv1alpha1 "github.com/oshribelay/namespace-label/api/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const namespaceLabelAlreadyExists = "a NamespaceLabel already exists in this namespace"

// nolint:unused
// log is for logging in this package.
var namespacelabellog = logf.Log.WithName("namespacelabel-resource")

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-namespacelabel-dana-io-v1alpha1-namespacelabel,mutating=false,failurePolicy=fail,sideEffects=None,groups=namespacelabel.dana.io,resources=namespacelabels,verbs=create;update,versions=v1alpha1,name=vnamespacelabel-v1alpha1.kb.io,admissionReviewVersions=v1

// NamespaceLabelCustomValidator struct is responsible for validating the NamespaceLabel resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
//type NamespaceLabelCustomValidator struct {
//	Client client.Client
//}
//
//var _ webhook.CustomValidator = &NamespaceLabelCustomValidator{}

type NamespaceLabelWebhook struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (w *NamespaceLabelWebhook) InjectDecoder(decoder *admission.Decoder) error {
	w.decoder = decoder
	return nil
}

func (w *NamespaceLabelWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	namespaceLabel := namespacelabelv1alpha1.NamespaceLabel{}
	if err := json.Unmarshal(req.Object.Raw, &namespaceLabel); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if req.AdmissionRequest.Operation == admissionv1.Create {
		exists, err := w.checkIfNamespaceLabelExistsInNamespace(ctx, &namespaceLabel)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}

		if exists {
			return admission.Denied(namespaceLabelAlreadyExists)
		}
	}

	return admission.Allowed("Validation passed")
}

func (w *NamespaceLabelWebhook) checkIfNamespaceLabelExistsInNamespace(ctx context.Context, namespaceLabel *namespacelabelv1alpha1.NamespaceLabel) (bool, error) {
	existingNsLabels := namespacelabelv1alpha1.NamespaceLabelList{}
	if err := w.Client.List(ctx, &existingNsLabels, client.InNamespace(namespaceLabel.Namespace)); err != nil {
		namespacelabellog.Error(err, "Failed to list NamespaceLabels", "name", namespaceLabel.GetName())
		return true, err
	}

	if len(existingNsLabels.Items) > 0 {
		return true, nil
	}

	return false, nil
}

// SetupNamespaceLabelWebhookWithManager registers the webhook for NamespaceLabel in the manager.
func SetupNamespaceLabelWebhookWithManager(mgr ctrl.Manager) error {
	webhook := &NamespaceLabelWebhook{Client: mgr.GetClient()}

	err := ctrl.NewWebhookManagedBy(mgr).For(&namespacelabelv1alpha1.NamespaceLabel{}).
		WithCustomPath("/validate-namespacelabel-dana-io-v1alpha1-namespacelabel").
		Complete()
	if err != nil {
		return err
	}

	mgr.GetWebhookServer().Register("/validate-namespacelabel-dana-io-v1alpha1-namespacelabel", &admission.Webhook{
		Handler: webhook,
	})

	return nil
}
