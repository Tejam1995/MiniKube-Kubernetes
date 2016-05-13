# rkt cat-manifest

For debugging or inspection you may want to extract the PodManifest to stdout.

```
# rkt cat-manifest --pretty-print UUID
{
  "acVersion":"0.7.0",
  "acKind":"PodManifest"
...
```

## Options

| Flag | Default | Options | Description |
| --- | --- | --- | --- |
| `--pretty-print` |  `false` | `true` or `false` | Apply indent to format the output |

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
