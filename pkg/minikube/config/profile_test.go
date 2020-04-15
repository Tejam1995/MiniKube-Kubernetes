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

package config

import (
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/localpath"
)

// TestListProfiles uses a different MINIKUBE_HOME with rest of tests since it relies on file list index
func TestListProfiles(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/default2/")
	if err != nil {
		t.Errorf("error getting dir path for %s : %v",miniDir, err)
	}
	valids, invalids, err := ListProfiles(true, miniDir)

	// test cases for valid profiles
	var valiProfilesCases = []struct {
		index      int
		expectName string
		vmDriver   string
	}{
		{0, "p1", "hyperkit"},
		{1, "p2_newformat", "virtualbox"},
	}
	t.Logf("minidir is %s",miniDir)
	t.Logf("Valid cases length = %d",len(valids))
	for _,x := range(valids) {
		t.Logf("valid : %s",x.Name)
	}
	for _,x := range(invalids) {
		t.Logf("invalid : %s",x.Name)
	}

	t.Logf("inValid cases length = %d",len(invalids))
	for _, vc := range valiProfilesCases {
		if valids[vc.index].Name != vc.expectName {
			t.Errorf("expected %s got %v", vc.expectName, valids[vc.index].Name)
		}
		if valids[vc.index].Config.Driver != vc.vmDriver {
			t.Errorf("expected %s got %v", vc.vmDriver, valids[vc.index].Config.Driver)
		}

	}

	// test cases for invalid profiles
	var invalidProfileCases = []struct {
		index      int
		expectName string
		vmDriver   string
	}{
		{0, "p3_empty", ""},
		{1, "p4_invalid_file", ""},
		{2, "p5_partial_config", ""},
	}

	// making sure it returns the invalid profiles
	for _, tt := range invalidProfileCases {
		if invalids[tt.index].Name != tt.expectName {
			t.Errorf("expected %s got %v", tt.expectName, invalids[tt.index].Name)
		}
	}

	if err != nil {
		t.Errorf("error listing profiles %v", err)
	}
}

func TestProfileNameValid(t *testing.T) {
	var testCases = []struct {
		name     string
		expected bool
	}{
		{"meaningful_name", true},
		{"meaningful_name@", false},
		{"n_a_m_e_2", true},
		{"n", false},
		{"_name", false},
		{"N__a.M--E12567", true},
	}
	for _, tt := range testCases {
		got := ProfileNameValid(tt.name)
		if got != tt.expected {
			t.Errorf("expected ProfileNameValid(%s)=%t but got %t ", tt.name, tt.expected, got)
		}
	}

}

func TestProfileNameInReservedKeywords(t *testing.T) {
	var testCases = []struct {
		name     string
		expected bool
	}{
		{"start", true},
		{"stop", true},
		{"status", true},
		{"delete", true},
		{"config", true},
		{"open", true},
		{"profile", true},
		{"addons", true},
		{"cache", true},
		{"logs", true},
		{"myprofile", false},
		{"log", false},
	}
	for _, tt := range testCases {
		got := ProfileNameInReservedKeywords(tt.name)
		if got != tt.expected {
			t.Errorf("expected ProfileNameInReservedKeywords(%s)=%t but got %t ", tt.name, tt.expected, got)
		}
	}
}

func TestProfileExists(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/.minikube2")
	if err != nil {
		t.Errorf("error getting dir path for ./testdata/.minikube : %v", err)
	}

	var testCases = []struct {
		name     string
		expected bool
	}{
		{"p1", true},
		{"p2_newformat", true},
		{"p3_empty", true},
		{"p4_invalid_file", true},
		{"p5_partial_config", true},
		{"p6_no_file", false},
	}
	for _, tt := range testCases {
		got := ProfileExists(tt.name, miniDir)
		if got != tt.expected {
			t.Errorf("expected ProfileExists(%q,%q)=%t but got %t ", tt.name, miniDir, tt.expected, got)
		}

	}

}

func TestCreateEmptyProfile(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/.minikube2")
	if err != nil {
		t.Errorf("error getting dir path for ./testdata/.minikube : %v", err)
	}

	var testCases = []struct {
		name      string
		expectErr bool
	}{
		{"p13", false},
		{"p_13", false},
	}
	for _, tc := range testCases {
		n := tc.name // capturing  loop variable
		gotErr := CreateEmptyProfile(n, miniDir)
		if gotErr != nil && tc.expectErr == false {
			t.Errorf("expected CreateEmptyProfile not to error but got err=%v", gotErr)
		}

		defer func() { // tear down
			err := DeleteProfile(n, miniDir)
			if err != nil {
				t.Errorf("error test tear down %v", err)
			}
		}()

	}

}

