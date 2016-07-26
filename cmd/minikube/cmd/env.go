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

package cmd

import (
	"fmt"
	"os"

	"github.com/docker/machine/libmachine"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
)

var (
	isUnsetMode bool
)

// envCmd represents the docker-env command
var dockerEnvCmd = &cobra.Command{
	Use:   "docker-env",
	Short: "sets up docker env variables; similar to '$(docker-machine env)'",
	Long:  `sets up docker env variables; similar to '$(docker-machine env)'`,
	Run: func(cmd *cobra.Command, args []string) {
		if isUnsetMode {
			fmt.Fprintln(os.Stdout, buildUnsetShellOutput())
			return
		}

		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		envMap, err := cluster.GetHostDockerEnv(api)
		if err != nil {
			glog.Errorln("Error setting machine env variable(s):", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, buildDockerEnvShellOutput(envMap))
	},
}

func buildDockerEnvShellOutput(envMap map[string]string) string {
	output := ""
	for env_name, env_val := range envMap {
		output += fmt.Sprintf("export %s=%s\n", env_name, env_val)
	}
	howToRun := "# Run this command to configure your shell: \n# eval $(minikube docker-env)"
	output += howToRun
	return output
}

func buildUnsetShellOutput() string {
	envArr := []string{"DOCKER_TLS_VERIFY", "DOCKER_HOST", "DOCKER_CERT_PATH"}
	output := ""
	for _, envName := range envArr {
		output += fmt.Sprintf("unset %s\n", envName)
	}
	howToRun := "# Run this command to unset docker-env variables: \n# eval $(minikube docker-env --unset)"
	output += howToRun
	return output
}

func init() {
	dockerEnvCmd.Flags().BoolVar(&isUnsetMode, "unset", false, "Specifies that `minikube docker-env` should return the commands to unset the docker environment variables")

	RootCmd.AddCommand(dockerEnvCmd)
}
