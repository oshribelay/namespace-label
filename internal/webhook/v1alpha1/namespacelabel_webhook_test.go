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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	namespacelabelv1alpha1 "github.com/oshribelay/namespace-label/api/v1alpha1"
	// TODO (user): Add any additional imports if needed
)

var _ = Describe("NamespaceLabel Webhook", func() {
	var (
		obj       *namespacelabelv1alpha1.NamespaceLabel
		oldObj    *namespacelabelv1alpha1.NamespaceLabel
		validator NamespaceLabelCustomValidator
	)

	BeforeEach(func() {
		obj = &namespacelabelv1alpha1.NamespaceLabel{}
		oldObj = &namespacelabelv1alpha1.NamespaceLabel{}
		scheme := runtime.NewScheme()
		Expect(namespacelabelv1alpha1.AddToScheme(scheme)).To(Succeed())
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		validator = NamespaceLabelCustomValidator{
			Client: fakeClient,
		}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
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
		// TODO (user): Add any teardown logic common to all tests
	})

	Context("When creating or updating NamespaceLabel under Validating Webhook", func() {
		It("should allow creation of namespacelabel if none are present in the same namespace", func() {
			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err.Error()).To(ContainSubstring("admission.Request not found in context"))
		})

		It("should deny creation of namespacelabel if one is present in the same namespace", func() {
			Expect(validator.Client.Create(ctx, obj)).To(Succeed())

			newObj := &namespacelabelv1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{Name: "second-test-object"},
				Spec: namespacelabelv1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{"env": "test"},
				},
			}

			_, err := validator.ValidateCreate(ctx, newObj)
			Expect(err).To(HaveOccurred())
		})
		// It("Should deny creation if a required field is missing", func() {
		//     By("simulating an invalid creation scenario")
		//     obj.SomeRequiredField = ""
		//     Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		// })
		//
		// It("Should admit creation if all required fields are present", func() {
		//     By("simulating an invalid creation scenario")
		//     obj.SomeRequiredField = "valid_value"
		//     Expect(validator.ValidateCreate(ctx, obj)).To(BeNil())
		// })
		//
		// It("Should validate updates correctly", func() {
		//     By("simulating a valid update scenario")
		//     oldObj.SomeRequiredField = "updated_value"
		//     obj.SomeRequiredField = "updated_value"
		//     Expect(validator.ValidateUpdate(ctx, oldObj, obj)).To(BeNil())
		// })
	})

})
