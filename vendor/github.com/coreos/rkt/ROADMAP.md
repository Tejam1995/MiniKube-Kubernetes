# rkt roadmap

This document defines a high level roadmap for rkt development.
The dates below should not be considered authoritative, but rather indicative of the projected timeline of the project.
The [milestones defined in GitHub](https://github.com/coreos/rkt/milestones) represent the most up-to-date state of affairs.

rkt is an implementation of the [App Container spec](https://github.com/appc/spec), which is still under active development on an approximately similar timeframe.
The version of the spec that rkt implements can be seen in the output of `rkt version`.

rkt's version 1.0 release marks the command line user interface and on-disk data structures as stable and reliable for external development. The (optional) API for pod inspection is not yet completely stabilized, but is quite usable.

### rkt 1.0 (February)
- stable CLI
- usable read-only API
- stable on-disk format (all upgrades should be backwards-compatible)
- different shared namespace execution modes [#1433](https://github.com/coreos/rkt/issues/1433)
- stage1 benchmarking [#1788](https://github.com/coreos/rkt/issues/1788)
- more advanced stage1 image configuration [#1425](https://github.com/coreos/rkt/issues/1425)
- packaged for more distributions
  - Fedora [#1304](https://github.com/coreos/rkt/issues/1304)

### rkt 1.1 (February)

- Enhanced DNS configuration [#2044](https://github.com/coreos/rkt/issues/2044)
- User configuration for stage1 [#2013](https://github.com/coreos/rkt/issues/2013)
- packaged for more distributions
  - Debian [#1307](https://github.com/coreos/rkt/issues/1307)

### rkt 1.2 (March)

- app exit status propagation [#1460](https://github.com/coreos/rkt/issues/1460)
- `rkt fly` as top-level command [#1889](https://github.com/coreos/rkt/issues/1889)
- fully integrated with `machinectl login` and `systemd-run` [#1463](https://github.com/coreos/rkt/issues/1463)
- IPv6 support [appc/cni#31](https://github.com/appc/cni/issues/31)
- packaged for more distributions
  - CentOS [#1305](https://github.com/coreos/rkt/issues/1305)

### rkt 1.3 (March)

- stable API
- full integration with Kubernetes (aka "rktnetes")
- full integration with `machinectl login` and `systemd-run` [#1463](https://github.com/coreos/rkt/issues/1463)
- support for unified cgroup hierarchy [#1757](https://github.com/coreos/rkt/issues/1757)
- attach to the app's stdin/stdout [#1799](https://github.com/coreos/rkt/issues/1799)
