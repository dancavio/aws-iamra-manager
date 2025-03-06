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
	"fmt"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"dancav.io/aws-iamra-manager/api/v1"
)

const defaultSessionDurationSeconds = 3600

// SetupAwsIamRaRoleProfileWebhookWithManager registers the webhook for AwsIamRaRoleProfile in the manager.
func SetupAwsIamRaRoleProfileWebhookWithManager(mgr ctrl.Manager) error {
	logger := logf.Log.WithName("awsiamraroleprofile-webhook")

	return ctrl.NewWebhookManagedBy(mgr).For(&v1.AwsIamRaRoleProfile{}).
		WithValidator(&AwsIamRaRoleProfileCustomValidator{logger}).
		WithDefaulter(&AwsIamRaRoleProfileCustomDefaulter{logger}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-cloud-dancav-io-v1-awsiamraroleprofile,mutating=true,failurePolicy=fail,sideEffects=NoneOnDryRun,groups=cloud.dancav.io,resources=awsiamraroleprofiles,verbs=create;update,versions=v1,name=mawsiamraroleprofile-v1.kb.io,admissionReviewVersions=v1

// AwsIamRaRoleProfileCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind AwsIamRaRoleProfile when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type AwsIamRaRoleProfileCustomDefaulter struct {
	logger logr.Logger
}

var _ webhook.CustomDefaulter = &AwsIamRaRoleProfileCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind AwsIamRaRoleProfile.
func (d *AwsIamRaRoleProfileCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	profile, ok := obj.(*v1.AwsIamRaRoleProfile)
	if !ok {
		return fmt.Errorf("expected an AwsIamRaRoleProfile object but got %T", obj)
	}
	d.logger.Info("Setting defaults for AwsIamRaRoleProfile", "name", profile.GetName())

	if profile.Spec.DurationSeconds == 0 {
		profile.Spec.DurationSeconds = defaultSessionDurationSeconds
	}

	return nil
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-cloud-dancav-io-v1-awsiamraroleprofile,mutating=false,failurePolicy=fail,sideEffects=None,groups=cloud.dancav.io,resources=awsiamraroleprofiles,verbs=create;update,versions=v1,name=vawsiamraroleprofile-v1.kb.io,admissionReviewVersions=v1

// AwsIamRaRoleProfileCustomValidator struct is responsible for validating the AwsIamRaRoleProfile resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type AwsIamRaRoleProfileCustomValidator struct {
	logger logr.Logger
}

var _ webhook.CustomValidator = &AwsIamRaRoleProfileCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type AwsIamRaRoleProfile.
func (v *AwsIamRaRoleProfileCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	profile, ok := obj.(*v1.AwsIamRaRoleProfile)
	if !ok {
		return nil, fmt.Errorf("expected an AwsIamRaRoleProfile object but got %T", obj)
	}
	v.logger.Info("Performing creation validation for AwsIamRaRoleProfile", "name", profile.GetName())
	return validateProfile(ctx, profile)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type AwsIamRaRoleProfile.
func (v *AwsIamRaRoleProfileCustomValidator) ValidateUpdate(ctx context.Context, _, newObj runtime.Object) (admission.Warnings, error) {
	profile, ok := newObj.(*v1.AwsIamRaRoleProfile)
	if !ok {
		return nil, fmt.Errorf("expected an AwsIamRaRoleProfile object for the newObj but got %T", newObj)
	}
	v.logger.Info("Performing update validation for AwsIamRaRoleProfile", "name", profile.GetName())
	return validateProfile(ctx, profile)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type AwsIamRaRoleProfile.
func (v *AwsIamRaRoleProfileCustomValidator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*v1.AwsIamRaRoleProfile)
	if !ok {
		return nil, fmt.Errorf("expected an AwsIamRaRoleProfile object but got %T", obj)
	}
	return nil, nil
}

func validateProfile(_ context.Context, profile *v1.AwsIamRaRoleProfile) (admission.Warnings, error) {
	taRegion, taErr := validateARN(field.NewPath("spec").Child("trustAnchorArn"), profile.Spec.TrustAnchorArn)
	profRegion, profErr := validateARN(field.NewPath("spec").Child("profileArn"), profile.Spec.ProfileArn)
	_, roleErr := validateARN(field.NewPath("spec").Child("roleArn"), profile.Spec.RoleArn)
	var allErrs []*field.Error
	for _, err := range []*field.Error{taErr, profErr, roleErr} {
		if err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if taRegion != "" && profRegion != "" && taRegion != profRegion {
		err := field.Invalid(field.NewPath("spec"), profile.Spec,
			"trust anchor and profile ARN regions must match")
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(v1.AwsIamRaRoleProfileGroupKind, profile.Name, allErrs)
}

func validateARN(path *field.Path, arn v1.ARN) (string, *field.Error) {
	if !arn.IsValid() {
		return "", field.Invalid(path, arn, "must be a valid ARN")
	}
	awsarn, err := arn.Parse()
	if err != nil {
		return "", field.Invalid(path, arn, err.Error())
	}
	return awsarn.Region, nil
}
