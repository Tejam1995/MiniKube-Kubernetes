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

package engine

import "k8s.io/minikube/pkg/libmachine/libmachine/auth"

const (
	DefaultPort = 2376
)

type Options struct {
	EngineName       string
	ArbitraryFlags   []string
	DNS              []string `json:"Dns"`
	GraphDir         string
	Env              []string
	Ipv6             bool
	InsecureRegistry []string
	Labels           []string
	LogLevel         string
	StorageDriver    string
	SelinuxEnabled   bool
	TLSVerify        bool `json:"TlsVerify"`
	RegistryMirror   []string
	InstallURL       string
}

type ConfigContext struct {
	DockerPort       int
	AuthOptions      auth.Options
	EngineOptions    Options
	DockerOptionsDir string
}
