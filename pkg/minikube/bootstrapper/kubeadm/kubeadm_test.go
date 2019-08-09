/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package kubeadm

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/util"
)

func TestGenerateKubeletConfig(t *testing.T) {
	tests := []struct {
		description string
		cfg         config.KubernetesConfig
		expected    string
		shouldErr   bool
	}{
		{
			description: "old docker",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.100",
				KubernetesVersion: constants.OldestKubernetesVersion,
				NodeName:          "minikube",
				ContainerRuntime:  "docker",
			},
			expected: `[Unit]
Wants=docker.socket

[Service]
ExecStart=
ExecStart=/var/lib/minikube/binaries/v1.10.13/kubelet --allow-privileged=true --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cadvisor-port=0 --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-dns=10.96.0.10 --cluster-domain=cluster.local --container-runtime=docker --fail-swap-on=false --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --pod-manifest-path=/etc/kubernetes/manifests

[Install]
`,
		},
		{
			description: "newest cri runtime",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.100",
				KubernetesVersion: constants.NewestKubernetesVersion,
				NodeName:          "minikube",
				ContainerRuntime:  "cri-o",
			},
			expected: `[Unit]
Wants=crio.service

[Service]
ExecStart=
ExecStart=/var/lib/minikube/binaries/v1.15.2/kubelet --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-dns=10.96.0.10 --cluster-domain=cluster.local --container-runtime=remote --container-runtime-endpoint=/var/run/crio/crio.sock --fail-swap-on=false --hostname-override=minikube --image-service-endpoint=/var/run/crio/crio.sock --kubeconfig=/etc/kubernetes/kubelet.conf --pod-manifest-path=/etc/kubernetes/manifests --runtime-request-timeout=15m

[Install]
`,
		},
		{
			description: "default containerd runtime",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.100",
				KubernetesVersion: constants.DefaultKubernetesVersion,
				NodeName:          "minikube",
				ContainerRuntime:  "containerd",
			},
			expected: `[Unit]
Wants=containerd.service

[Service]
ExecStart=
ExecStart=/var/lib/minikube/binaries/v1.15.2/kubelet --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-dns=10.96.0.10 --cluster-domain=cluster.local --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --fail-swap-on=false --hostname-override=minikube --image-service-endpoint=unix:///run/containerd/containerd.sock --kubeconfig=/etc/kubernetes/kubelet.conf --pod-manifest-path=/etc/kubernetes/manifests --runtime-request-timeout=15m

[Install]
`,
		},
		{
			description: "docker with custom image repository",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.100",
				KubernetesVersion: constants.DefaultKubernetesVersion,
				NodeName:          "minikube",
				ContainerRuntime:  "docker",
				ImageRepository:   "docker-proxy-image.io/google_containers",
			},
			expected: `[Unit]
Wants=docker.socket

[Service]
ExecStart=
ExecStart=/var/lib/minikube/binaries/v1.15.2/kubelet --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-dns=10.96.0.10 --cluster-domain=cluster.local --container-runtime=docker --fail-swap-on=false --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --pod-infra-container-image=docker-proxy-image.io/google_containers/pause:3.1 --pod-manifest-path=/etc/kubernetes/manifests

[Install]
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			runtime, err := cruntime.New(cruntime.Config{Type: tc.cfg.ContainerRuntime})
			if err != nil {
				t.Fatalf("runtime: %v", err)
			}

			got, err := NewKubeletConfig(tc.cfg, runtime)
			if err != nil && !tc.shouldErr {
				t.Errorf("got unexpected error generating config: %v", err)
				return
			}
			if err == nil && tc.shouldErr {
				t.Errorf("expected error but got none, config: %s", got)
				return
			}

			diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(tc.expected),
				B:        difflib.SplitLines(string(got)),
				FromFile: "Expected",
				ToFile:   "Got",
				Context:  1,
			})
			if err != nil {
				t.Fatalf("diff error: %v", err)
			}
			if diff != "" {
				t.Errorf("unexpected diff:\n%s", diff)
			}
		})
	}
}

func getExtraOpts() []util.ExtraOption {
	return util.ExtraOptionSlice{
		util.ExtraOption{
			Component: Apiserver,
			Key:       "fail-no-swap",
			Value:     "true",
		},
		util.ExtraOption{
			Component: ControllerManager,
			Key:       "kube-api-burst",
			Value:     "32",
		},
		util.ExtraOption{
			Component: Scheduler,
			Key:       "scheduler-name",
			Value:     "mini-scheduler",
		},
		util.ExtraOption{
			Component: Kubeadm,
			Key:       "ignore-preflight-errors",
			Value:     "true",
		},
		util.ExtraOption{
			Component: Kubeadm,
			Key:       "dry-run",
			Value:     "true",
		},
	}
}

func getExtraOptsPodCidr() []util.ExtraOption {
	return util.ExtraOptionSlice{
		util.ExtraOption{
			Component: Kubeadm,
			Key:       "pod-network-cidr",
			Value:     "192.168.32.0/20",
		},
	}
}

func recentReleases() ([]string, error) {
	// test the 6 most recent releases
	versions := []string{"v1.15", "v1.14", "v1.13", "v1.12", "v1.11", "v1.10"}
	foundNewest := false
	foundDefault := false

	for _, v := range versions {
		if strings.HasPrefix(constants.NewestKubernetesVersion, v) {
			foundNewest = true
		}
		if strings.HasPrefix(constants.DefaultKubernetesVersion, v) {
			foundDefault = true
		}
	}

	if !foundNewest {
		return nil, fmt.Errorf("No tests exist yet for newest minor version: %s", constants.NewestKubernetesVersion)
	}

	if !foundDefault {
		return nil, fmt.Errorf("No tests exist yet for default minor version: %s", constants.DefaultKubernetesVersion)
	}

	return versions, nil
}

func TestGenerateConfig(t *testing.T) {
	extraOpts := getExtraOpts()
	extraOptsPodCidr := getExtraOptsPodCidr()
	versions, err := recentReleases()
	if err != nil {
		t.Errorf("versions: %v", err)
	}
	tests := []struct {
		name      string
		runtime   string
		shouldErr bool
		cfg       config.KubernetesConfig
	}{
		{"default", "docker", false, config.KubernetesConfig{}},
		{"containerd", "containerd", false, config.KubernetesConfig{}},
		{"crio", "crio", false, config.KubernetesConfig{}},
		{"options", "docker", false, config.KubernetesConfig{ExtraOptions: extraOpts}},
		{"crio-options-gates", "crio", false, config.KubernetesConfig{ExtraOptions: extraOpts, FeatureGates: "a=b"}},
		{"unknown-component", "docker", true, config.KubernetesConfig{ExtraOptions: util.ExtraOptionSlice{util.ExtraOption{Component: "not-a-real-component", Key: "killswitch", Value: "true"}}}},
		{"containerd-api-port", "containerd", false, config.KubernetesConfig{NodePort: 12345}},
		{"containerd-pod-network-cidr", "containerd", false, config.KubernetesConfig{ExtraOptions: extraOptsPodCidr}},
		{"image-repository", "docker", false, config.KubernetesConfig{ImageRepository: "test/repo"}},
	}
	for _, version := range versions {
		for _, tc := range tests {
			runtime, err := cruntime.New(cruntime.Config{Type: tc.runtime})
			if err != nil {
				t.Fatalf("runtime: %v", err)
			}
			tname := tc.name + "_" + version
			t.Run(tname, func(t *testing.T) {
				cfg := tc.cfg
				cfg.NodeIP = "1.1.1.1"
				cfg.NodeName = "mk"
				cfg.KubernetesVersion = version + ".0"

				got, err := generateConfig(cfg, runtime)
				if err != nil && !tc.shouldErr {
					t.Fatalf("got unexpected error generating config: %v", err)
				}
				if err == nil && tc.shouldErr {
					t.Fatalf("expected error but got none, config: %s", got)
				}
				if tc.shouldErr {
					return
				}
				expected, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s/%s.yaml", version, tc.name))
				if err != nil {
					t.Fatalf("unable to read testdata: %v", err)
				}
				diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
					A:        difflib.SplitLines(string(expected)),
					B:        difflib.SplitLines(string(got)),
					FromFile: "Expected",
					ToFile:   "Got",
					Context:  1,
				})
				if err != nil {
					t.Fatalf("diff error: %v", err)
				}
				if diff != "" {
					t.Errorf("unexpected diff:\n%s\n===== [RAW OUTPUT] =====\n%s", diff, got)
				}
			})
		}
	}
}
