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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"dancav.io/aws-iamra-manager/api/v1"
)

var _ = Describe("AwsIamRaRoleProfile Webhook", func() {
	var (
		obj       *v1.AwsIamRaRoleProfile
		oldObj    *v1.AwsIamRaRoleProfile
		validator AwsIamRaRoleProfileCustomValidator
		defaulter AwsIamRaRoleProfileCustomDefaulter
	)

	BeforeEach(func() {
		obj = &v1.AwsIamRaRoleProfile{}
		oldObj = &v1.AwsIamRaRoleProfile{}
		validator = AwsIamRaRoleProfileCustomValidator{}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		defaulter = AwsIamRaRoleProfileCustomDefaulter{}
		Expect(defaulter).NotTo(BeNil(), "Expected defaulter to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
		// TODO (user): Add any setup logic common to all tests
	})

	AfterEach(func() {
		// TODO (user): Add any teardown logic common to all tests
	})

	Context("When creating AwsIamRaRoleProfile under Defaulting Webhook", func() {
		// TODO (user): Add logic for defaulting webhooks
		// Example:
		// It("Should apply defaults when a required field is empty", func() {
		//     By("simulating a scenario where defaults should be applied")
		//     obj.SomeFieldWithDefault = ""
		//     By("calling the Default method to apply defaults")
		//     defaulter.Default(ctx, obj)
		//     By("checking that the default values are set")
		//     Expect(obj.SomeFieldWithDefault).To(Equal("default_value"))
		// })
	})

	Context("When creating or updating AwsIamRaRoleProfile under Validating Webhook", func() {
		// TODO (user): Add logic for validating webhooks

		It("Should admit creation if all required fields are present", func() {
			By("simulating a valid creation scenario")
			obj.Spec.TrustAnchorArn = "arn:aws:rolesanywhere:us-east-1:123:trust-anchor/foo"
			obj.Spec.ProfileArn = "arn:aws:rolesanywhere:us-east-1:123:profile/bar"
			obj.Spec.RoleArn = "arn:aws:iam::123:role/baz"
			Expect(validator.ValidateCreate(ctx, obj)).To(BeNil())
		})

		It("Should deny creation if ARN regions don't match", func() {
			By("simulating an invalid creation scenario")
			obj.Spec.TrustAnchorArn = "arn:aws:rolesanywhere:us-east-1:123:trust-anchor/foo"
			obj.Spec.ProfileArn = "arn:aws:rolesanywhere:us-west-2:123:profile/bar"
			obj.Spec.RoleArn = "arn:aws:iam::123:role/baz"
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should deny creation if ARNs are invalid", func() {
			By("simulating an invalid creation scenario")
			obj.Spec.TrustAnchorArn = "arn:aws:rolesanywhere:us-east-1:123:trust-anchor/foo"
			obj.Spec.ProfileArn = "arn:aws:rolesanywhere:us-west-1:123:profile/bar"
			obj.Spec.RoleArn = "baz"
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})
	})

})
