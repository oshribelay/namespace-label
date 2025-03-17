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
	"time"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	namespacelabelv1alpha1 "github.com/oshribelay/namespace-label/api/v1alpha1"
)

var _ = Describe("NamespaceLabel Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			resourceName = "test-resource"
			timeout      = time.Second * 10
			interval     = time.Second * 1
		)

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		namespacelabel := &namespacelabelv1alpha1.NamespaceLabel{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind NamespaceLabel")
			// Delete any existing resources to ensure clean state.
			_ = k8sClient.Delete(ctx, namespacelabel)

			// Create the NamespaceLabel if it does not exist
			err := k8sClient.Get(ctx, typeNamespacedName, namespacelabel)
			if err != nil && errors.IsNotFound(err) {
				resource := &namespacelabelv1alpha1.NamespaceLabel{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: namespacelabelv1alpha1.NamespaceLabelSpec{
						Labels: map[string]string{
							"testLabel": "test-a",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			By("Cleanup the specific resource instance NamespaceLabel")
			// Get the latest version of the NamespaceLabel
			updatedNsLabel := &namespacelabelv1alpha1.NamespaceLabel{}
			err := k8sClient.Get(ctx, typeNamespacedName, updatedNsLabel)
			if err == nil {
				By("deleting the NamespaceLabel")
				Expect(k8sClient.Delete(ctx, updatedNsLabel)).To(Succeed())

				By("waiting for the controller to handle deletion")
				Eventually(func() string {
					err := k8sClient.Get(ctx, typeNamespacedName, updatedNsLabel)
					if errors.IsNotFound(err) {
						return "deleted"
					}
					if err != nil {
						return fmt.Sprintf("error checking resource: %v", err)
					}
					return fmt.Sprintf("resource still exists with finalizers: %v", updatedNsLabel.Finalizers)
				}, timeout, interval).Should(Equal("deleted"), "resource should be fully deleted by the controller")
			}
		})
		It("should create the NamespaceLabel if it doesn't exist", func() {
			By("verifying the NamespaceLabel resource exists")
			createdNsLabel := &namespacelabelv1alpha1.NamespaceLabel{}
			Eventually(func() error {
				return k8sClient.Get(ctx, typeNamespacedName, createdNsLabel)
			}, timeout, interval).Should(Succeed())

			By("verifying the ConfigMap exists")
			Eventually(func() bool {
				createdConfigMap := &corev1.ConfigMap{}
				err := k8sClient.Get(
					ctx,
					types.NamespacedName{Name: "namespace-label-protected-labels", Namespace: "namespace-label-system"},
					createdConfigMap,
				)
				if err != nil {
					return false
				}
				data, exists := createdConfigMap.Data["k8s.io"]
				return exists && data == ""
			}, timeout, interval).Should(BeTrue(), "ConfigMap should exist with protected labels")

			By("verifying details are correct")
			Expect(createdNsLabel.Spec.Labels).To(Equal(map[string]string{
				"testLabel": "test-a",
			}))
		})

		It("should update the namespace labels when NamespaceLabel is updated", func() {
			By("updating the NamespaceLabel resource")
			updatedNsLabel := &namespacelabelv1alpha1.NamespaceLabel{}
			Eventually(func() error {
				if err := k8sClient.Get(ctx, typeNamespacedName, updatedNsLabel); err != nil {
					return err
				}
				updatedNsLabel.Spec.Labels["testLabel"] = "test-b"
				return k8sClient.Update(ctx, updatedNsLabel)
			}, timeout, interval).Should(Succeed())

			By("verifying the labels were applied")
			namespace := &corev1.Namespace{}
			Eventually(func() map[string]string {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "default"}, namespace); err != nil {
					return nil
				}
				return namespace.Labels
			}, timeout, interval).Should(HaveKeyWithValue("testLabel", "test-b"), "namespace labels should match updated labels")
		})

		It("should delete the labels from the namespace when the NamespaceLabel is deleted", func() {
			By("deleting the NamespaceLabel resource")
			deletedNsLabel := &namespacelabelv1alpha1.NamespaceLabel{}
			Eventually(func() error {
				if err := k8sClient.Get(ctx, typeNamespacedName, deletedNsLabel); err != nil {
					return err
				}
				return k8sClient.Delete(ctx, deletedNsLabel)
			}, timeout, interval).Should(Succeed())

			By("waiting for the controller to handle deletion")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, deletedNsLabel)
				if errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("verifying the labels were deleted from the namespace")
			namespace := &corev1.Namespace{}
			Eventually(func() map[string]string {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "default"}, namespace); err != nil {
					return nil
				}
				return namespace.Labels
			}, timeout, interval).ShouldNot(HaveKeyWithValue("testLabel", "test-a"), "namespace labels should have been deleted")
		})

		It("should not apply protected label updates to the namespace", func() {
			By("creating the invalid NamespaceLabel")
			invalidResource := &namespacelabelv1alpha1.NamespaceLabel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-resource",
					Namespace: "default",
				},
				Spec: namespacelabelv1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"k8s.io": "test-invalid",
					},
				},
			}
			Expect(k8sClient.Create(ctx, invalidResource)).To(Succeed())

			By("verifying the label was not applied to the namespace")
			invalidNamespace := &corev1.Namespace{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: invalidResource.Name, Namespace: invalidResource.Namespace}, invalidResource)
				if errors.IsNotFound(err) {
					return false
				}
				err = k8sClient.Get(ctx, types.NamespacedName{Name: "default"}, invalidNamespace)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Eventually(func() map[string]string {
				return invalidNamespace.Labels
			}, timeout, interval).ShouldNot(HaveKeyWithValue("k8s.io", "test-invalid"), "protected label should not have been applied to the namespace")
			Eventually(func() bool {
				err := k8sClient.Delete(ctx, invalidResource)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})
	})
})