func TestCreateProfile(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/.minikube2")
	if err != nil {
		t.Errorf("error getting dir path for ./testdata/.minikube : %v", err)
	}

	var testCases = []struct {
		name      string
		cfg       *ClusterConfig
		expectErr bool
	}{
		{"p_empty_config", &ClusterConfig{}, false},
		{"p_partial_config", &ClusterConfig{KubernetesConfig: KubernetesConfig{
			ShouldLoadCachedImages: false}}, false},
		{"p_partial_config2", &ClusterConfig{
			KeepContext: false, KubernetesConfig: KubernetesConfig{
				ShouldLoadCachedImages: false}}, false},
	}
	for _, tc := range testCases {
		n := tc.name // capturing  loop variable
		gotErr := SaveProfile(n, tc.cfg, miniDir)
		if gotErr != nil && tc.expectErr == false {
			t.Errorf("expected CreateEmptyProfile not to error but got err=%v", gotErr)
		}

		defer func() { // tear down

			err := DeleteProfile(n, miniDir)
			if err != nil {
				t.Errorf("error test tear down %v", err)
			}
		}()
	}

}

func TestDeleteProfile(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/.minikube2")
	if err != nil {
		t.Errorf("error getting dir path for ./testdata/.minikube : %v", err)
	}

	err = CreateEmptyProfile("existing_prof", miniDir)
	if err != nil {
		t.Errorf("error setting up TestDeleteProfile %v", err)
	}

	var testCases = []struct {
		name      string
		expectErr bool
	}{
		{"existing_prof", false},
		{"non_existing_prof", false},
	}
	for _, tc := range testCases {
		gotErr := DeleteProfile(tc.name, miniDir)
		if gotErr != nil && tc.expectErr == false {
			t.Errorf("expected CreateEmptyProfile not to error but got err=%v", gotErr)
		}
	}

}

func TestGetPrimaryControlPlane(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/.minikube2")
	if err != nil {
		t.Errorf("error getting dir path for ./testdata/.minikube : %v", err)
	}

	var tests = []struct {
		description  string
		profile      string
		expectedIP   string
		expectedPort int
		expectedName string
	}{
		{"old style", "p1", "192.168.64.75", 8443, "minikube"},
		{"new style", "p2_newformat", "192.168.99.136", 8443, "m01"},
	}

	for _, tc := range tests {
		cc, err := DefaultLoader.LoadConfigFromFile(tc.profile, miniDir)
		if err != nil {
			t.Fatalf("Failed to load config for %s", tc.description)
		}

		n, err := PrimaryControlPlane(cc)
		if err != nil {
			t.Fatalf("Unexpexted error getting primary control plane: %v", err)
		}

		if n.Name != tc.expectedName {
			t.Errorf("Unexpected name. expected: %s, got: %s", tc.expectedName, n.Name)
		}

		if n.IP != tc.expectedIP {
			t.Errorf("Unexpected name. expected: %s, got: %s", tc.expectedIP, n.IP)
		}

		if n.Port != tc.expectedPort {
			t.Errorf("Unexpected name. expected: %d, got: %d", tc.expectedPort, n.Port)
		}

	}

}

func TestProfileDirs(t *testing.T) {
	miniHome, err := filepath.Abs("./testdata/delete-all")
	if err != nil {
		t.Errorf("error getting dir path for ./testdata/.minikube : %v", err)
	}

	dirs, err := profileDirs(miniHome)
	if err != nil {
		t.Errorf("error profileDirs: %v", err)
	}
	if len(dirs) != 8 {
		t.Errorf("expected length of dirs to be %d but got %d:", 12, len(dirs))
	}

}

func TestProfileFilePath(t *testing.T) {
	var testsCases = []struct {
		profile  string
		miniHome string
		expected string
	}{
		{"p1", "/var/T/all6479", "/var/T/all6479/.minikube/profiles/p1/config.json"},
		{"p1_underscore", "/var/T/all6479", "/var/T/all6479/.minikube/profiles/p1_underscore/config.json"},
	}

	for _, tc := range testsCases {
		p := profileFilePath(tc.profile, tc.miniHome)
		if p != tc.expected {
			t.Errorf("expected profile file path to be %q but got %q", tc.expected, p)
		}
	}

	// trying one without specifying minihome
	p := profileFilePath("p3")
	expected := filepath.Join(localpath.MiniPath(),"profiles","p3","config.json")
	if p != expected {
		t.Errorf("expected profile file path to be %q but got %q", expected, p)
	}


}
