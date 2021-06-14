---
title: "FAQ"
linkTitle: "FAQ"
weight: 3
description: >
  Frequently Asked Questions
---


## Can I run an older Kubernetes version with minikube? Do I have to downgrade my minikube version?

You do not need to download an older minikube to run an older kubernetes version.
You can create a Kubernetes cluster with any version you desire using `--kubernetes-version` flag.

Example:

```bash
minikube start --kubernetes-version=v1.15.0
```


## Docker Driver: How can I set minikube's cgroup manager?

By default minikube uses the `cgroupfs` cgroup manager for Kubernetes clusters. If you are on a system with a systemd cgroup manager, this could cause conflicts.
To use the `systemd` cgroup manager, run:

```bash
minikube start --force-systemd=true
```

## How can I run minikube with the Docker driver if I have an existing cluster with a VM driver?

First please ensure your Docker service is running. Then you need to either:  

(a) Delete the existing cluster and create a new one

```bash
minikube delete
minikube start --driver=docker
```

Alternatively, (b) Create a second cluster with a different profile name:

```bash
minikube start -p p1 --driver=docker 
```

## Does minikube support IPv6?

minikube currently doesn't support IPv6. However, it is on the [roadmap]({{< ref "/docs/contrib/roadmap.en.md" >}}). You can also refer to the [open issue](https://github.com/kubernetes/minikube/issues/8535).

## How can I prevent password prompts on Linux?

The easiest approach is to use the `docker` driver, as the backend service always runs as `root`.

`none` users may want to try `CHANGE_MINIKUBE_NONE_USER=true`, where kubectl and such will work without `sudo`. See [environment variables]({{< ref "/docs/handbook/config.md#environment-variables" >}}) for more details.  

Alternatively, you can configure `sudo` to never prompt for commands issued by minikube.

## How can I ignore system verification?

[kubeadm](https://github.com/kubernetes/kubeadm), minikube's bootstrapper, verifies a list of features on the host system before installing Kubernetes. In the case you get an error and still want to try minikube despite your system's limitation, you can skip verification by starting minikube with this extra option:

```shell
minikube start --extra-config kubeadm.ignore-preflight-errors=SystemVerification
```

## What is the minimum resource allocation necessary for a Knative setup using minikube?

Please allocate sufficient resources for Knative setup using minikube, especially when running minikube cluster on your local machine. We recommend allocating at least 6 CPUs and 8G memory:

```shell
minikube start --cpus 6 --memory 8000
```

## Do I need to install kubectl locally?

No, minikube comes with a built-in kubectl installation. See [minikube's kubectl documentation]({{< ref "docs/handbook/kubectl.md" >}}).

## How can I opt-in to beta release notifications?

Simply run the following command to be enrolled into beta notifications:
```
minikube config set WantBetaUpdateNotification true
```

## Can I get rid of the emoji in minikube's outpuut?

Yes! If you prefer not having emoji in your minikube output 😔 , just set the `MINIKUBE_IN_STYLE` environment variable to `0` or `false`:

```
MINIKUBE_IN_STYLE=0 minikube start

```
