// +build integration

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

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"golang.org/x/build/kubernetes/api"
)

// ProxyAddress is the location of the currently configured HTTP server
var ProxyAddress string

// validateFunc are for subtests that share a single setup
type validateFunc func(context.Context, *testing.T, string)

// TestFunctional are functionality tests which can safely share a profile in parallel
func TestFunctional(t *testing.T) {

	profile := Profile("functional")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	// Serial tests
	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"StartWithProxy", validateStartWithProxy}, // Set everything else up for success
			{"KubeContext", validateKubeContext},       // Racy: must come immediately after "minikube start"
			{"ConfigCmd", validateConfigCmd},           // Each subtest causes necessary side effects
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				tc.validator(ctx, t, profile)
			})
		}
	})

	// Now that we are out of the woods, lets go.
	MaybeParallel(t)

	// Parallelized tests
	t.Run("parallel", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"AddonManager", validateAddonManager},
			{"ComponentHealth", validateComponentHealth},
			{"DashboardCmd", validateDashboardCmd},
			{"DNS", validateDNS},
			{"LogsCmd", validateLogsCmd},
			{"MountCmd", validateMountCmd},
			{"ProfileCmd", validateProfileCmd},
			{"ServicesCmd", validateServicesCmd},
			{"PersistentVolumeClaim", validatePersistentVolumeClaim},
			{"TunnelCmd", validateTunnelCmd},
			{"SSHCmd", validateSSHCmd},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				MaybeParallel(t)
				tc.validator(ctx, t, profile)
			})
		}
	})
}

func validateStartWithProxy(ctx context.Context, t *testing.T, profile string) {
	srv, err := startHTTPProxy(t)
	if err != nil {
		t.Fatalf("Failed to set up the test proxy: %s", err)
	}
	ProxyAddress = srv.Addr
	args := []string{"env", fmt.Sprintf("HTTP_PROXY=%s", ProxyAddress), "NO_PROXY=", Target(), "start", "-p", profile}
	rr, err := Run(ctx, t, "/usr/bin/env", args...)
	if err != nil {
		t.Errorf("%s failed: %v", args, err)
	}

	want := "Found network options:"
	if !strings.Contains(rr.Stdout.String(), want) {
		t.Errorf("start stdout=%s, want: *%s*", rr.Stdout.String(), want)
	}

	want = "You appear to be using a proxy"
	if !strings.Contains(rr.Stderr.String(), want) {
		t.Errorf("start stderr=%s, want: *%s*", rr.Stderr.String(), want)
	}
}

// validateKubeContext asserts that kubectl is properly configured (race-condition prone!)
func validateKubeContext(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(ctx, t, "kubectl", "config", "current-context")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	if !strings.Contains(rr.Stdout.String(), profile) {
		t.Errorf("current-context = %q, want %q", rr.Stdout.String(), profile)
	}
}

// validateAddonManager asserts that the kube-addon-manager pod is deployed properly
func validateAddonManager(ctx context.Context, t *testing.T, profile string) {
	if _, err := PodWait(ctx, t, profile, "kube-system", "component=kube-addon-manager", 1*time.Minute); err != nil {
		t.Errorf("wait: %v", err)
	}
}

// validateComponentHealth asserts that all Kubernetes components are healthy
func validateComponentHealth(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(ctx, t, "kubectl", "--context", profile, "get", "cs", "-o=json")
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
	cs := api.ComponentStatusList{}
	d := json.NewDecoder(bytes.NewReader(rr.Stdout.Bytes()))
	if err := d.Decode(&cs); err != nil {
		t.Fatalf("decode: %v", err)
	}

	for _, i := range cs.Items {
		status := api.ConditionFalse
		for _, c := range i.Conditions {
			if c.Type != api.ComponentHealthy {
				continue
			}
			status = c.Status
		}
		if status != api.ConditionTrue {
			t.Errorf("unexpected status: %v - item: %+v", status, i)
		}
	}
}

