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

// log is for logging in this package.
var logger = logf.Log.WithName("awsiamrasession-webhook")

// SetupAwsIamRaSessionWebhookWithManager registers the webhook for AwsIamRaSession in the manager.
func SetupAwsIamRaSessionWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1.AwsIamRaSession{}).
		WithValidator(&AwsIamRaSessionCustomValidator{}).
		WithDefaulter(&AwsIamRaSessionCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-cloud-dancav-io-v1-awsiamrasession,mutating=true,failurePolicy=fail,sideEffects=None,groups=cloud.dancav.io,resources=awsiamrasessions,verbs=create;update,versions=v1,name=mawsiamrasession-v1.kb.io,admissionReviewVersions=v1

// AwsIamRaSessionCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind AwsIamRaSession when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type AwsIamRaSessionCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &AwsIamRaSessionCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind AwsIamRaSession.
func (d *AwsIamRaSessionCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	session, ok := obj.(*v1.AwsIamRaSession)
	if !ok {
		return fmt.Errorf("expected an AwsIamRaSession object but got %T", obj)
	}
	logger.Info("Setting defaults for AwsIamRaSession", "name", session.GetName())

	if session.Spec.DurationSeconds == 0 {
		session.Spec.DurationSeconds = defaultSessionDurationSeconds
	}
	// TODO: Support cluster/namespace defaults for profile and trust anchor

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-cloud-dancav-io-v1-awsiamrasession,mutating=false,failurePolicy=fail,sideEffects=None,groups=cloud.dancav.io,resources=awsiamrasessions,verbs=create;update,versions=v1,name=vawsiamrasession-v1.kb.io,admissionReviewVersions=v1

// AwsIamRaSessionCustomValidator struct is responsible for validating the AwsIamRaSession resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type AwsIamRaSessionCustomValidator struct {
	//TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &AwsIamRaSessionCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type AwsIamRaSession.
func (v *AwsIamRaSessionCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	session, ok := obj.(*v1.AwsIamRaSession)
	if !ok {
		return nil, fmt.Errorf("expected an AwsIamRaSession object but got %T", obj)
	}
	logger.Info("Performing creation validation for AwsIamRaSession", "name", session.GetName())
	return validateSession(ctx, session)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type AwsIamRaSession.
func (v *AwsIamRaSessionCustomValidator) ValidateUpdate(ctx context.Context, _, newObj runtime.Object) (admission.Warnings, error) {
	session, ok := newObj.(*v1.AwsIamRaSession)
	if !ok {
		return nil, fmt.Errorf("expected an AwsIamRaSession object for the newObj but got %T", newObj)
	}
	logger.Info("Performing update validation for AwsIamRaSession", "name", session.GetName())
	return validateSession(ctx, session)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type AwsIamRaSession.
func (v *AwsIamRaSessionCustomValidator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*v1.AwsIamRaSession)
	if !ok {
		return nil, fmt.Errorf("expected an AwsIamRaSession object but got %T", obj)
	}
	// TODO: Is there any validation to do here on deletion?
	return nil, nil
}

// TODO: Is this where I should do more of the validation that's in the controller, like existence of the cert secret?
func validateSession(_ context.Context, session *v1.AwsIamRaSession) (admission.Warnings, error) {
	taRegion, taErr := validateARN(field.NewPath("spec").Child("trustAnchorArn"), session.Spec.TrustAnchorArn)
	profRegion, profErr := validateARN(field.NewPath("spec").Child("profileArn"), session.Spec.ProfileArn)
	_, roleErr := validateARN(field.NewPath("spec").Child("roleArn"), session.Spec.RoleArn)
	var allErrs []*field.Error
	for _, err := range []*field.Error{taErr, profErr, roleErr} {
		if err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if taRegion != "" && profRegion != "" && taRegion != profRegion {
		err := field.Invalid(field.NewPath("spec"), session.Spec,
			"trust anchor and profile ARN regions must match")
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(v1.AwsIamRaSessionGroupKind, session.Name, allErrs)
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
