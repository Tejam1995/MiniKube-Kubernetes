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

package oci

import (
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// SysInfo Info represents common system Information between docker and podman that minikube cares
type SysInfo struct {
	CPUs        int   // CPUs is Number of CPUs
	TotalMemory int64 // TotalMemory Total available ram
}

// dockerSysInfo represents the output of docker system info --format '{{json .}}'
type dockerSysInfo struct {
	ID                string      `json:"ID"`
	Containers        int         `json:"Containers"`
	ContainersRunning int         `json:"ContainersRunning"`
	ContainersPaused  int         `json:"ContainersPaused"`
	ContainersStopped int         `json:"ContainersStopped"`
	Images            int         `json:"Images"`
	Driver            string      `json:"Driver"`
	DriverStatus      [][]string  `json:"DriverStatus"`
	SystemStatus      interface{} `json:"SystemStatus"`
	Plugins           struct {
		Volume        []string    `json:"Volume"`
		Network       []string    `json:"Network"`
		Authorization interface{} `json:"Authorization"`
		Log           []string    `json:"Log"`
	} `json:"Plugins"`
	MemoryLimit        bool      `json:"MemoryLimit"`
	SwapLimit          bool      `json:"SwapLimit"`
	KernelMemory       bool      `json:"KernelMemory"`
	KernelMemoryTCP    bool      `json:"KernelMemoryTCP"`
	CPUCfsPeriod       bool      `json:"CpuCfsPeriod"`
	CPUCfsQuota        bool      `json:"CpuCfsQuota"`
	CPUShares          bool      `json:"CPUShares"`
	CPUSet             bool      `json:"CPUSet"`
	PidsLimit          bool      `json:"PidsLimit"`
	IPv4Forwarding     bool      `json:"IPv4Forwarding"`
	BridgeNfIptables   bool      `json:"BridgeNfIptables"`
	BridgeNfIP6Tables  bool      `json:"BridgeNfIp6tables"`
	Debug              bool      `json:"Debug"`
	NFd                int       `json:"NFd"`
	OomKillDisable     bool      `json:"OomKillDisable"`
	NGoroutines        int       `json:"NGoroutines"`
	SystemTime         time.Time `json:"SystemTime"`
	LoggingDriver      string    `json:"LoggingDriver"`
	CgroupDriver       string    `json:"CgroupDriver"`
	NEventsListener    int       `json:"NEventsListener"`
	KernelVersion      string    `json:"KernelVersion"`
	OperatingSystem    string    `json:"OperatingSystem"`
	OSType             string    `json:"OSType"`
	Architecture       string    `json:"Architecture"`
	IndexServerAddress string    `json:"IndexServerAddress"`
	RegistryConfig     struct {
		AllowNondistributableArtifactsCIDRs     []interface{} `json:"AllowNondistributableArtifactsCIDRs"`
		AllowNondistributableArtifactsHostnames []interface{} `json:"AllowNondistributableArtifactsHostnames"`
		InsecureRegistryCIDRs                   []string      `json:"InsecureRegistryCIDRs"`
		IndexConfigs                            struct {
			DockerIo struct {
				Name     string        `json:"Name"`
				Mirrors  []interface{} `json:"Mirrors"`
				Secure   bool          `json:"Secure"`
				Official bool          `json:"Official"`
			} `json:"docker.io"`
		} `json:"IndexConfigs"`
		Mirrors []interface{} `json:"Mirrors"`
	} `json:"RegistryConfig"`
	NCPU              int           `json:"NCPU"`
	MemTotal          int64         `json:"MemTotal"`
	GenericResources  interface{}   `json:"GenericResources"`
	DockerRootDir     string        `json:"DockerRootDir"`
	HTTPProxy         string        `json:"HttpProxy"`
	HTTPSProxy        string        `json:"HttpsProxy"`
	NoProxy           string        `json:"NoProxy"`
	Name              string        `json:"Name"`
	Labels            []interface{} `json:"Labels"`
	ExperimentalBuild bool          `json:"ExperimentalBuild"`
	ServerVersion     string        `json:"ServerVersion"`
	ClusterStore      string        `json:"ClusterStore"`
	ClusterAdvertise  string        `json:"ClusterAdvertise"`
	Runtimes          struct {
		Runc struct {
			Path string `json:"path"`
		} `json:"runc"`
	} `json:"Runtimes"`
	DefaultRuntime string `json:"DefaultRuntime"`
	Swarm          struct {
		NodeID           string      `json:"NodeID"`
		NodeAddr         string      `json:"NodeAddr"`
		LocalNodeState   string      `json:"LocalNodeState"`
		ControlAvailable bool        `json:"ControlAvailable"`
		Error            string      `json:"Error"`
		RemoteManagers   interface{} `json:"RemoteManagers"`
	} `json:"Swarm"`
	LiveRestoreEnabled bool   `json:"LiveRestoreEnabled"`
	Isolation          string `json:"Isolation"`
	InitBinary         string `json:"InitBinary"`
	ContainerdCommit   struct {
		ID       string `json:"ID"`
		Expected string `json:"Expected"`
	} `json:"ContainerdCommit"`
	RuncCommit struct {
		ID       string `json:"ID"`
		Expected string `json:"Expected"`
	} `json:"RuncCommit"`
	InitCommit struct {
		ID       string `json:"ID"`
		Expected string `json:"Expected"`
	} `json:"InitCommit"`
	SecurityOptions []string    `json:"SecurityOptions"`
	ProductLicense  string      `json:"ProductLicense"`
	Warnings        interface{} `json:"Warnings"`
	ClientInfo      struct {
		Debug    bool          `json:"Debug"`
		Plugins  []interface{} `json:"Plugins"`
		Warnings interface{}   `json:"Warnings"`
	} `json:"ClientInfo"`
}

// podmanSysInfo represents the output of podman system info --format '{{json .}}'
type podmanSysInfo struct {
	Host struct {
		BuildahVersion string `json:"BuildahVersion"`
		CgroupVersion  string `json:"CgroupVersion"`
		Conmon         struct {
			Package string `json:"package"`
			Path    string `json:"path"`
			Version string `json:"version"`
		} `json:"Conmon"`
		Distribution struct {
			Distribution string `json:"distribution"`
			Version      string `json:"version"`
		} `json:"Distribution"`
		MemFree    int   `json:"MemFree"`
		MemTotal   int64 `json:"MemTotal"`
		OCIRuntime struct {
			Name    string `json:"name"`
			Package string `json:"package"`
			Path    string `json:"path"`
			Version string `json:"version"`
		} `json:"OCIRuntime"`
		SwapFree    int    `json:"SwapFree"`
		SwapTotal   int    `json:"SwapTotal"`
		Arch        string `json:"arch"`
		Cpus        int    `json:"cpus"`
		Eventlogger string `json:"eventlogger"`
		Hostname    string `json:"hostname"`
		Kernel      string `json:"kernel"`
		Os          string `json:"os"`
		Rootless    bool   `json:"rootless"`
		Uptime      string `json:"uptime"`
	} `json:"host"`
	Registries struct {
		Search []string `json:"search"`
	} `json:"registries"`
	Store struct {
		ConfigFile     string `json:"ConfigFile"`
		ContainerStore struct {
			Number int `json:"number"`
		} `json:"ContainerStore"`
		GraphDriverName string `json:"GraphDriverName"`
		GraphOptions    struct {
		} `json:"GraphOptions"`
		GraphRoot   string `json:"GraphRoot"`
		GraphStatus struct {
			BackingFilesystem string `json:"Backing Filesystem"`
			NativeOverlayDiff string `json:"Native Overlay Diff"`
			SupportsDType     string `json:"Supports d_type"`
			UsingMetacopy     string `json:"Using metacopy"`
		} `json:"GraphStatus"`
		ImageStore struct {
			Number int `json:"number"`
		} `json:"ImageStore"`
		RunRoot    string `json:"RunRoot"`
		VolumePath string `json:"VolumePath"`
	} `json:"store"`
}

// DaemonInfo returns docker system info
func DaemonInfo(ociBin string) (SysInfo, error) {
	if ociBin == Podman {
		return podmanSystemInfo()
	}
	return dockerSystemInfo()
}

func dockerSystemInfo() (SysInfo, error) {
	var ds dockerSysInfo
	var s SysInfo
	cmd := exec.Command(Docker, "system", "info", "--format", "'{{json .}}'")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return s, errors.Wrap(err, "get docker system info")
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &ds); err != nil {
		return s, errors.Wrapf(err, "unmarshal docker system info")
	}
	s.CPUs = ds.NCPU
	s.TotalMemory = ds.MemTotal
	return s, nil
}

func podmanSystemInfo() (SysInfo, error) {
	var ps podmanSysInfo
	var s SysInfo
	cmd := exec.Command(Podman, "system", "info", "--format", "'{{json .}}'")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return s, errors.Wrap(err, "get podman system info")
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &ps); err != nil {
		return s, errors.Wrapf(err, "unmarshal podman system info")
	}
	s.CPUs = ps.Host.Cpus
	s.TotalMemory = ps.Host.MemTotal
	return s, nil
}
