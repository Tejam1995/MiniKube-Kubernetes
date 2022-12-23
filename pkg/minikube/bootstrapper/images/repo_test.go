/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package images

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/google/go-cmp/cmp"
)

func Test_kubernetesRepo(t *testing.T) {
	kv := semver.MustParse("1.23.0")
	tests := []struct {
		mirror  string
		version semver.Version
		want    string
	}{
		{
			"",
			kv,
			DefaultKubernetesRepo(kv),
		},
		{
			"mirror.k8s.io",
			kv,
			"mirror.k8s.io",
		},
		{
			"",
			// all versions -1.21
			semver.MustParse("1.21.0"),
			OldDefaultKubernetesRepo,
		},
		{
			"",
			semver.MustParse("1.22.17"),
			OldDefaultKubernetesRepo,
		},
		{
			"",
			semver.MustParse("1.22.18"),
			NewDefaultKubernetesRepo,
		},
		{
			"",
			semver.MustParse("1.23.14"),
			OldDefaultKubernetesRepo,
		},
		{
			"",
			semver.MustParse("1.23.15"),
			NewDefaultKubernetesRepo,
		},
		{
			"",
			semver.MustParse("1.24.8"),
			OldDefaultKubernetesRepo,
		},
		{
			"",
			semver.MustParse("1.24.9"),
			NewDefaultKubernetesRepo,
		},
		{
			"",
			// all versions 1.25+
			semver.MustParse("1.25.0"),
			NewDefaultKubernetesRepo,
		},
	}
	for _, tc := range tests {
		got := kubernetesRepo(tc.mirror, tc.version)
		if !cmp.Equal(got, tc.want) {
			t.Errorf("mirror mismatch for v%s, want: %s, got: %s", tc.version, tc.want, got)
		}
	}

}
