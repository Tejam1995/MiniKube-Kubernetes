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

// Expo is expontential backoff retry
func Expo(callback func() error, initInterv time.Duration, maxTime time.Duration) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxTime
	b.InitialInterval = initInterv
	b.RandomizationFactor = 0.5
	b.Multiplier = 1.5
	b.Reset()
	return backoff.Retry(callback, b)
}