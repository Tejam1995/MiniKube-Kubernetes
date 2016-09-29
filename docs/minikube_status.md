## minikube status

Gets the status of a local kubernetes cluster.

### Synopsis


Gets the status of a local kubernetes cluster.

```
minikube status
```

### Options

```
      --format string   Go template format string for the status output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
For the list accessible variables for the template, see the struct values here:https://godoc.org/k8s.io/minikube/cmd/minikube/cmd#Status (default "minikubeVM: {{.MinikubeStatus}}\nlocalkube: {{.LocalkubeStatus}}\n")
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --log-flush-frequency duration     Maximum number of seconds between log flushes (default 5s)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --show-libmachine-logs             Whether or not to show logs from libmachine.
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minikube](minikube.md)	 - Minikube is a tool for managing local Kubernetes clusters.

