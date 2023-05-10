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

package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecording(t *testing.T) {
	recorder := NewHistoryRecorder()
	recorder.Record("foo")
	recorder.Record("bar")
	recorder.Record("qix")
	assert.Equal(t, recorder.History(), []string{"foo", "bar", "qix"})
}

func TestFormattedRecording(t *testing.T) {
	recorder := NewHistoryRecorder()
	recorder.Recordf("%s, %s and %s", "foo", "bar", "qix")
	assert.Equal(t, recorder.History()[0], "foo, bar and qix")
}
