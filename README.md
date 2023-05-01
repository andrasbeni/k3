# k3
A kubernetes operator built with kubebuilder that deploys an application of your choice (e.g. nginx "It works!") and
operates it on top of Kubernetes using a custom resource.

## Description
A `k3` `Spec` consist of (i.e. user of this operator can specify):
* replicas : number of replicas
* host : host where the application is accessible (ie. in the browser)
* image : container image (and tag) (eg. nginx:latest )

By specifying the above three options, a user can launch the application
and open in it the browser using HTTPS.

### Components installed by the operator

The operator creates a `Deployment` resource that manages `replicas` number of `Pod`s based on the `image`. This deployment's name is derived from the K3 resource's name by adding the suffix `-deployment`.
Similarly, a `Service` resource is created to expose the `Pod`s in the previously mentioned `Deployment`. The name suffix for this `Service` is `-service`.
Last but not least, an `Ingress` resource is also created to allow external access to the `Service`. The name of the `Ingress` is suffexd with `-ingress`. For this ingress to work, you need to intall additional resources (see the next section). 


### Components required by the operator

The k3 operator expects the following resources to be installed:

- And `Ingress Controller` (e.g. ngnix)
- An `Ingress Class` that contains the configuration for the `Ingress Controller` that implements the `Ingress`. It's name must be the concatenation of the `k3` resource's name and the suffix `-ingress-class`.
- A `cert-manager` that will manage certificates for the `Ingress Controller`.
- An `Issuer` that takes care of requesting and renewing TLS certificates. 

Part of the task is uploading the project to GitHub (or a similar code hosting platform),
configure CI and publish container images to a registry automatically (we recommend
GitHub Actions and Container registry).
The README should contain installation instructions using the published container
image including the installation of the operator and any additional components that's
necessary for the operator to work.
Besides the above instructions there are no additional requirements or limitations.
Here are a few tips though:

Similarly to the above, Ingress is probably the easiest way to publish the
application (using an ingress controller, we recommend nginx)
For HTTPS cert-manager and Let's Encrypt became the industry standard.
Components (and their configuration) like nginx ingress and cert-manager can beprerequisites for your operator (ie. it doesn't have to install them), but include
instructions for setting them up in the README.


## Getting Started
You’ll need a Kubernetes cluster to run against.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Prerequisites

Install Prerequisite resources (see Components required by the operator) according to https://cert-manager.io/docs/tutorials/acme/nginx-ingress/.

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Optional: Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/operator:tag
```

Or if you want to use the ltest image built by CI:
```sh
make deploy IMG=gcr.io/alert-vim-385406/operator:latest
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

