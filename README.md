# Minikube

Run Kubernetes locally

[![Build Status](https://travis-ci.org/kubernetes/minikube.svg?branch=master)](https://travis-ci.org/kubernetes/minikube)

## Background

Minikube is a tool that makes it easy to run Kubernetes locally. Minikube runs
a single-node Kubernetes cluster inside a VM on your laptop for users looking
to try out Kubernetes or develop with it day-to-day.

## Design

Minikube uses [libmachine](https://github.com/docker/machine/tree/master/libmachine) for provisioning VMs, and [localkube](https://github.com/redspread/localkube)
for running the cluster.

For more information about minikube, see the [proposal](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/local-cluster-ux.md).

## Goals

* Works across multiple OSes - OS X, Linux and Windows primarily.
* Single command setup and teardown UX.
* Unified UX across OSes
* Minimal dependencies on third party software.
* Minimal resource overhead.
* Replace any other alternatives to local cluster deployment.

## Non Goals

* Simplifying kubernetes production deployment experience. Kube-deploy is attempting to tackle this problem.
* Supporting all possible deployment configurations of Kubernetes like various types of storage, networking, etc.

## Build Requirements

* A recent Go distribution (>1.6)
* If you're not on Linux, you'll need a Docker installation

## Build Instructions

```shell
make out/minikube
```

## Requirements For Running Minikube
* VirtualBox installation
* VT-x/AMD-v virtualization must be enabled in BIOS

## Run Instructions

Start the cluster with:

```console
$ ./out/minikube start
Starting local Kubernetes cluster...
2016/04/19 11:41:26 Machine exists!
2016/04/19 11:41:27 Kubernetes is available at https://192.168.99.100:443.
2016/04/19 11:41:27 Run this command to use the cluster: 
2016/04/19 11:41:27 kubectl config set-cluster minikube --server=https://192.168.99.100:443 --certificate-authority=$HOME/.minikube/apiserver.crt
```

Access the cluster by adding `-s=https://192.168.x.x:443` to every `kubectl` command, or run the commands below:

```shell
kubectl config set-cluster minikube --server=https://192.168.99.100:443 --certificate-authority=$HOME/.minikube/apiserver.crt
kubectl config set-context minikube --cluster=minikube
kubectl config use-context minikube
```

by running those commands, you may use `kubectl` normally

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for an overview of how to send pull requests.

### Running Tests

#### Unit Tests

Unit tests are run on Travis before code is merged. To run as part of a development cycle:

```shell
make test
```

#### Integration Tests

Integration tests are currently run manually. 
To run them, build the binary and run the tests:

```shell
make integration
```

#### Conformance Tests

These are kubernetes tests that run against an arbitrary cluster and exercise a wide range of kubernetes features.
You can run these against minikube by following these steps:

* Clone the kubernetes repo somewhere on your system.
* Run `make quick-release` in the k8s repo.
* Start up a minikube cluster with: `minikube start`.
* Set these two environment variables:
```shell
export KUBECONFIG=$HOME/.kube/config
export KUBERNETES_CONFORMANCE_TEST=y
```
* Run the tests (from the k8s repo):
```shell
go run hack/e2e.go -v --test --test_args="--ginkgo.focus=\[Conformance\]" --check_version_skew=false --check_node_count=false
```
