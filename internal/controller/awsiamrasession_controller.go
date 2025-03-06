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
	"dancav.io/aws-iamra-manager/api/v1"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AwsIamRaSessionReconciler reconciles a AwsIamRaSession object
type AwsIamRaSessionReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Recorder   record.EventRecorder
	KubeConfig *rest.Config
}

const (
	reasonInactive = "Inactive"
	reasonUpdated  = "Updated"
)

// +kubebuilder:rbac:groups=cloud.dancav.io,resources=awsiamrasessions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.dancav.io,resources=awsiamrasessions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.dancav.io,resources=awsiamrasessions/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=list;watch;get
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *AwsIamRaSessionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Received reconcile request for AwsIamRaSession")

	var session v1.AwsIamRaSession
	if err := r.Get(ctx, req.NamespacedName, &session); err != nil {
		logger.Info("unable to fetch AwsIamRaSession")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO: List all pods with annotation "cloud.dancav.io/aws-iamra-session-name" matching this session's name

	k, err := kubernetes.NewForConfig(r.KubeConfig)
	if err != nil {
		logger.Error(err, "unable to create clientset")
		return ctrl.Result{}, err
	}

	listOps := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.annotations.%s=%s", v1.SessionNamePodAnnotationKey, session.Name),
	}
	podList, err := k.CoreV1().Pods(req.Namespace).List(ctx, listOps)
	if err != nil {
		logger.Error(err, "unable to query API for pods")
		return ctrl.Result{}, err
	}

	logger.Info("Found pods using this session", "pods", podList.Items)

	// TODO: not much to reconcile anymore.. maybe keep track of the pods using the session (in its status),
	// and here we just check if cleanup is necessary? update existing pods somehow if session params change?

	//k, err := kubernetes.NewForConfig(r.KubeConfig)
	//if err != nil {
	//	logger.Error(err, "unable to create clientset")
	//	return ctrl.Result{}, err
	//}
	//
	//listOps := metav1.ListOptions{
	//	LabelSelector: fmt.Sprintf("%s=%s", v1.SessionNamePodLabelKey, session.Name),
	//}
	//podList, err := k.CoreV1().Pods(req.Namespace).List(ctx, listOps)
	//if err != nil {
	//	logger.Error(err, "unable to query API for pods")
	//	return ctrl.Result{}, err
	//}
	//
	//if len(podList.Items) == 0 {
	//	r.Recorder.Event(&session, corev1.EventTypeWarning, reasonInactive,
	//		"Found no pods matching selector")
	//	return ctrl.Result{}, nil
	//}
	//
	//var nextRequeue *time.Time
	//credsRefreshed := false
	//for _, pod := range podList.Items {
	//	expirationForPod, refreshedPod, err := iamram.ReconcilePod(ctx, k, r.KubeConfig, &session, pod)
	//	if err != nil {
	//		return ctrl.Result{}, fmt.Errorf(
	//			"couldn't reconcile credentials for pod %s: %w", pod.Name, err)
	//	}
	//	// Use the first pod's requeue time, since it's the earliest.
	//	if nextRequeue == nil {
	//		nextRequeueForPod := expirationForPod.Add(-1 * iamram.ExpirationBufferSeconds * time.Second)
	//		nextRequeue = &nextRequeueForPod
	//	}
	//	if refreshedPod {
	//		credsRefreshed = true
	//	}
	//}
	//
	//if err := r.Status().Update(ctx, &session); err != nil {
	//	logger.Error(err, "unable to update AwsIamRaSession status")
	//	return ctrl.Result{}, err
	//}
	//
	//// TODO: emit more useful events, including success and all failure cases
	//// TODO: how do other k8s controllers do error-handling/logging/eventing?
	//
	//if credsRefreshed {
	//	r.Recorder.Event(&session, corev1.EventTypeNormal, reasonUpdated,
	//		"Successfully issued new session credentials")
	//}

	return ctrl.Result{RequeueAfter: time.Hour}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AwsIamRaSessionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.AwsIamRaSession{}).
		Named("awsiamrasession").
		Complete(r)
}
