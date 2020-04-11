/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

// Package kverify verifies a running kubernetes cluster is healthy
package kverify

import (
	"runtime"
	"time"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
)

// NodePressure verfies that node disks are healthy are not under pressure.
func NodePressure(cs *kubernetes.Clientset, cc config.ClusterConfig, timeout time.Duration) error {
	glog.Info("waiting to verify healty disk ...")
	start := time.Now()
	defer func() {
		glog.Infof("duration metric: took %s to wait for k8s-apps to be running ...", time.Since(start))
	}()

	ns, err := cs.CoreV1().Nodes().List(meta.ListOptions{})
	if err != nil {
		glog.Infof("failed to get nodes nodes: %v", err)
	}

	for _, n := range ns.Items {
		glog.Infof("node storage ephemeral capacity is %s", n.Status.Capacity.StorageEphemeral())
		glog.Infof("node cpu capacity is %s", n.Status.Capacity.Cpu().AsDec())
		for _, c := range n.Status.Conditions {
			if c.Type == v1.NodeDiskPressure && c.Status == v1.ConditionTrue {
				out.ErrT(out.FailureType, "node {{.name}} has unwanted condition {{.condition_type}} : Reason {{.reason}} Message: {{.message}}", out.V{"name": n.Name, "condition_type": c.Type, "reason": c.Reason, "message": c.Message})
				out.WarningT("The node on {{.name}} has ran out of disk space. please consider allocating more disk using or pruning un-used images", out.V{"name": n.Name})
				if driver.IsKIC(cc.Driver) && runtime.GOOS != "linux" {
					out.T(out.Tip, "Please increase Docker Desktop's disk image size.")
					if runtime.GOOS == "darwin" {
						out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
					}
					if runtime.GOOS == "windows" {
						out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
					}
				} else { // VM-drivers
					out.T(out.Tip, "You can specify a larger disk for your cluster using `minikube start --disk` ")
				}
			}

			if c.Type == v1.NodeMemoryPressure && c.Status == v1.ConditionTrue {
				out.ErrT(out.FailureType, "node {{.name}} has unwanted condition {{.condition_type}} : Reason {{.reason}} Message: {{.message}}", out.V{"name": n.Name, "condition_type": c.Type, "reason": c.Reason, "message": c.Message})
				out.WarningT("The node on {{.name}} has ran of memory.", out.V{"name": n.Name})
				if driver.IsKIC(cc.Driver) && runtime.GOOS != "linux" {
					out.T(out.Tip, "Please increase Docker Desktop's memory.")
					if runtime.GOOS == "darwin" {
						out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
					}
					if runtime.GOOS == "windows" {
						out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
					}
				} else {
					out.T(out.Tip, "You can specify a larger memory size for your cluster using `minikube start --memory` ")
				}
			}

			if c.Type == v1.NodeMemoryPressure && c.Status == v1.ConditionTrue {
				out.ErrT(out.FailureType, "node {{.name}} has unwanted condition {{.condition_type}} : Reason {{.reason}} Message: {{.message}}", out.V{"name": n.Name, "condition_type": c.Type, "reason": c.Reason, "message": c.Message})
				out.WarningT("The node on {{.name}} has ran of memory.", out.V{"name": n.Name})
				if driver.IsKIC(cc.Driver) && runtime.GOOS != "linux" {
					out.T(out.Tip, "Please increase Docker Desktop's memory.")
					if runtime.GOOS == "darwin" {
						out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
					}
					if runtime.GOOS == "windows" {
						out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
					}
				} else {
					out.T(out.Tip, "You can specify more memory for your cluster using `minikube start --memory` ")
				}
			}

		}
	}

	// // NodeReady means kubelet is healthy and ready to accept pods.
	// NodeReady NodeConditionType = "Ready"
	// // NodePIDPressure means the kubelet is under pressure due to insufficient available PID.
	// NodePIDPressure NodeConditionType = "PIDPressure"
	// // NodeNetworkUnavailable means that network for the node is not correctly configured.
	// NodeNetworkUnavailable NodeConditionType = "NetworkUnavailable"

	// start := time.Now()

	// 	// equivalent to manual check of 'kubectl --context profile get serviceaccount default'
	// 	sas, err := cs.CoreV1().ServiceAccounts("default").List(meta.ListOptions{})
	// 	if err != nil {
	// 		glog.Infof("temproary error waiting for default SA: %v", err)
	// 		return err
	// 	}
	// 	for _, sa := range sas.Items {
	// 		if sa.Name == "default" {
	// 			glog.Infof("found service account: %q", sa.Name)
	// 			return nil
	// 		}
	// 	}
	// 	return fmt.Errorf("couldn't find default service account")
	// if err := wait.PollImmediate(kconst.APICallRetryInterval, timeout, checkRunning); err != nil {
	// 	return errors.Wrapf(err, "checking k8s-apps to be running")
	// }
	// glog.Infof("duration metric: took %s to wait for k8s-apps to be running ...", time.Since(start))

	return nil
}
