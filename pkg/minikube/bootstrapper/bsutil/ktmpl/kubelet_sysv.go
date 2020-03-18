/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package ktmpl

RestartWrapperTemplate = template.Must(template.New("kubeletSysVTemplate").Parse(`#!/bin/bash
# Wrapper script to emulate systemd restart on non-systemd systems
binary=$1
unitfile=$2
args=""

while [[ -x "${binary}" ]]; do
  if [[ -f "${unitfile}" ]]; then
          args=$(egrep "^ExecStart=${binary}" "${unitfile}" | cut -d" " -f2-)
  fi
  ${binary} ${args}
  sleep 1
done
```

KubeletSysVTemplate = template.Must(template.New("kubeletSysVTemplate").Parse(`#!/bin/sh
# SysV style init script for kubelet

readonly KUBELET={{.KubeletPath}}
readonly KUBELET_WRAPPER={{.KubeletWrapperPath}}
readonly KUBELET_PIDFILE="/var/run/kubelet.pid"
readonly KUBELET_LOGFILE=/var/run/nohup.out

if [[ ! -x "${KUBELET}" ]]; then
	echo "$KUBELET not present or not executable"
	exit 1
fi

function start() {
    cd /var/run
    nohup "${KUBELET_WRAPPER}" &
    echo $! > "${KUBELET_PIDFILE}"
}

function stop() {
    if [[ -f "${KUBELET_PIDFILE}" ]]; then
        kill $(cat ${KUBELET_PIDFILE})
    fi
    pkill "${KUBELET_WRAPPER}"
    pkill kubelet
}


case "$1" in
    start)
        start
		;;

    stop)
        stop
		;;

    restart)
        stop
        start
		;;

    status)
        pgrep kubelet
		;;

	*)
		echo "Usage: service kubelet {start|stop|restart|status}"
		exit 1
		;;
esac
`)