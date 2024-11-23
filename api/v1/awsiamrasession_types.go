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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ARN string

// AwsIamRaSessionSpec defines the desired state of AwsIamRaSession.
type AwsIamRaSessionSpec struct {
	// TODO: add annotated comments
	Region         string `json:"region,omitempty"`
	CertSecret     string `json:"certSecret,omitempty"`
	TrustAnchorArn ARN    `json:"trustAnchorArn,omitempty"`
	ProfileArn     ARN    `json:"profileArn,omitempty"`
	RoleArn        ARN    `json:"roleArn,omitempty"`
}

// AwsIamRaSessionStatus defines the observed state of AwsIamRaSession.
type AwsIamRaSessionStatus struct {
	ExpirationTime metav1.Time `json:"expirationTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AwsIamRaSession is the Schema for the awsiamrasessions API.
type AwsIamRaSession struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsIamRaSessionSpec   `json:"spec,omitempty"`
	Status AwsIamRaSessionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AwsIamRaSessionList contains a list of AwsIamRaSession.
type AwsIamRaSessionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsIamRaSession `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsIamRaSession{}, &AwsIamRaSessionList{})
}
