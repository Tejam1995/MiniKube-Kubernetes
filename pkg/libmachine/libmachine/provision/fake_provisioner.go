/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package provision

import (
	"bytes"
	"os/exec"

	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/serviceaction"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/libmachine/swarm"
)

type FakeDetector struct {
	Provisioner
}

func (fd *FakeDetector) DetectProvisioner(_ drivers.Driver) (Provisioner, error) {
	return fd.Provisioner, nil
}

type FakeProvisioner struct{}

func NewFakeProvisioner(_ drivers.Driver) Provisioner {
	return &FakeProvisioner{}
}

func (fp *FakeProvisioner) RunCmd(*exec.Cmd) (*runner.RunResult, error) {
	return &runner.RunResult{}, nil
}

func (fp *FakeProvisioner) String() string {
	return "fakeprovisioner"
}

func (fp *FakeProvisioner) GetDockerOptionsDir() string {
	return ""
}

func (fp *FakeProvisioner) GetAuthOptions() auth.Options {
	return auth.Options{}
}

func (fp *FakeProvisioner) GetSwarmOptions() swarm.Options {
	return swarm.Options{}
}

func (fp *FakeProvisioner) Package(_ string, _ pkgaction.PackageAction) error {
	return nil
}

func (fp *FakeProvisioner) Hostname() (string, error) {
	return "", nil
}

func (fp *FakeProvisioner) SetHostname(_ string) error {
	return nil
}

func (fp *FakeProvisioner) CompatibleWithHost() bool {
	return true
}

func (fp *FakeProvisioner) Provision(_ swarm.Options, _ auth.Options, _ engine.Options) error {
	return nil
}

func (fp *FakeProvisioner) Service(_ string, _ serviceaction.ServiceAction) error {
	return nil
}

func (fp *FakeProvisioner) GetDriver() drivers.Driver {
	return nil
}

func (fp *FakeProvisioner) SetOsReleaseInfo(_ *OsRelease) {}

func (fp *FakeProvisioner) GetOsReleaseInfo() (*OsRelease, error) {
	return nil, nil
}

type NetstatProvisioner struct {
	*FakeProvisioner
}

// x7NOTE: I'm puzzled...
func (p *NetstatProvisioner) RunCmd(_ *exec.Cmd) (*runner.RunResult, error) {
	return &runner.RunResult{Stdout: *bytes.NewBuffer([]byte(`Active Internet connections (servers and established)
Proto Recv-Q Send-Q Local Address           Foreign Address         State
tcp        0      0 0.0.0.0:ssh             0.0.0.0:*               LISTEN
tcp        0     72 192.168.25.141:ssh      192.168.25.1:63235      ESTABLISHED
tcp        0      0 :::2376                 :::*                    LISTEN
tcp        0      0 :::ssh                  :::*                    LISTEN
Active UNIX domain sockets (servers and established)
Proto RefCnt Flags       Type       State         I-Node Path
unix  2      [ ACC ]     STREAM     LISTENING      17990 /var/run/acpid.socket
unix  2      [ ACC ]     SEQPACKET  LISTENING      14233 /run/udev/control
unix  2      [ ACC ]     STREAM     LISTENING      19365 /var/run/docker.sock
unix  3      [ ]         STREAM     CONNECTED      19774
unix  3      [ ]         STREAM     CONNECTED      19775
unix  3      [ ]         DGRAM                     14243
unix  3      [ ]         DGRAM                     14242`))}, nil
}

func NewNetstatProvisioner() Provisioner {
	return &NetstatProvisioner{
		&FakeProvisioner{},
	}
}
