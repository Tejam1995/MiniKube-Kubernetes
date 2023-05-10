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
	"fmt"
	"os/exec"
	"regexp"

	"github.com/pkg/errors"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/cruntimeInstaller"
	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/libmachine/swarm"
)

var (
	ErrUnknownYumOsRelease = errors.New("unknown OS for Yum repository")
	engineConfigTemplate   = `[Service]
ExecStart=
ExecStart=/usr/bin/dockerd -H tcp://0.0.0.0:{{.DockerPort}} -H unix:///var/run/docker.sock --storage-driver {{.EngineOptions.StorageDriver}} --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
Environment={{range .EngineOptions.Env}}{{ printf "%q" . }} {{end}}
`
	majorVersionRE = regexp.MustCompile(`^(\d+)(\..*)?`)
)

type PackageListInfo struct {
	OsRelease        string
	OsReleaseVersion string
}

func init() {
	Register("RedHat", &RegisteredProvisioner{
		New: func(d drivers.Driver) Provisioner {
			return NewRedHatProvisioner("rhel", d)
		},
	})
}

func NewRedHatProvisioner(osReleaseID string, d drivers.Driver) *RedHatProvisioner {
	systemdProvisioner := NewSystemdProvisioner(osReleaseID, d)
	systemdProvisioner.Commander = RedHatCommander{Driver: d}
	return &RedHatProvisioner{
		systemdProvisioner,
	}
}

type RedHatProvisioner struct {
	SystemdProvisioner
}

func (provisioner *RedHatProvisioner) String() string {
	return "redhat"
}

func (provisioner *RedHatProvisioner) SetHostname(hostname string) error {
	// we have to have SetHostname here as well to use the RedHat provisioner
	// RunCmd to add the tty allocation
	cmd := fmt.Sprintf(
		"sudo hostname %s && echo %q | sudo tee /etc/hostname",
		hostname,
		hostname,
	)
	if _, err := provisioner.RunCmd(exec.Command("bash", "-c", cmd)); err != nil {
		return err
	}

	cmd = fmt.Sprintf(
		"if grep -xq 127.0.1.1.* /etc/hosts; then sudo sed -i 's/^127.0.1.1.*/127.0.1.1 %s/g' /etc/hosts; else echo '127.0.1.1 %s' | sudo tee -a /etc/hosts; fi",
		hostname,
		hostname,
	)
	if _, err := provisioner.RunCmd(exec.Command("bash", "-c", cmd)); err != nil {
		return err
	}

	return nil
}

func (provisioner *RedHatProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var packageAction string

	switch action {
	case pkgaction.Install:
		packageAction = "install"
	case pkgaction.Remove:
		packageAction = "remove"
	case pkgaction.Purge:
		packageAction = "remove"
	case pkgaction.Upgrade:
		packageAction = "upgrade"
	}

	if _, err := provisioner.RunCmd(exec.Command("sudo", "-E", "yum", packageAction, "-y", name)); err != nil {
		return err
	}

	return nil
}

func (provisioner *RedHatProvisioner) dockerDaemonResponding() bool {
	log.Debug("checking docker daemon")

	if out, err := provisioner.RunCmd(exec.Command("sudo", "docker", "version")); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		log.Debugf("'sudo docker version' output:\n%s", out)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *RedHatProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	if !provisioner.Driver.IsManaged() {
		return nil
	}

	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = &authOptions
	provisioner.EngineOptions = &engineOptions
	swarmOptions.Env = engineOptions.Env

	// set default storage driver for redhat
	storageDriver, err := decideStorageDriver(provisioner, "overlay2", engineOptions.StorageDriver)
	if err != nil {
		return err
	}
	provisioner.EngineOptions.StorageDriver = storageDriver

	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	for _, pkg := range provisioner.Packages {
		log.Debugf("installing base package: name=%s", pkg)
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	// update OS -- this is needed for libdevicemapper and the docker install
	cmd := "sudo -E yum -y update -x docker-*"
	if _, err := provisioner.RunCmd(exec.Command("bash", "-c", cmd)); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	klog.Infof("installing container runtime into the machine")

	rnr, err := provisioner.Driver.GetRunner()
	if err != nil {
		return errors.Wrap(err, "while getting runner for cruntime installer")
	}

	instllr := cruntimeInstaller.DetectCRuntimeInstaller(provisioner.EngineOptions, rnr, provisioner.Driver.DriverName(), provisioner.AuthOptions)
	return instllr.InstallCRuntime()
}
