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

package storageclass

import (
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
)

// DisableDefaultStorageClass disables the default storage class provisioner
// The addon-manager and kubectl apply cannot delete storageclasses
func DisableDefaultStorageClass() error {
	client, err := util.GetClientSet()
	if err != nil {
		return err
	}

	err = client.Storage().StorageClasses().Delete(constants.DefaultStorageClassProvisioner, &meta_v1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "Error deleting default storage class %s", constants.DefaultStorageClassProvisioner)
	}

	return nil
}
