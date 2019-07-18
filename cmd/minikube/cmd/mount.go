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
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	cmdUtil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/third_party/go9p/ufs"
)

// nineP is the value of --type used for the 9p filesystem.
const nineP = "9p"

// placeholders for flag values
var mountIP string
var mountVersion string
var mountType string
var isKill bool
var uid string
var gid string
var mSize int
var options []string
var mode uint

// supportedFilesystems is a map of filesystem types to not warn against.
var supportedFilesystems = map[string]bool{nineP: true}

// mountCmd represents the mount command
var mountCmd = &cobra.Command{
	Use:   "mount [flags] <source directory>:<target directory>",
	Short: "Mounts the specified directory into minikube",
	Long:  `Mounts the specified directory into minikube.`,
	Run: func(cmd *cobra.Command, args []string) {
		if isKill {
			if err := cmdUtil.KillMountProcess(); err != nil {
				exit.WithError("Error killing mount process", err)
			}
			os.Exit(0)
		}

		if len(args) != 1 {
			exit.Usage(`Please specify the directory to be mounted: 
	minikube mount <source directory>:<target directory>   (example: "/host-home:/vm-home")`)
		}
		mountString := args[0]
		idx := strings.LastIndex(mountString, ":")
		if idx == -1 { // no ":" was present
			exit.UsageT(`mount argument "{{.value}}" must be in form: <source directory>:<target directory>`, console.Arg{"value": amountString})
		}
		hostPath := mountString[:idx]
		vmPath := mountString[idx+1:]
		if _, err := os.Stat(hostPath); err != nil {
			if os.IsNotExist(err) {
				exit.WithCodeT(exit.NoInput, "Cannot find directory {{.path}} for mount", console.Arg{"path": hostPath})
			} else {
				exit.WithError("stat failed", err)
			}
		}
		if len(vmPath) == 0 || !strings.HasPrefix(vmPath, "/") {
			exit.UsageT("Target directory {{.path}} must be an absolute path", console.Arg{"path": vmPath})
		}
		var debugVal int
		if glog.V(1) {
			debugVal = 1 // ufs.StartServer takes int debug param
		}
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()
		host, err := api.Load(config.GetMachineName())

		if err != nil {
			exit.WithError("Error loading api", err)
		}
		if host.Driver.DriverName() == constants.DriverNone {
			exit.Usage(`'none' driver does not support 'minikube mount' command`)
		}
		var ip net.IP
		if mountIP == "" {
			ip, err = cluster.GetVMHostIP(host)
			if err != nil {
				exit.WithError("Error getting the host IP address to use from within the VM", err)
			}
		} else {
			ip = net.ParseIP(mountIP)
			if ip == nil {
				exit.WithCode(exit.Data, "error parsing the input ip address for mount")
			}
		}
		port, err := cmdUtil.GetPort()
		if err != nil {
			exit.WithError("Error finding port for mount", err)
		}

		cfg := &cluster.MountConfig{
			Type:    mountType,
			UID:     uid,
			GID:     gid,
			Version: mountVersion,
			MSize:   mSize,
			Port:    port,
			Mode:    os.FileMode(mode),
			Options: map[string]string{},
		}

		for _, o := range options {
			if !strings.Contains(o, "=") {
				cfg.Options[o] = ""
				continue
			}
			parts := strings.Split(o, "=")
			cfg.Options[parts[0]] = parts[1]
		}

		console.OutT(console.Mounting, "Mounting host path {{.sourcePath}} into VM as {{.destinationPath}} ...", console.Arg{"sourcePath": hostPath, "destinationPath": vmPath})
		console.OutT(console.Option, "Mount type:   {{.name}}", console.Arg{"type": cfg.Type})
		console.OutT(console.Option, "User ID:      {{.userID}}", console.Arg{"userID", cfg.UID})
		console.OutT(console.Option, "Group ID:     {{.groupID}}", console.Arg{"groupID", cfg.GID})
		console.OutT(console.Option, "Version:      {{.version}}", console.Arg{"version", cfg.Version})
		console.OutT(console.Option, "Message Size: {{.size}}", console.Arg{"size", cfg.MSize})
		console.OutT(console.Option, "Permissions:  {{.octalMode}} ({{.writtenMode}})", cfg.Mode, cfg.Mode)
		console.OutT(console.Option, "Options:      {{.options}}", cfg.Options)

		// An escape valve to allow future hackers to try NFS, VirtFS, or other FS types.
		if !supportedFilesystems[cfg.Type] {
			console.OutT(console.WarningType, "{{.type}} is not yet a supported filesystem. We will try anyways!", console.Arg{"type": cfg.Type})
		}

		var wg sync.WaitGroup
		if cfg.Type == nineP {
			wg.Add(1)
			go func() {
				console.OutT(console.Fileserver, "Userspace file server: ")
				ufs.StartServer(net.JoinHostPort(ip.String(), strconv.Itoa(port)), debugVal, hostPath)
				console.OutT(console.Stopped, "Userspace file server is shutdown")
				wg.Done()
			}()
		}

		// Use CommandRunner, as the native docker ssh service dies when Ctrl-C is received.
		runner, err := machine.CommandRunner(host)
		if err != nil {
			exit.WithError("Failed to get command runner", err)
		}

		// Unmount if Ctrl-C or kill request is received.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			for sig := range c {
				console.OutT(console.Unmount, "Unmounting {{.path}} ...", console.Arg{"path": vmPath})
				err := cluster.Unmount(runner, vmPath)
				if err != nil {
					console.ErrT(console.FailureType, "Failed unmount: {{.error}}", console.Arg{"error": err})
				}
				exit.WithCode(exit.Interrupted, "Received {{.name}} signal", console.Arg{"name": sig})
			}
		}()

		err = cluster.Mount(runner, ip.String(), vmPath, cfg)
		if err != nil {
			exit.WithError("mount failed", err)
		}
		console.OutT(console.SuccessType, "Successfully mounted {{.sourcePath}} to {{.destinationPath}}", console.Arg{"sourcePath": hostPath, "destinationPath": vmPath})
		console.OutLn("")
		console.OutT(console.Notice, "NOTE: This process must stay alive for the mount to be accessible ...")
		wg.Wait()
	},
}

func init() {
	mountCmd.Flags().StringVar(&mountIP, "ip", "", "Specify the ip that the mount should be setup on")
	mountCmd.Flags().StringVar(&mountType, "type", nineP, "Specify the mount filesystem type (supported types: 9p)")
	mountCmd.Flags().StringVar(&mountVersion, "9p-version", constants.DefaultMountVersion, "Specify the 9p version that the mount should use")
	mountCmd.Flags().BoolVar(&isKill, "kill", false, "Kill the mount process spawned by minikube start")
	mountCmd.Flags().StringVar(&uid, "uid", "docker", "Default user id used for the mount")
	mountCmd.Flags().StringVar(&gid, "gid", "docker", "Default group id used for the mount")
	mountCmd.Flags().UintVar(&mode, "mode", 0755, "File permissions used for the mount")
	mountCmd.Flags().StringSliceVar(&options, "options", []string{}, "Additional mount options, such as cache=fscache")
	mountCmd.Flags().IntVar(&mSize, "msize", constants.DefaultMsize, "The number of bytes to use for 9p packet payload")
	RootCmd.AddCommand(mountCmd)
}
