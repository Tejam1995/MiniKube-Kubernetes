// +build integration

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

package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/hashicorp/go-getter"
	"k8s.io/minikube/pkg/util/retry"
)

func TestSkaffold(t *testing.T) {
	// get unique profile for test
	MaybeParallel(t)
	profile := UniqueProfileName("skaffold")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(5))
	defer CleanupWithLogs(t, profile, cancel)

	// install latest skaffold release
	tf, err := installSkaffold()
	if err != nil {
		t.Fatalf("skaffold release installation failed: %v", err)
	}
	defer os.Remove(tf.Name())

	// start minikube cluster
	args := append([]string{"start", "-p", profile, "--memory=2200"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("starting minikube: %v\n%s", err, rr.Output())
	}

	// make sure "skaffold run" exits without failure
	cmd := exec.CommandContext(ctx, tf.Name(), "run", "--kube-context", profile, "--status-check=true", "--port-forward=false")
	cmd.Dir = "testdata/skaffold"
	rr, err = Run(t, cmd)
	if err != nil {
		t.Fatalf("error running skaffold: %v\n%s", err, rr.Output())
	}

	// make sure expected deployment is running
	if _, err := PodWait(ctx, t, profile, "default", "app=leeroy-app", Minutes(1)); err != nil {
		t.Fatalf("failed waiting for pod leeroy-app: %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "default", "app=leeroy-web", Minutes(1)); err != nil {
		t.Fatalf("failed waiting for pod leeroy-web: %v", err)
	}
}

// installSkaffold installs the latest release of skaffold
func installSkaffold() (f *os.File, err error) {
	tf, err := ioutil.TempFile("", "skaffold.exe")
	if err != nil {
		return tf, err
	}
	tf.Close()

	url := "https://storage.googleapis.com/skaffold/releases/latest/skaffold-%s-amd64"
	url = fmt.Sprintf(url, runtime.GOOS)
	if runtime.GOOS == "windows" {
		url += ".exe"
	}

	if err := retry.Expo(func() error { return getter.GetFile(tf.Name(), url) }, 3*time.Second, Minutes(3)); err != nil {
		return tf, err
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(tf.Name(), 0700); err != nil {
			return tf, err
		}
	}
	return tf, nil
}
