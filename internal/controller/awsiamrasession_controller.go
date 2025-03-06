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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
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

	k, err := kubernetes.NewForConfig(r.KubeConfig)
	if err != nil {
		logger.Error(err, "unable to create client")
		return ctrl.Result{}, err
	}

	podList, err := k.CoreV1().Pods(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Error(err, "unable to query API for pods")
		return ctrl.Result{}, err
	}

	var relatedPods []corev1.Pod
	var relatedPodNames []string
	for _, pod := range podList.Items {
		if metav1.HasAnnotation(pod.ObjectMeta, v1.SessionNamePodAnnotationKey) {
			if pod.Annotations[v1.SessionNamePodAnnotationKey] == session.Name {
				relatedPods = append(relatedPods, pod)
				relatedPodNames = append(relatedPodNames, pod.Name)
			}
		}
	}

	logger.Info("Found pods using this session", "pods", relatedPodNames)

	// TODO: update config for all pods

	// TODO: emit events and/or conditions

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AwsIamRaSessionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.AwsIamRaSession{}).
		Named("awsiamrasession").
		Complete(r)
}
