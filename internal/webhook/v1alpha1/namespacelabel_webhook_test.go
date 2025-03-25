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
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	namespacelabelv1alpha1 "github.com/oshribelay/namespace-label/api/v1alpha1"
)

var _ = Describe("NamespaceLabel Webhook", func() {
	var (
		obj     *namespacelabelv1alpha1.NamespaceLabel
		oldObj  *namespacelabelv1alpha1.NamespaceLabel
		webhook NamespaceLabelWebhook
	)

	BeforeEach(func() {
		obj = &namespacelabelv1alpha1.NamespaceLabel{}
		oldObj = &namespacelabelv1alpha1.NamespaceLabel{}
		scheme := runtime.NewScheme()
		Expect(namespacelabelv1alpha1.AddToScheme(scheme)).To(Succeed())
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		webhook = NamespaceLabelWebhook{
			Client: fakeClient,
		}
		Expect(webhook).NotTo(BeNil(), "Expected validator to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")

		obj = &namespacelabelv1alpha1.NamespaceLabel{
			Spec: namespacelabelv1alpha1.NamespaceLabelSpec{
				Labels: map[string]string{"env": "test"},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "test-object"},
		}

		oldObj = obj.DeepCopy()
	})

	AfterEach(func() {
	})

	Context("When creating or updating NamespaceLabel under Validating Webhook", func() {
		It("should allow creation of namespacelabel if none are present in the same namespace", func() {
			request := createAdmissionRequest(obj, admissionv1.Create)
			response := webhook.Handle(ctx, request)
			Expect(response.Allowed).To(BeTrue())
		})

		It("should deny creation of namespacelabel if one is present in the same namespace", func() {
			Expect(webhook.Client.Create(ctx, obj)).To(Succeed())

			newObj := &namespacelabelv1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{Name: "second-test-object"},
				Spec: namespacelabelv1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{"env": "test"},
				},
			}

			request := createAdmissionRequest(newObj, admissionv1.Create)
			response := webhook.Handle(ctx, request)
			Expect(response.Allowed).To(Not(BeTrue()))
		})
	})

})

func createAdmissionRequest(obj *namespacelabelv1alpha1.NamespaceLabel, operation admissionv1.Operation) admission.Request {
	rawObj, err := json.Marshal(obj)
	Expect(err).To(Not(HaveOccurred()))
	request := admissionv1.AdmissionRequest{
		UID:       "test-uid",
		Kind:      metav1.GroupVersionKind{Group: "namespacelabel.dana.io", Version: "v1alpha1", Kind: "NamespaceLabel"},
		Resource:  metav1.GroupVersionResource{Group: "namespacelabel.dana.io", Version: "v1alpha1", Resource: "namespacelabels"},
		Operation: operation,
		Object:    runtime.RawExtension{Raw: rawObj},
	}

	admRequest := admission.Request{AdmissionRequest: request}
	return admRequest
}
