/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package machine

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/download"
)

type copyFailRunner struct {
	runner.Runner
}

func (copyFailRunner) Copy(_ assets.CopyableFile) error {
	return fmt.Errorf("test error during copy file")
}

func newFakeCommandRunnerCopyFail() runner.Runner {
	return copyFailRunner{runner.NewFakeCommandRunner()}
}

func TestCopyBinary(t *testing.T) {
	var tc = []struct {
		lastUpdateCheckFilePath string
		src, dst, desc          string
		err                     bool
		runner                  runner.Runner
	}{
		{
			desc:   "not existing src",
			dst:    "/tmp/testCopyBinary1",
			src:    "/tmp/testCopyBinary2",
			err:    true,
			runner: runner.NewFakeCommandRunner(),
		},
		{
			desc:   "src /etc/hosts",
			dst:    "/tmp/testCopyBinary1",
			src:    "/etc/hosts",
			err:    false,
			runner: runner.NewFakeCommandRunner(),
		},
		{
			desc:   "existing src, copy fail",
			dst:    "/etc/passwd",
			src:    "/etc/hosts",
			err:    true,
			runner: newFakeCommandRunnerCopyFail(),
		},
	}
	for _, test := range tc {
		t.Run(test.desc, func(t *testing.T) {
			err := CopyBinary(test.runner, test.src, test.dst)
			if err != nil && !test.err {
				t.Fatalf("Error %v expected but not occurred", err)
			}
			if err == nil && test.err {
				t.Fatal("Unexpected error")
			}
		})
	}
}

func TestCacheBinariesForBootstrapper(t *testing.T) {
	download.DownloadMock = download.CreateDstDownloadMock

	minikubeHome := t.TempDir()

	var tc = []struct {
		version, clusterBootstrapper string
		minikubeHome                 string
		err                          bool
	}{
		{
			version:             "v1.16.0",
			clusterBootstrapper: bootstrapper.Kubeadm,
			err:                 false,
			minikubeHome:        minikubeHome,
		},
		{
			version:             "invalid version",
			clusterBootstrapper: bootstrapper.Kubeadm,
			err:                 true,
			minikubeHome:        minikubeHome,
		},
	}
	for _, test := range tc {
		t.Run(test.version, func(t *testing.T) {
			t.Setenv("MINIKUBE_HOME", test.minikubeHome)
			err := CacheBinariesForBootstrapper(test.version, test.clusterBootstrapper, nil, "")
			if err != nil && !test.err {
				t.Fatalf("Got unexpected error %v", err)
			}
			if err == nil && test.err {
				t.Fatalf("Expected error but got %v", err)
			}
		})
	}
}

func TestExcludedBinariesNotDownloaded(t *testing.T) {
	clusterBootstrapper := bootstrapper.Kubeadm
	binaryList := bootstrapper.GetCachedBinaryList(clusterBootstrapper)
	binaryToExclude := binaryList[0]

	download.DownloadMock = func(src, dst string) error {
		if strings.Contains(src, binaryToExclude) {
			t.Errorf("Excluded binary was downloaded! Binary to exclude: %s", binaryToExclude)
		}
		return download.CreateDstDownloadMock(src, dst)
	}

	minikubeHome := t.TempDir()
	t.Setenv("MINIKUBE_HOME", minikubeHome)

	if err := CacheBinariesForBootstrapper("v1.16.0", clusterBootstrapper, []string{binaryToExclude}, ""); err != nil {
		t.Errorf("Failed to cache binaries: %v", err)
	}
}
