/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/util"
)

var addonsScan = &cobra.Command{
	Use:    "scan",
	Short:  "Scans all minikube addon images for security vulnerabilities",
	Long:   "Scans all minikube addon images for security vulnerabilities",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		addonMap, err := download.AddonList()
		if err != nil {
			exit.Error(reason.InternalAddonScan, "getting addon list", err)
		}

		oldStatusMap, err := download.AddonStatus()
		// If we can't find the old status file, we can just proceed
		if err != nil {
			klog.Warningf("failed to get old addon status list: %v", err)
		}

		addonStatus := make(map[string]util.AddonStatus)
		for a, i := range addonMap {
			klog.Infof("scanning addon %s", a)
			images := i.(map[interface{}]interface{})
			status := util.AddonStatus{Enabled: true, Manual: false, ManualReason: "", Images: []util.ImageStatus{}}
			if s, ok := oldStatusMap[a]; ok {
				if !s.Enabled && s.Manual {
					// Addon was manually disabled, respect that
					status = s
					addonStatus[a] = status
					continue
				}
			}
			for _, image := range images {
				klog.Infof("scanning image %s for addon %s", image, a)
				imageStatus := util.ImageStatus{Image: image.(string), CVEs: []util.CVE{}}
				snyk := exec.Command("snyk", "container", "test", image.(string), "--json", "--severity-threshold=high")
				out, err := snyk.Output()
				if err == nil {
					klog.Infof("no vulnerabilities found for %s", image)
					continue
				}
				outmap := make(map[string]interface{})
				err = json.Unmarshal(out, &outmap)
				if err != nil {
					klog.Errorf("error unmarshalling json for %s: %v", image, err)
				}
				// The vulnerabilities entry won't show up if there was an error from snyk
				if vulnz, ok := outmap["vulnerabilities"].([]interface{}); ok {
					for _, v := range vulnz {
						vuln := v.(map[string]interface{})
						logCVE := true
						for _, c := range imageStatus.CVEs {
							if c.Name == vuln["title"].(string) {
								klog.Infof("already logged CVE %s for image %s, skipping", c.Name, image)
								logCVE = false
								break
							}
						}
						if !logCVE {
							continue
						}
						klog.Infof("CVE for %s: %s, %s, %s, %v", image, vuln["title"], vuln["packageName"], vuln["severity"], vuln["nearestFixedInVersion"])
						status.Enabled = false
						updatedVersion := ""
						if uv, ok := vuln["nearestFixedInVersion"].(string); ok {
							updatedVersion = uv
						}
						cve := util.CVE{
							Name:           vuln["title"].(string),
							PackageName:    vuln["packageName"].(string),
							Severity:       vuln["severity"].(string),
							UpdatedVersion: updatedVersion,
						}
						imageStatus.CVEs = append(imageStatus.CVEs, cve)
					}
				}
				if len(imageStatus.CVEs) > 0 {
					status.Images = append(status.Images, imageStatus)
				}
			}
			addonStatus[a] = status
		}
		writeStatusYAML(addonStatus)
	},
}

func init() {
	AddonsCmd.AddCommand(addonsScan)
}

func writeStatusYAML(addonStatus map[string]util.AddonStatus) {
	statusYaml, err := yaml.Marshal(addonStatus)
	if err != nil {
		exit.Error(reason.InternalAddonScan, "marshalling addon status list", err)
	}
	err = os.WriteFile("hack/addons/status.yaml", statusYaml, 0777)
	if err != nil {
		exit.Error(reason.InternalAddonScan, "writing addon status list", err)
	}
}
