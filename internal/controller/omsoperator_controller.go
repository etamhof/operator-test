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
	"io"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	whatnotv1alpha1 "something.com/my/http-op/api/v1alpha1"
)

// OmsOperatorReconciler reconciles a OmsOperator object
type OmsOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var endpoints = [...]string{
	"/o2ims-infrastructureInventory/api_versions",
	"/o2ims-infrastructureInventory/v1",
	"/o2ims-infrastructureInventory/v1/api_versions",
	"/o2ims-infrastructureInventory/v1/deploymentManagers",
	"/o2ims-infrastructureInventory/v1/deploymentManagers/id",
}

const urlPrefix = "http://localhost:9000"

// +kubebuilder:rbac:groups=whatnot.etamhof,resources=omsoperators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=whatnot.etamhof,resources=omsoperators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=whatnot.etamhof,resources=omsoperators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OmsOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *OmsOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconcile")

	omsoperator := &whatnotv1alpha1.OmsOperator{}
	err := r.Get(ctx, req.NamespacedName, omsoperator)
	cond := meta.FindStatusCondition(omsoperator.Status.Conditions, "Done")
	if cond != nil {
		log.Info("Already finished")
		return ctrl.Result{}, nil
	}
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get crd")
		return ctrl.Result{}, err
	}

	url := urlPrefix + endpoints[omsoperator.Spec.EndPoint]
	resp, err := http.Get(url)
	if err != nil {
		log.Error(err, "Failed to get url "+url)

		meta.SetStatusCondition(&omsoperator.Status.Conditions, metav1.Condition{Type: "Done",
			Status: metav1.ConditionTrue, Reason: "Failed",
			Message: "Done with error: " + err.Error()})

		if err := r.Status().Update(ctx, omsoperator); err != nil {
			log.Error(err, "Failed to update crd status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	bodybytes, _ := io.ReadAll(resp.Body)
	body := string(bodybytes[:])
	log.Info("GET URL: "+url, "Status", resp.Status, "Body", body)

	meta.SetStatusCondition(&omsoperator.Status.Conditions, metav1.Condition{Type: "Done",
		Status: metav1.ConditionTrue, Reason: "Done",
		Message: "Done: " + "HTTP " + resp.Status + " - " + body})

	if err := r.Status().Update(ctx, omsoperator); err != nil {
		log.Error(err, "Failed to update crd status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OmsOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&whatnotv1alpha1.OmsOperator{}).
		Complete(r)
}
