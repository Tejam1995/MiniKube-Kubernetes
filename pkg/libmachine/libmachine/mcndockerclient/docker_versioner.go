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

package mcndockerclient

import "fmt"

var CurrentDockerVersioner DockerVersioner = &defaultDockerVersioner{}

type DockerVersioner interface {
	DockerVersion(host DockerHost) (string, error)
}

func DockerVersion(host DockerHost) (string, error) {
	return CurrentDockerVersioner.DockerVersion(host)
}

type defaultDockerVersioner struct{}

func (dv *defaultDockerVersioner) DockerVersion(host DockerHost) (string, error) {
	client, err := DockerClient(host)
	if err != nil {
		return "", fmt.Errorf("unable to query docker version: %s", err)
	}

	version, err := client.Version()
	if err != nil {
		return "", fmt.Errorf("unable to query docker version: %s", err)
	}

	return version.Version, nil
}
