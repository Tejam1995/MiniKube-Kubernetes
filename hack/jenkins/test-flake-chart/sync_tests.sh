#!/bin/bash

# Copyright 2021 The Kubernetes Authors All rights reserved.
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

# This script is called once per integration test. If all integration tests that
# have registered themselves in the started environment list have also
# registered themselves in the finished environment list, this script reports
# flakes or uploads flakes to flake data.
# 
# This script expects the following env variables:
# MINIKUBE_LOCATION: The Github location being run on (e.g. master, 11000).
# COMMIT: Commit hash the tests ran on.
# ROOT_JOB_ID: Job ID to use for synchronization.

set -o pipefail

BUCKET_PATH="gs://minikube-builds/logs/${MINIKUBE_LOCATION}/${COMMIT:0:7}"
STARTED_LIST=$(gsutil cat "${BUCKET_PATH}/started_environments_${ROOT_JOB_ID}.txt" | sort | uniq)

if [ $? -ne 0 ]; then
  echo "Unable to read environment list. Likely being run before all tests are ready or after tests have already been uploaded." 1>&2
  exit 0
fi

set -eu -o pipefail

FINISHED_LIST=$(mktemp)
gsutil cat "${BUCKET_PATH}/finished_environments_${ROOT_JOB_ID}.txt"\
  | sort\
  | uniq > "${FINISHED_LIST}"

STARTED_COUNT=$(echo "${STARTED_LIST}" | wc -l)
FINISHED_COUNT=$(\
  echo "${STARTED_LIST}"\
  | join - "${FINISHED_LIST}"\
  | wc -l)

if [ ${STARTED_COUNT} -ne ${FINISHED_COUNT} ]; then
  echo "Started environments are not all finished! Started: ${STARTED_LIST}, Finished: $(cat ${FINISHED_LIST}))"
  exit 0
fi

# Prevent other invocations of this script from uploading the same thing multiple times.
gsutil rm "${BUCKET_PATH}/started_environments_${ROOT_JOB_ID}.txt"

# At this point, we know all integration tests are done and we can process all summaries safely.

# Get directory of this script.
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [[ "${MINIKUBE_LOCATION}" == "master" ]]; then
  for ENVIRONMENT in ${STARTED_LIST}; do
    SUMMARY="${BUCKET_PATH}/${ENVIRONMENT}_summary.json"
    "${DIR}/upload_tests.sh" "${SUMMARY}" || true
  done
else
  "${DIR}/report_flakes.sh" "${MINIKUBE_LOCATION}" "${COMMIT:0:7}" "${FINISHED_LIST}"
fi

gsutil rm "${BUCKET_PATH}/finished_environments_${ROOT_JOB_ID}.txt"
rm "${FINISHED_LIST}"
