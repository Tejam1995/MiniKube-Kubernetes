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

package kubeconfig

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubernetes/pkg/scheduler/api/latest"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	pkgutil "k8s.io/minikube/pkg/util"
)

// Update sets up kubeconfig to be used by kubectl
func Update(h *host.Host, c *cfg.Config) *Setup {
	addr, err := h.Driver.GetURL()
	if err != nil {
		exit.WithError("Failed to get driver URL", err)
	}
	addr = strings.Replace(addr, "tcp://", "https://", -1)
	addr = strings.Replace(addr, ":2376", ":"+strconv.Itoa(c.KubernetesConfig.NodePort), -1)
	if c.KubernetesConfig.APIServerName != constants.APIServerName {
		addr = strings.Replace(addr, c.KubernetesConfig.NodeIP, c.KubernetesConfig.APIServerName, -1)
	}

	kcs := &Setup{
		ClusterName:          cfg.GetMachineName(),
		ClusterServerAddress: addr,
		ClientCertificate:    constants.MakeMiniPath("client.crt"),
		ClientKey:            constants.MakeMiniPath("client.key"),
		CertificateAuthority: constants.MakeMiniPath("ca.crt"),
		KeepContext:          viper.GetBool("keep-context"),
		EmbedCerts:           viper.GetBool("embed-certs"),
	}
	kcs.setPath(Path())
	if err := update(kcs); err != nil {
		exit.WithError("Failed to setup kubeconfig", err)
	}
	return kcs
}

// update reads config from disk, adds the minikube settings, and writes it back.
// activeContext is true when minikube is the CurrentContext
// If no CurrentContext is set, the given name will be used.
func update(cfg *Setup) error {
	glog.Infoln("Using kubeconfig: ", cfg.fileContent())

	// read existing config or create new if does not exist
	config, err := readOrNew(cfg.fileContent())
	if err != nil {
		return err
	}

	err = Populate(cfg, config)
	if err != nil {
		return err
	}

	// write back to disk
	if err := writeToFile(config, cfg.fileContent()); err != nil {
		return errors.Wrap(err, "writing kubeconfig")
	}
	return nil
}

// UpdateIP overwrites the IP stored in kubeconfig with the provided IP.
func UpdateIP(ip net.IP, filename string, machineName string) (bool, error) {
	if ip == nil {
		return false, fmt.Errorf("error, empty ip passed")
	}
	kip, err := extractIP(filename, machineName)
	if err != nil {
		return false, err
	}
	if kip.Equal(ip) {
		return false, nil
	}
	kport, err := Port(filename, machineName)
	if err != nil {
		return false, err
	}
	con, err := readOrNew(filename)
	if err != nil {
		return false, errors.Wrap(err, "Error getting kubeconfig status")
	}
	// Safe to lookup server because if field non-existent getIPFromKubeconfig would have given an error
	con.Clusters[machineName].Server = "https://" + ip.String() + ":" + strconv.Itoa(kport)
	err = writeToFile(con, filename)
	if err != nil {
		return false, err
	}
	// Kubeconfig IP reconfigured
	return true, nil
}

// writeToFile encodes the configuration and writes it to the given file.
// If the file exists, it's contents will be overwritten.
func writeToFile(config runtime.Object, filename string) error {
	if config == nil {
		glog.Errorf("could not write to '%s': config can't be nil", filename)
	}

	// encode config to YAML
	data, err := runtime.Encode(latest.Codec, config)
	if err != nil {
		return errors.Errorf("could not write to '%s': failed to encode config: %v", filename, err)
	}

	// create parent dir if doesn't exist
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrapf(err, "Error creating directory: %s", dir)
		}
	}

	// write with restricted permissions
	if err := ioutil.WriteFile(filename, data, 0600); err != nil {
		return errors.Wrapf(err, "Error writing file %s", filename)
	}
	if err := pkgutil.MaybeChownDirRecursiveToMinikubeUser(dir); err != nil {
		return errors.Wrapf(err, "Error recursively changing ownership for dir: %s", dir)
	}

	return nil
}

// readOrNew retrieves Kubernetes client configuration from a file.
// If no files exists, an empty configuration is returned.
func readOrNew(filename string) (*api.Config, error) {
	data, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		return api.NewConfig(), nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "Error reading file %q", filename)
	}

	// decode config, empty if no bytes
	config, err := decode(data)
	if err != nil {
		return nil, errors.Errorf("could not read config: %v", err)
	}

	// initialize nil maps
	if config.AuthInfos == nil {
		config.AuthInfos = map[string]*api.AuthInfo{}
	}
	if config.Clusters == nil {
		config.Clusters = map[string]*api.Cluster{}
	}
	if config.Contexts == nil {
		config.Contexts = map[string]*api.Context{}
	}

	return config, nil
}

// extractIP returns the IP address stored for minikube in the kubeconfig specified
func extractIP(filename, machineName string) (net.IP, error) {
	con, err := readOrNew(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting kubeconfig status")
	}
	cluster, ok := con.Clusters[machineName]
	if !ok {
		return nil, errors.Errorf("Kubeconfig does not have a record of the machine cluster")
	}
	kurl, err := url.Parse(cluster.Server)
	if err != nil {
		return net.ParseIP(cluster.Server), nil
	}
	kip, _, err := net.SplitHostPort(kurl.Host)
	if err != nil {
		return net.ParseIP(kurl.Host), nil
	}
	ip := net.ParseIP(kip)
	return ip, nil
}

// decode reads a Config object from bytes.
// Returns empty config if no bytes.
func decode(data []byte) (*api.Config, error) {
	// if no data, return empty config
	if len(data) == 0 {
		return api.NewConfig(), nil
	}

	config, _, err := latest.Codec.Decode(data, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Error decoding config from data: %s", string(data))
	}

	return config.(*api.Config), nil
}
