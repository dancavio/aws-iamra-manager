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
	aws "github.com/aws/aws-sdk-go-v2/aws/arn"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RoleProfilePodAnnotationKey = "cloud.dancav.io/aws-iamra-role-profile"
	CertSecretPodAnnotationKey  = "cloud.dancav.io/aws-iamra-cert-secret"
)

type ARN string

// AwsIamRaRoleProfileSpec defines the desired state of AwsIamRaRoleProfile.
type AwsIamRaRoleProfileSpec struct {
	// TODO: add more annotated comments

	// +kubebuilder:validation:Required
	TrustAnchorArn ARN `json:"trustAnchorArn,omitempty"`
	// +kubebuilder:validation:Required
	ProfileArn ARN `json:"profileArn,omitempty"`
	// +kubebuilder:validation:Required
	RoleArn ARN `json:"roleArn,omitempty"`

	DurationSeconds int32  `json:"durationSeconds,omitempty"`
	RoleSessionName string `json:"roleSessionName,omitempty"`
}

func (arn ARN) IsValid() bool {
	return aws.IsARN(string(arn))
}

func (arn ARN) Parse() (aws.ARN, error) {
	return aws.Parse(string(arn))
}

// TODO: use status conditions and add printable columns (e.g. expiration)
// Some docs: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
// Example: https://github.com/fluxcd/kustomize-controller/blob/8ba6b2028f121c7986aeeee84dd2db0cd5d1a685/api/v1/kustomization_types.go#L294-L295

// AwsIamRaRoleProfileStatus defines the observed state of AwsIamRaRoleProfile.
type AwsIamRaRoleProfileStatus struct {
	// TODO: don't need this anymore; maybe list of pods using session?
	ExpirationTimes map[string]metav1.Time `json:"expirationTimes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AwsIamRaRoleProfile is the Schema for the awsIamRaRoleProfiles API.
type AwsIamRaRoleProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsIamRaRoleProfileSpec   `json:"spec,omitempty"`
	Status AwsIamRaRoleProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AwsIamRaRoleProfileList contains a list of AwsIamRaRoleProfile.
type AwsIamRaRoleProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsIamRaRoleProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsIamRaRoleProfile{}, &AwsIamRaRoleProfileList{})
}
