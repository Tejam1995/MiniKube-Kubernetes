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
	"errors"
	"fmt"
	"io"
	"os"

	"k8s.io/minikube/pkg/minikube/constants"
)

const (
	WantUpdateNotification    = "WantUpdateNotification"
	ReminderWaitPeriodInHours = "ReminderWaitPeriodInHours"
	WantReportError           = "WantReportError"
	WantReportErrorPrompt     = "WantReportErrorPrompt"
)

type configFile interface {
	io.ReadWriter
}

type setFn func(string, string) error
type MinikubeConfig map[string]interface{}

type Setting struct {
	name        string
	set         func(MinikubeConfig, string, string) error
	validations []setFn
	callbacks   []setFn
}

func Get(name string) (string, error) {
	m, err := ReadConfig()
	if err != nil {
		return "", err
	}
	if val, ok := m[name]; ok {
		return fmt.Sprintf("%v", val), nil
	} else {
		return "", errors.New("specified key could not be found in config")
	}
}

// ReadConfig reads in the JSON minikube config
func ReadConfig() (MinikubeConfig, error) {
	f, err := os.Open(constants.ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("Could not open file %s: %s", constants.ConfigFile, err)
	}
	m, err := decode(f)
	if err != nil {
		return nil, fmt.Errorf("Could not decode config %s: %s", constants.ConfigFile, err)
	}

	return m, nil
}

func decode(r io.Reader) (MinikubeConfig, error) {
	var data MinikubeConfig
	err := json.NewDecoder(r).Decode(&data)
	return data, err
}
