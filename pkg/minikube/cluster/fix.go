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

package cluster

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util/retry"
)

// hostRunner is a minimal host.Host based interface for running commands
type hostRunner interface {
	RunSSHCommand(string) (string, error)
}

const (
	// The maximum the guest VM clock is allowed to be ahead and behind. This value is intentionally
	// large to allow for inaccurate methodology, but still small enough so that certificates are likely valid.
	maxClockDesyncSeconds = 2.1
)

// fixHost fixes up a previously configured VM so that it is ready to run Kubernetes
func fixHost(api libmachine.API, mc config.MachineConfig) (*host.Host, error) {
	out.T(out.Waiting, "Reconfiguring existing host ...")

	start := time.Now()
	glog.Infof("fixHost starting: %s", mc.Name)
	defer func() {
		glog.Infof("fixHost completed within %s", time.Since(start))
	}()

	h, err := api.Load(mc.Name)
	if err != nil {
		return h, errors.Wrap(err, "Error loading existing host. Please try running [minikube delete], then run [minikube start] again.")
	}

	s, err := h.Driver.GetState()
	if err != nil {
		return h, errors.Wrap(err, "Error getting state for host")
	}

	if s == state.Running {
		out.T(out.Running, `Using the running {{.driver_name}} "{{.profile_name}}" VM ...`, out.V{"driver_name": mc.VMDriver, "profile_name": mc.Name})
	} else {
		out.T(out.Restarting, `Starting existing {{.driver_name}} VM for "{{.profile_name}}" ...`, out.V{"driver_name": mc.VMDriver, "profile_name": mc.Name})
		if err := h.Driver.Start(); err != nil {
			return h, errors.Wrap(err, "driver start")
		}
		if err := api.Save(h); err != nil {
			return h, errors.Wrap(err, "save")
		}
	}

	e := engineOptions(mc)
	if len(e.Env) > 0 {
		h.HostOptions.EngineOptions.Env = e.Env
		glog.Infof("Detecting provisioner ...")
		provisioner, err := provision.DetectProvisioner(h.Driver)
		if err != nil {
			return h, errors.Wrap(err, "detecting provisioner")
		}
		if err := provisioner.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions); err != nil {
			return h, errors.Wrap(err, "provision")
		}
	}

	if h.DriverName == driver.Mock {
		return h, nil
	}

	if err := postStartSetup(h, mc); err != nil {
		return h, errors.Wrap(err, "post-start")
	}

	glog.Infof("Configuring auth for driver %s ...", h.Driver.DriverName())
	if err := h.ConfigureAuth(); err != nil {
		return h, &retry.RetriableError{Err: errors.Wrap(err, "Error configuring auth on host")}
	}
	return h, ensureSyncedGuestClock(h, mc.VMDriver)
}

// ensureGuestClockSync ensures that the guest system clock is relatively in-sync
func ensureSyncedGuestClock(h hostRunner, drv string) error {
	if !driver.IsVM(drv) {
		return nil
	}
	d, err := guestClockDelta(h, time.Now())
	if err != nil {
		glog.Warningf("Unable to measure system clock delta: %v", err)
		return nil
	}
	if math.Abs(d.Seconds()) < maxClockDesyncSeconds {
		glog.Infof("guest clock delta is within tolerance: %s", d)
		return nil
	}
	if err := adjustGuestClock(h, time.Now()); err != nil {
		return errors.Wrap(err, "adjusting system clock")
	}
	return nil
}

// guestClockDelta returns the approximate difference between the host and guest system clock
// NOTE: This does not currently take into account ssh latency.
func guestClockDelta(h hostRunner, local time.Time) (time.Duration, error) {
	out, err := h.RunSSHCommand("date +%s.%N")
	if err != nil {
		return 0, errors.Wrap(err, "get clock")
	}
	glog.Infof("guest clock: %s", out)
	ns := strings.Split(strings.TrimSpace(out), ".")
	secs, err := strconv.ParseInt(strings.TrimSpace(ns[0]), 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "atoi")
	}
	nsecs, err := strconv.ParseInt(strings.TrimSpace(ns[1]), 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "atoi")
	}
	// NOTE: In a synced state, remote is a few hundred ms ahead of local
	remote := time.Unix(secs, nsecs)
	d := remote.Sub(local)
	glog.Infof("Guest: %s Remote: %s (delta=%s)", remote, local, d)
	return d, nil
}

// adjustSystemClock adjusts the guest system clock to be nearer to the host system clock
func adjustGuestClock(h hostRunner, t time.Time) error {
	out, err := h.RunSSHCommand(fmt.Sprintf("sudo date -s @%d", t.Unix()))
	glog.Infof("clock set: %s (err=%v)", out, err)
	return err
}
