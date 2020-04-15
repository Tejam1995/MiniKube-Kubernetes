#!/bin/bash

# Copyright 2016 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# This script runs the integration tests on an OSX machine for the Hyperkit Driver

# The script expects the following env variables:
# MINIKUBE_LOCATION: GIT_COMMIT from upstream build.
# COMMIT: Actual commit ID from upstream build
# EXTRA_BUILD_ARGS (optional): Extra args to be passed into the minikube integrations tests
# access_token: The Github API access token. Injected by the Jenkins credential provider. 


set -e

OS_ARCH="darwin-amd64"
VM_DRIVER="docker"
JOB_NAME="Docker_macOS"
EXTRA_START_ARGS=""
EXPECTED_DEFAULT_DRIVER="hyperkit"

osascript -e 'quit app "Docker"'
sudo /Applications/Docker.app/Contents/MacOS/Docker --quit-after-install --unattended || true
# repeating above command without sudo because  https://github.com/docker/for-mac/issues/882#issuecomment-516946766
/Applications/Docker.app/Contents/MacOS/Docker --quit-after-install --unattended || true

# restart docker on mac for a fresh test
osascript -e 'quit app "Docker"'; open -a Docker /Applications/Docker.app/Contents/MacOS/Docker --quit-after-install --unattended  ; while [ -z "$(docker info 2> /dev/null )" ]; do printf "."; sleep 1; done; echo "" || true

mkdir -p cron && gsutil -qm rsync "gs://minikube-builds/${MINIKUBE_LOCATION}/cron" cron || echo "FAILED TO GET CRON FILES"
install cron/cleanup_and_reboot_Darwin.sh $HOME/cleanup_and_reboot.sh || echo "FAILED TO INSTALL CLEANUP"
echo "*/30 * * * * $HOME/cleanup_and_reboot.sh" | crontab
crontab -l

source common.sh
