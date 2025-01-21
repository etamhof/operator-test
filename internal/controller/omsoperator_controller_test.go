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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	whatnotv1alpha1 "something.com/my/http-op/api/v1alpha1"
)

var _ = Describe("OmsOperator Controller", func() {
	FContext("When reconciling a resource", Ordered, func() {
		const resourceName = "test-resource"

		var ctx = context.Background()
		var resource *whatnotv1alpha1.OmsOperator

		var typeNamespacedName = types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}

		BeforeAll(func() {

		})

		BeforeEach(func() {
			By("creating the custom resource for the Kind OmsOperator")
			omsoperator := &whatnotv1alpha1.OmsOperator{}
			err := k8sClient.Get(ctx, typeNamespacedName, omsoperator)
			if err != nil && errors.IsNotFound(err) {
				resource = &whatnotv1alpha1.OmsOperator{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: whatnotv1alpha1.OmsOperatorSpec{
						EndPoint: 0,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &whatnotv1alpha1.OmsOperator{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Cleanup the specific resource instance OmsOperator")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &OmsOperatorReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			resource := &whatnotv1alpha1.OmsOperator{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			Expect(resource.Status.Conditions[0].Type).Should(Equal("Done"))
			Expect(resource.Status.Conditions[0].Message).Should(ContainSubstring("uriPrefix"))

			resource.Name = "xxx"
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should successfully reconcile the resource2", func() {
			By("Reconciling the created resource")
			resource.Spec.EndPoint = 1
			err := k8sClient.Update(ctx, resource)
			Expect(err).NotTo(HaveOccurred())
			controllerReconciler := &OmsOperatorReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			resource := &whatnotv1alpha1.OmsOperator{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			Expect(resource.Status.Conditions[0].Type).Should(Equal("Done"))
			Expect(resource.Status.Conditions[0].Message).Should(ContainSubstring("oCloudId"))

		})
	})
})
