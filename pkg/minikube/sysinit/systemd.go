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

// sysinit provides an abstraction over init systems like systemctl
package sysinit

import (
	"os/exec"
)

// Systemd is a service manager for systemd distributions
type Systemd struct {
	r Runner
}

// Name returns the name of the init system
func (s *Systemd) Name() string {
	return "systemd"
}

// reload reloads systemd configuration
func (s *Systemd) reload() error {
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "daemon-reload"))
	return err == nil
}

// Active checks if a service is running
func (s *Systemd) Active(svc string) bool {
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "is-active", "--quiet", "service", svc))
	return err == nil
}

// Start starts a service
func (s *Systemd) Start(svc string) error {
	if err := s.reload() {
		return err
	}
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "start", svc))
	return err
}

// Restart restarts a service
func (s *Systemd) Restart(svc string) error {
	if err := s.reload() {
		return err
	}
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "restart", svc))
	return err
}

// Stop stops a service
func (s *Systemd) Stop(svc string) error {
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "restart", svc))
	return err
}

func usesSystemd(r Runner) bool {
	_, err := r.RunCmd(exec.Command("systemctl", "--version"))
	return err == nil
}
