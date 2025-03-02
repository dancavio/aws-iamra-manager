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

package v1

import (
	"context"
	"dancav.io/aws-iamra-manager/api/v1"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	certSecretVolumeName = "aws-iamra-cert-secret"
	// TODO: this should be configurable
	sidecarContainerImage = "ghcr.io/dancavio/aws-iamra-manager/sidecar:0.2.0"
	sidecarContainerName  = "aws-iamra-manager"
	sidecarCertMountPath  = "/iamram/certs"
)

var (
	sidecarContainerRestartPolicy = corev1.ContainerRestartPolicyAlways

	podlog = logf.Log.WithName("pod-webhook")
)

// SetupPodWebhookWithManager registers the webhook for Pod in the manager.
func SetupPodWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&corev1.Pod{}).
		WithDefaulter(&PodCustomDefaulter{
			client: mgr.GetClient(),
		}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate--v1-pod,mutating=true,failurePolicy=fail,sideEffects=NoneOnDryRun,groups="",resources=pods,verbs=create,versions=v1,name=mpod-v1.kb.io,admissionReviewVersions=v1

// PodCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Pod when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type PodCustomDefaulter struct {
	client client.Client
}

var _ webhook.CustomDefaulter = &PodCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Pod.
func (d *PodCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("expected a Pod object but got %T", obj)
	}

	podlog.Info("Defaulting for new pod")

	if sessionName, ok := pod.Annotations[v1.SessionNamePodAnnotationKey]; ok {
		podlog.Info("Injecting AWS IAM RA session manager into new pod",
			"sessionName", sessionName)
		return d.mutatePodSpec(ctx, pod, sessionName)
	}

	return nil
}

func (d *PodCustomDefaulter) mutatePodSpec(ctx context.Context, pod *corev1.Pod, sessionName string) error {
	var certSecretName string
	var ok bool
	if certSecretName, ok = pod.Annotations[v1.CertSecretPodAnnotationKey]; !ok {
		return fmt.Errorf("must specify annotation %s", v1.CertSecretPodAnnotationKey)
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes,
		corev1.Volume{
			Name: certSecretVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: certSecretName,
				},
			},
		},
	)
	//for i := range pod.Spec.Containers {
	//	container := &pod.Spec.Containers[i]
	//	// TODO: set the necessary AWS env vars
	//	//podlog.Info("Injected credential volume into container", "containerName", container.Name)
	//}

	return d.injectSidecar(ctx, pod, sessionName)
}

func (d *PodCustomDefaulter) injectSidecar(ctx context.Context, pod *corev1.Pod, sessionName string) error {
	sessionNsName := types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      sessionName,
	}
	var session v1.AwsIamRaSession
	if err := d.client.Get(ctx, sessionNsName, &session); err != nil {
		logger.Info("unable to fetch AwsIamRaSession")
		return err
	}
	logger.Info("found AwsIamRaSession object, injecting sidecar now")

	pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{
		Name:          sidecarContainerName,
		Image:         sidecarContainerImage,
		RestartPolicy: &sidecarContainerRestartPolicy,
		// TODO: set the update-credentials command
		Command: []string{"sleep", "infinity"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      certSecretVolumeName,
				ReadOnly:  true,
				MountPath: sidecarCertMountPath,
			},
		},
	})

	return nil
}
