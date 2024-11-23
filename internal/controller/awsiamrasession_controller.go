/*
Copyright 2024.

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudv1 "dancav.io/aws-iamra-manager/api/v1"
)

// AwsIamRaSessionReconciler reconciles a AwsIamRaSession object
type AwsIamRaSessionReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=cloud.dancav.io,resources=awsiamrasessions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.dancav.io,resources=awsiamrasessions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.dancav.io,resources=awsiamrasessions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *AwsIamRaSessionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var session cloudv1.AwsIamRaSession
	if err := r.Get(ctx, req.NamespacedName, &session); err != nil {
		logger.Error(err, "unable to fetch AwsIamRaSession")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	certSecretName := session.Spec.CertSecret
	certSecretRef := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      certSecretName,
	}
	var certSecret v1.Secret
	if err := r.Get(ctx, certSecretRef, &certSecret); err != nil {
		logger.Error(err, "unable to fetch certificate secret",
			"secretName", certSecretName)
		r.Recorder.Eventf(&session, v1.EventTypeWarning, "Failed",
			"Certificate secret \"%s\" does not exist", certSecretName)
		return ctrl.Result{}, err
	}

	// TODO: issue creds and do something with them
	// TODO: emit more useful events, including success

	session.Status.ExpirationTime = metav1.Now()
	if err := r.Status().Update(ctx, &session); err != nil {
		logger.Error(err, "unable to update AwsIamRaSession status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AwsIamRaSessionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudv1.AwsIamRaSession{}).
		Named("awsiamrasession").
		Complete(r)
}