// validateDashboardCmd asserts that the dashboard command works
func validateDashboardCmd(ctx context.Context, t *testing.T, profile string) {
	args := []string{"dashboard", "--url", "-p", profile, "--alsologtostderr", "-v=1"}
	ss, err := Start(ctx, t, Target(), args...)
	if err != nil {
		t.Errorf("%s failed: %v", args, err)
	}
	defer func() {
		ss.Stop(t)
	}()

	start := time.Now()
	s, err := ReadLineWithTimeout(ss.Stdout, 300*time.Second)
	if err != nil {
		t.Fatalf("failed to read url within %s: %v\n", time.Since(start), err)
	}

	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		t.Fatalf("failed to parse %q: %v", s, err)
	}

	resp, err := retryablehttp.Get(u.String())
	if err != nil {
		t.Errorf("failed get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unable to read http response body: %v", err)
		}
		t.Errorf("%s returned status code %d, expected %d.\nbody:\n%s", u, resp.StatusCode, http.StatusOK, body)
	}
}

// validateDNS asserts that all Kubernetes DNS is healthy
func validateDNS(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(ctx, t, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "busybox.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	names, err := PodWait(ctx, t, profile, "default", "integration-test=busybox", 3*time.Minute)
	if err != nil {
		t.Fatalf("wait: %v", err)
	}

	rr, err = Run(ctx, t, "kubectl", "--context", profile, "exec", names[0], "nslookup", "kubernetes.default")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}

	want := []byte("10.96.0.1")
	if !bytes.Contains(rr.Stdout.Bytes(), want) {
		t.Errorf("nslookup: got=%q, want=*%q*", rr.Stdout.Bytes(), want)
	}
}

// validateConfigCmd asserts basic "config" command functionality
func validateConfigCmd(ctx context.Context, t *testing.T, profile string) {
	tests := []struct {
		args    []string
		wantOut string
		wantErr string
	}{
		{[]string{"unset", "cpus"}, "", ""},
		{[]string{"get", "cpus"}, "", "Error: specified key could not be found in config"},
		{[]string{"set", "cpus", "2"}, "! These changes will take effect upon a minikube delete and then a minikube start", ""},
		{[]string{"get", "cpus"}, "2", ""},
		{[]string{"unset", "cpus"}, "", ""},
		{[]string{"get", "cpus"}, "", "Error: specified key could not be found in config"},
	}

	for _, tc := range tests {
		args := append([]string{"-p", profile, "config"}, tc.args...)
		rr, err := Run(ctx, t, Target(), args...)
		if err != nil && tc.wantErr == "" {
			t.Errorf("unexpected failure: %s failed: %v", rr.Args, err)
		}

		got := strings.TrimSpace(rr.Stdout.String())
		if got != tc.wantOut {
			t.Errorf("%s stdout got: %q, want: %q", rr.Command(), got, tc.wantOut)
		}
		got = strings.TrimSpace(rr.Stderr.String())
		if got != tc.wantErr {
			t.Errorf("%s stderr got: %q, want: %q", rr.Command(), got, tc.wantErr)
		}
	}
}

// validateLogsCmd asserts basic "logs" command functionality
func validateLogsCmd(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(ctx, t, Target(), "-p", profile, "logs")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	for _, word := range []string{"Docker", "apiserver", "Linux", "kubelet"} {
		if !strings.Contains(rr.Stdout.String(), word) {
			t.Errorf("minikube logs missing expected word: %q", word)
		}
	}
}

// validateProfileCmd asserts basic "profile" command functionality
func validateProfileCmd(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(ctx, t, Target(), "profile", "list")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
}

// validateServiceCmd asserts basic "service" command functionality
func validateServicesCmd(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(ctx, t, Target(), "-p", profile, "service", "list")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	if !strings.Contains(rr.Stdout.String(), "kubernetes") {
		t.Errorf("service list got %q, wanted *kubernetes*", rr.Stdout.String())
	}
}

// validateSSHCmd asserts basic "ssh" command functionality
func validateSSHCmd(ctx context.Context, t *testing.T, profile string) {
	if NoneDriver() {
		t.Skipf("skipping: ssh unsupported by none")
	}
	want := "hello\r\n"
	rr, err := Run(ctx, t, Target(), "-p", profile, "ssh", fmt.Sprintf("echo hello"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	if rr.Stdout.String() != want {
		t.Errorf("%v = %q, want = %q", rr.Args, rr.Stdout.String(), want)
	}
}

// startHTTPProxy runs a local http proxy and sets the env vars for it.
func startHTTPProxy(t *testing.T) (*http.Server, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get an open port")
	}

	addr := fmt.Sprintf("localhost:%d", port)
	proxy := goproxy.NewProxyHttpServer()
	srv := &http.Server{Addr: addr, Handler: proxy}
	go func(s *http.Server, t *testing.T) {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("Failed to start http server for proxy mock")
		}
	}(srv, t)
	return srv, nil
}
