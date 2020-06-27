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

package v162

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

const (
	// WantUpdateNotification is the key for WantUpdateNotification
	WantUpdateNotification = "WantUpdateNotification"
	// ReminderWaitPeriodInHours is the key for WantUpdateNotification
	ReminderWaitPeriodInHours = "ReminderWaitPeriodInHours"
	// WantReportError is the key for WantReportError
	WantReportError = "WantReportError"
	// WantReportErrorPrompt is the key for WantReportErrorPrompt
	WantReportErrorPrompt = "WantReportErrorPrompt"
	// WantKubectlDownloadMsg is the key for WantKubectlDownloadMsg
	WantKubectlDownloadMsg = "WantKubectlDownloadMsg"
	// WantNoneDriverWarning is the key for WantNoneDriverWarning
	WantNoneDriverWarning = "WantNoneDriverWarning"
	// MachineProfile is the key for MachineProfile
	MachineProfile = "profile"
	// ShowDriverDeprecationNotification is the key for ShowDriverDeprecationNotification
	ShowDriverDeprecationNotification = "ShowDriverDeprecationNotification"
	// ShowBootstrapperDeprecationNotification is the key for ShowBootstrapperDeprecationNotification
	ShowBootstrapperDeprecationNotification = "ShowBootstrapperDeprecationNotification"
)

var (
	// ErrKeyNotFound is the error returned when a key doesn't exist in the config file
	ErrKeyNotFound = errors.New("specified key could not be found in config")
)

// MinikubeConfig represents minikube config
type MinikubeConfig map[string]interface{}

// Load loads the kubernetes and machine config for the current machine
func Load(profile string) (*MachineConfig, error) {
	return DefaultLoader.LoadConfigFromFile(profile)
}

// Loader loads the kubernetes and machine config based on the machine profile name
type Loader interface {
	LoadConfigFromFile(profile string, miniHome ...string) (*MachineConfig, error)
}

type simpleConfigLoader struct{}

// DefaultLoader is the default config loader
var DefaultLoader Loader = &simpleConfigLoader{}

func (c *simpleConfigLoader) LoadConfigFromFile(profileName string, miniHome ...string) (*MachineConfig, error) {
	var cc MachineConfig
	// Move to profile package
	path := profileFilePath(profileName, miniHome...)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, err
	}
	return &cc, nil
}

func IsValid(machineConfig *MachineConfig) bool {
	if machineConfig == nil {
		return false
	}

	if machineConfig.Memory == 0 || machineConfig.MinikubeISO == "" || machineConfig.VMDriver == "" {
		return false
	}

	return true
}
