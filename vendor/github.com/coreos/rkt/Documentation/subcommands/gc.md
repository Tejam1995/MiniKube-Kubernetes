# rkt gc

rkt has a built-in garbage collection command that is designed to be run periodically from a timer or cron job.
Stopped pods are moved to the garbage and cleaned up during a subsequent garbage collection pass.
Each `gc` pass removes any pods remaining in the garbage past the grace period.
[Read more about the pod lifecycle][gc-docs].

[gc-docs]: ../devel/pod-lifecycle.md#garbage-collection

```
# rkt gc --grace-period=30m0s
Moving pod "21b1cb32-c156-4d26-82ae-eda1ab60f595" to garbage
Moving pod "5dd42e9c-7413-49a9-9113-c2a8327d08ab" to garbage
Moving pod "f07a4070-79a9-4db0-ae65-a090c9c393a3" to garbage
```

On the next pass, the pods are removed:

```
# rkt gc --grace-period=30m0s
Garbage collecting pod "21b1cb32-c156-4d26-82ae-eda1ab60f595"
Garbage collecting pod "5dd42e9c-7413-49a9-9113-c2a8327d08ab"
Garbage collecting pod "f07a4070-79a9-4db0-ae65-a090c9c393a3"
```

## Options

| Flag | Default | Options | Description |
| --- | --- | --- | --- |
| `--expire-prepared` |  `24h0m0s` | A time | Duration to wait before expiring prepared pods |
| `--grace-period` |  `30m0s` | A time | Duration to wait before discarding inactive pods from garbage |

## Global options

| Flag | Default | Options | Description |
| --- | --- | --- | --- |
| `--debug` |  `false` | `true` or `false` | Prints out more debug information to `stderr` |
| `--dir` | `/var/lib/rkt` | A directory path | Path to the `rkt` data directory |
| `--insecure-options` |  none | <ul><li>**none**: All security features are enabled</li><li>**http**: Allow HTTP connections. Be warned that this will send any credentials as clear text.</li><li>**image**: Disables verifying image signatures</li><li>**tls**: Accept any certificate from the server and any host name in that certificate</li><li>**ondisk**: Disables verifying the integrity of the on-disk, rendered image before running. This significantly speeds up start time.</li><li>**all**: Disables all security checks</li></ul>  | Comma-separated list of security features to disable |
| `--local-config` |  `/etc/rkt` | A directory path | Path to the local configuration directory |
| `--system-config` |  `/usr/lib/rkt` | A directory path | Path to the system configuration directory |
| `--trust-keys-from-https` |  `false` | `true` or `false` | Automatically trust gpg keys fetched from https |
| `--user-config` |  `` | A directory path | Path to the user configuration directory |
