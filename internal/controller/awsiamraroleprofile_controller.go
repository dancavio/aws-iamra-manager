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
	"dancav.io/aws-iamra-manager/internal/iamram"
	"errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AwsIamRaRoleProfileReconciler reconciles a AwsIamRaRoleProfile object
type AwsIamRaRoleProfileReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Recorder   record.EventRecorder
	KubeConfig *rest.Config
}

// +kubebuilder:rbac:groups=cloud.dancav.io,resources=awsiamraroleprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.dancav.io,resources=awsiamraroleprofiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.dancav.io,resources=awsiamraroleprofiles/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=list;watch;get
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *AwsIamRaRoleProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Received reconcile request for AwsIamRaRoleProfile")

	var profile v1.AwsIamRaRoleProfile
	if err := r.Get(ctx, req.NamespacedName, &profile); err != nil {
		logger.Info("unable to fetch AwsIamRaRoleProfile")
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

	var updatablePods []corev1.Pod
	var updatablePodNames []string
	for _, pod := range podList.Items {
		// TODO: might need to requeue if any pods are pending?
		if podNeedsUpdate(pod, profile) {
			updatablePods = append(updatablePods, pod)
			updatablePodNames = append(updatablePodNames, types.NamespacedName{
				Namespace: pod.Namespace,
				Name:      pod.Name,
			}.String())
		}
	}

	logger.Info("Found pods using this profile", "pods", updatablePodNames)
	profile.Status.ActivePods = updatablePodNames
	if err := r.Status().Update(ctx, &profile); err != nil {
		logger.Error(err, "unable to update AwsIamRaRoleProfile status")
		return ctrl.Result{}, err
	}

	anyFailures := false
	for _, pod := range updatablePods {
		logger.Info("Updating config for pod", "pod", pod.Name)
		if err := iamram.ReconcilePod(ctx, k, r.KubeConfig, &profile, pod); err != nil {
			anyFailures = true
		}
	}

	var finalError error
	if anyFailures {
		finalError = errors.New("failed to update one or more pods")
	}
	return ctrl.Result{}, finalError
}

func podNeedsUpdate(pod corev1.Pod, profile v1.AwsIamRaRoleProfile) bool {
	return metav1.HasAnnotation(pod.ObjectMeta, v1.RoleProfilePodAnnotationKey) &&
		pod.Annotations[v1.RoleProfilePodAnnotationKey] == profile.Name &&
		pod.Status.Phase != corev1.PodFailed && pod.Status.Phase != corev1.PodSucceeded
}

// SetupWithManager sets up the controller with the Manager.
func (r *AwsIamRaRoleProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.AwsIamRaRoleProfile{}).
		Named("awsiamraroleprofile").
		Complete(r)
}
