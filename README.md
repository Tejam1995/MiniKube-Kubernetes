# Minikube

[![Build Status](https://travis-ci.org/kubernetes/minikube.svg?branch=master)](https://travis-ci.org/kubernetes/minikube)
[![codecov](https://codecov.io/gh/kubernetes/minikube/branch/master/graph/badge.svg)](https://codecov.io/gh/kubernetes/minikube)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bhttps%3A%2F%2Fgithub.com%2Fkubernetes%2Fminikube.svg?size=small)](https://app.fossa.io/projects/git%2Bhttps%3A%2F%2Fgithub.com%2Fkubernetes%2Fminikube?ref=badge_small)

<img src="https://github.com/kubernetes/minikube/raw/master/logo/logo.png" width="100">

## What is Minikube?

Minikube is a tool that makes it easy to run Kubernetes locally. Minikube runs a single-node Kubernetes cluster inside a VM on your laptop for users looking to try out Kubernetes or develop with it day-to-day.

## Installation
### macOS
```shell
brew cask install minikube
```

### Linux
```shell
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube && sudo mv minikube /usr/local/bin/
```

### Windows
Download the [minikube-windows-amd64.exe](https://storage.googleapis.com/minikube/releases/latest/minikube-windows-amd64.exe) file, rename it to `minikube.exe` and add it to your path

### Other ways to install:

* [Linux] [Arch Linux AUR](https://aur.archlinux.org/packages/minikube/)
* [Windows] [Chocolatey](https://chocolatey.org/packages/Minikube)

We also released a Debian package and Windows installer on our [releases page](https://github.com/kubernetes/minikube/releases)
If you maintain a minikube package, please feel free to add it here.

### Requirements
* [kubectl](https://kubernetes.io/docs/tasks/kubectl/install/)
* macOS
    * [xhyve driver](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#xhyve-driver), [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [VMware Fusion](https://www.vmware.com/products/fusion)
* Linux
    * [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [KVM](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#kvm-driver)
* Windows
    * [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [Hyper-V](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperV-driver)
* VT-x/AMD-v virtualization must be enabled in BIOS
* Internet connection on first run


## Quickstart

Here's a brief demo of minikube usage.
If you want to change the VM driver add the appropriate `--vm-driver=xxx` flag to `minikube start`. Minikube Supports
the following drivers:

* virtualbox
* vmwarefusion
* [KVM](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#kvm-driver)
* [xhyve](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#xhyve-driver)
* [Hyper-V](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperV-driver)


```shell
$ minikube start
Starting local Kubernetes v1.6.0 cluster...
Starting VM...
SSH-ing files into VM...
Setting up certs...
Starting cluster components...
Connecting to cluster...
Setting up kubeconfig...
Kubectl is now configured to use the cluster.

$ kubectl run hello-minikube --image=gcr.io/google_containers/echoserver:1.4 --port=8080
deployment "hello-minikube" created
$ kubectl expose deployment hello-minikube --type=NodePort
service "hello-minikube" exposed

# We have now launched an echoserver pod but we have to wait until the pod is up before curling/accessing it
# via the exposed service.
# To check whether the pod is up and running we can use the following:
$ kubectl get pod
NAME                              READY     STATUS              RESTARTS   AGE
hello-minikube-3383150820-vctvh   1/1       ContainerCreating   0          3s
# We can see that the pod is still being created from the ContainerCreating status
$ kubectl get pod
NAME                              READY     STATUS    RESTARTS   AGE
hello-minikube-3383150820-vctvh   1/1       Running   0          13s
# We can see that the pod is now Running and we will now be able to curl it:
$ curl $(minikube service hello-minikube --url)
CLIENT VALUES:
client_address=192.168.99.1
command=GET
real path=/
...
$ minikube stop
Stopping local Kubernetes cluster...
Machine stopped.
```

## Interacting With your Cluster

### Kubectl

The `minikube start` command creates a "[kubectl context](https://kubernetes.io/docs/user-guide/kubectl/v1.6/#-em-set-context-em-)" called "minikube".
This context contains the configuration to communicate with your minikube cluster.

Minikube sets this context to default automatically, but if you need to switch back to it in the future, run:

`kubectl config use-context minikube`,

or pass the context on each command like this: `kubectl get pods --context=minikube`.

### Dashboard

To access the [Kubernetes Dashboard](http://kubernetes.io/docs/user-guide/ui/), run this command in a shell after starting minikube to get the address:
```shell
minikube dashboard
```

### Services

To access a service exposed via a node port, run this command in a shell after starting minikube to get the address:
```shell
minikube service [-n NAMESPACE] [--url] NAME
```

## Design

Minikube uses [libmachine](https://github.com/docker/machine/tree/master/libmachine) for provisioning VMs, and [localkube](https://github.com/kubernetes/minikube/tree/master/pkg/localkube) (originally written and donated to this project by [RedSpread](https://redspread.com/)) for running the cluster.

For more information about minikube, see the [proposal](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/local-cluster-ux.md).

## Additional Links:
* [**Advanced Topics and Tutorials**](https://github.com/kubernetes/minikube/blob/master/docs/README.md)
* [**Contributing**](https://github.com/kubernetes/minikube/blob/master/CONTRIBUTING.md)
* [**Development Guide**](https://github.com/kubernetes/minikube/blob/master/docs/contributors/README.md)

## Community

* [**#minikube on Kubernetes Slack**](https://kubernetes.slack.com)
* [**kubernetes-dev mailing list** ](https://groups.google.com/forum/#!forum/kubernetes-dev)
(If you are posting to the list please prefix your subject with "minikube: ")
