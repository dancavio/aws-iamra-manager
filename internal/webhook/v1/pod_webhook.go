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
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"strconv"
)

const (
	certSecretVolumeName        = "aws-iamra-cert-secret"
	sidecarContainerImageEnvVar = "AWS_IAMRA_MANAGER_SIDECAR_IMAGE"
	sidecarContainerName        = "aws-iamra-manager"
	sidecarCertMountPath        = "/iamram/certs"
)

var (
	sidecarContainerImage         string
	sidecarContainerRestartPolicy = corev1.ContainerRestartPolicyAlways
)

// SetupPodWebhookWithManager registers the webhook for Pod in the manager.
func SetupPodWebhookWithManager(mgr ctrl.Manager) error {
	var ok bool
	if sidecarContainerImage, ok = os.LookupEnv(sidecarContainerImageEnvVar); !ok {
		return fmt.Errorf("%s environment variable must be set", sidecarContainerImageEnvVar)
	}
	logger := logf.Log.WithName("pod-webhook")

	return ctrl.NewWebhookManagedBy(mgr).For(&corev1.Pod{}).
		WithDefaulter(&PodCustomDefaulter{
			client: mgr.GetClient(),
			logger: logger,
		}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate--v1-pod,mutating=true,reinvocationPolicy=IfNeeded,failurePolicy=fail,sideEffects=NoneOnDryRun,groups="",resources=pods,verbs=create,versions=v1,name=mpod-v1.kb.io,admissionReviewVersions=v1

// PodCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Pod when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type PodCustomDefaulter struct {
	client client.Client
	logger logr.Logger
}

var _ webhook.CustomDefaulter = &PodCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Pod.
func (d *PodCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("expected a Pod object but got %T", obj)
	}

	d.logger.Info("Defaulting for new pod", "pod", pod.GenerateName)

	if profileName, ok := pod.Annotations[v1.RoleProfilePodAnnotationKey]; ok {
		d.logger.Info("Injecting AWS IAM RA credential server into new pod",
			"profileName", profileName)
		return d.mutatePodSpec(ctx, pod, profileName)
	}

	return nil
}

func (d *PodCustomDefaulter) mutatePodSpec(ctx context.Context, pod *corev1.Pod, profileName string) error {
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

	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  "AWS_EC2_METADATA_SERVICE_ENDPOINT",
			Value: "http://127.0.0.1:9911/",
		})
	}

	return d.injectSidecar(ctx, pod, profileName)
}

func (d *PodCustomDefaulter) injectSidecar(ctx context.Context, pod *corev1.Pod, profileName string) error {
	profileNsName := types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      profileName,
	}
	var profile v1.AwsIamRaRoleProfile
	if err := d.client.Get(ctx, profileNsName, &profile); err != nil {
		d.logger.Info("unable to fetch AwsIamRaRoleProfile")
		return err
	}
	d.logger.Info("found AwsIamRaRoleProfile object, injecting sidecar now")

	command := []string{
		"serve-credentials",
		"-t", string(profile.Spec.TrustAnchorArn),
		"-p", string(profile.Spec.ProfileArn),
		"-r", string(profile.Spec.RoleArn),
		"-d", strconv.Itoa(int(profile.Spec.DurationSeconds)),
	}
	if profile.Spec.RoleSessionName != "" {
		command = append(command, "-n", profile.Spec.RoleSessionName)
	}

	d.logger.Info("creating sidecar container", "command", command)
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{
		Name:          sidecarContainerName,
		Image:         sidecarContainerImage,
		RestartPolicy: &sidecarContainerRestartPolicy,
		Command:       command,
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
