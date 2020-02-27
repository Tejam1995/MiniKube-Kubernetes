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

package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/minikube/localpath"
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
	// DockerEnv contains the environment variables
	DockerEnv []string
	// DockerOpt contains the option parameters
	DockerOpt []string
	// ExtraOptions contains extra options (if any)
	ExtraOptions ExtraOptionSlice
	// AddonList contains the list of addons
	AddonList []string
)

// ErrNotExist is the error returned when a config does not exist
type ErrNotExist struct {
	s string
}

func (e *ErrNotExist) Error() string {
	return e.s
}

// IsNotExist returns whether the error means a nonexistent configuration
func IsNotExist(err error) bool {
	if _, ok := err.(*ErrNotExist); ok {
		return true
	}
	return false
}

// MinikubeConfig represents minikube config
type MinikubeConfig map[string]interface{}

// Get gets a named value from config
func Get(name string) (string, error) {
	m, err := ReadConfig(localpath.ConfigFile())
	if err != nil {
		return "", err
	}
	return get(name, m)
}

func get(name string, config MinikubeConfig) (string, error) {
	if val, ok := config[name]; ok {
		return fmt.Sprintf("%v", val), nil
	}
	return "", ErrKeyNotFound
}

// WriteConfig writes a minikube config to the JSON file
func WriteConfig(configFile string, m MinikubeConfig) error {
	f, err := os.Create(configFile)
	if err != nil {
		return fmt.Errorf("create %s: %s", configFile, err)
	}
	defer f.Close()
	err = encode(f, m)
	if err != nil {
		return fmt.Errorf("encode %s: %s", configFile, err)
	}
	return nil
}

// ReadConfig reads in the JSON minikube config
func ReadConfig(configFile string) (MinikubeConfig, error) {
	f, err := os.Open(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("open %s: %v", localpath.ConfigFile(), err)
	}
	defer f.Close()

	m, err := decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %v", localpath.ConfigFile(), err)
	}

	return m, nil
}

func decode(r io.Reader) (MinikubeConfig, error) {
	var data MinikubeConfig
	err := json.NewDecoder(r).Decode(&data)
	return data, err
}

func encode(w io.Writer, m MinikubeConfig) error {
	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}

	_, err = w.Write(b)

	return err
}

// Load loads the kubernetes and machine config for the current machine
func Load(profile string) (*ClusterConfig, error) {
	return DefaultLoader.LoadConfigFromFile(profile)
}

// Write writes the kubernetes and machine config for the current machine
func Write(profile string, cc *ClusterConfig) error {
	return DefaultLoader.WriteConfigToFile(profile, cc)
}

// Loader loads the kubernetes and machine config based on the machine profile name
type Loader interface {
	LoadConfigFromFile(profile string, miniHome ...string) (*ClusterConfig, error)
	WriteConfigToFile(profileName string, cc *ClusterConfig, miniHome ...string) error
}

type simpleConfigLoader struct{}

// DefaultLoader is the default config loader
var DefaultLoader Loader = &simpleConfigLoader{}

func (c *simpleConfigLoader) LoadConfigFromFile(profileName string, miniHome ...string) (*ClusterConfig, error) {
	var cc ClusterConfig
	// Move to profile package
	path := profileFilePath(profileName, miniHome...)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, &ErrNotExist{fmt.Sprintf("cluster %q does not exist", profileName)}
		}
		return nil, errors.Wrap(err, "stat")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read")
	}

	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	return &cc, nil
}

func (c *simpleConfigLoader) WriteConfigToFile(profileName string, cc *ClusterConfig, miniHome ...string) error {
	// Move to profile package
	path := profileFilePath(profileName, miniHome...)
	contents, err := json.MarshalIndent(cc, "", "	")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, contents, 0644)
}
