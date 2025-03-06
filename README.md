# aws-iamra-manager

## TODOs

* Consider switching to the IMDS approach described here: https://aws.amazon.com/blogs/security/connect-your-on-premises-kubernetes-cluster-to-aws-apis-using-iam-roles-anywhere/

## Development notes

### kubebuilder init

To initialize project:

```shell
kubebuilder init --domain dancavallaro.com --repo dancavallaro.com/aws-iamra-manager
```

Initialize CRD:

```shell
kubebuilder create api --group cloud --version v1 --kind AwsIamRaRoleProfile
```

Scaffold defaulting and validating webhooks:

```shell
kubebuilder create webhook --group cloud --version v1 --kind AwsIamRaRoleProfile --defaulting --programmatic-validation
```

Pod injection webhook:

```shell
kubebuilder create webhook --group core --version v1 --kind Pod --defaulting
```

### Updating sidecar container

To build multi-platform images I first needed to create a customer builder:

```shell
docker buildx create \
  --name multiplatbuilder \
  --driver docker-container \
  --bootstrap --use
```

Then:

1. Update `release_version` in justfile to release a new version.
2. Enable multiplatform builder: `docker buildx use multiplatbuilder`
3. Then build and push to GitHub: `just build-multiplatform true`.
4To start using the new sidecar container, need to update the version in 
   `pod_webhook.go`, and release a new version of the controller.
   * TODO: The sidecar version should be configurable in the controller so it 
     doesn't require releasing a new build.

### Updating controller

1. Update the image version by updating the `IMG` variable in the Makefile.
2. Build the controller: `make all`
3. Build and publish multi-platform image: `make docker-buildx`
4. Update the install manifest (`dist/install.yaml`): `make build-installer`

## Getting Started

### Prerequisites
- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/aws-iamra-manager:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/aws-iamra-manager:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/dancavio/aws-iamra-manager/main/dist/install.yaml
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

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

