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
	// +kubebuilder:validation:Required
	TrustAnchorArn ARN `json:"trustAnchorArn,omitempty"`

	// +kubebuilder:validation:Required
	ProfileArn ARN `json:"profileArn,omitempty"`

	// +kubebuilder:validation:Required
	RoleArn ARN `json:"roleArn,omitempty"`

	// +kubebuilder:validation:Minimum=900
	// +kubebuilder:validation:Maximum=43200
	DurationSeconds int32 `json:"durationSeconds,omitempty"`

	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=64
	RoleSessionName string `json:"roleSessionName,omitempty"`
}

func (arn ARN) IsValid() bool {
	return aws.IsARN(string(arn))
}

func (arn ARN) Parse() (aws.ARN, error) {
	return aws.Parse(string(arn))
}

// AwsIamRaRoleProfileStatus defines the observed state of AwsIamRaRoleProfile.
type AwsIamRaRoleProfileStatus struct {
	ActivePods []string `json:"activePods,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="RoleArn",type=string,JSONPath=`.spec.roleArn`

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
