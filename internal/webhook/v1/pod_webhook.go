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
	certv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	certSecretVolumeName       = "aws-iamra-cert-secret"
	sidecarContainerImage      = "ghcr.io/dancavio/aws-iamra-manager/sidecar:0.1.0"
	sidecarContainerName       = "aws-iamra-manager"
	sidecarCredentialMountPath = "/iamram/certs"
	// TODO: make this configurable, but with a default
	defaultCredentialMountPath = "/root/.aws"
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
		return fmt.Errorf("expected an Pod object but got %T", obj)
	}

	podlog.Info("Defaulting for Pod", "name", pod.Name, "labels", pod.Labels)

	if sessionName, ok := pod.Labels[v1.SessionNamePodLabelKey]; ok {
		podlog.Info("Injecting AWS IAM RA session manager into new pod",
			"sessionName", v1.SessionNamePodLabelKey)
		return d.mutatePodSpec(ctx, pod, sessionName)
	}

	return nil
}

func (d *PodCustomDefaulter) mutatePodSpec(ctx context.Context, pod *corev1.Pod, sessionName string) error {
	err := d.client.Create(ctx, &certv1.Certificate{})
	if err != nil {
		return fmt.Errorf("error creating certificate: %w", err)
	}

	var certSecretName string
	var ok bool
	if certSecretName, ok = pod.Labels[v1.CertSecretPodLabelKey]; !ok {
		return fmt.Errorf("must specify label %s", v1.CertSecretPodLabelKey)
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes,
		corev1.Volume{
			Name: sessionName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: certSecretVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: certSecretName,
				},
			},
		},
	)
	// TODO: Support selecting a specific container
	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      sessionName,
			ReadOnly:  true,
			MountPath: defaultCredentialMountPath, // TODO: Allow this to be configurable
		})
		podlog.Info("Injected credential volume into container", "containerName", container.Name)
	}
	injectSidecar(pod, sessionName)

	return nil
}

func injectSidecar(pod *corev1.Pod, sessionName string) {
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{
		Name:          sidecarContainerName,
		Image:         sidecarContainerImage,
		RestartPolicy: &sidecarContainerRestartPolicy,
		Command:       []string{"sleep", "infinity"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      sessionName,
				ReadOnly:  false,
				MountPath: defaultCredentialMountPath,
			},
			{
				Name:      certSecretVolumeName,
				ReadOnly:  true,
				MountPath: sidecarCredentialMountPath,
			},
		},
	})
}
