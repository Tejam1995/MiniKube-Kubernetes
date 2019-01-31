// Package cruntime contains code specific to container runtimes
package cruntime

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// CommandRunner is the subset of bootstrapper.CommandRunner this package consumes
type CommandRunner interface {
	Run(string) error
}

// Manager is a common interface for container runtimes
type Manager interface {
	// Name is a human readable name for a runtime
	Name() string
	// Enable idempotently enables this runtime on a host
	Enable(CommandRunner) error
	// Disable idempotently disables this runtime on a host
	Disable(CommandRunner) error
	// Load an image idempotently into the runtime on a host
	LoadImage(CommandRunner, string) error
	// Active returns whether or not a runtime is active on a host
	Active(CommandRunner) bool
	// KubeletOptions returns kubelet options for a runtime.
	KubeletOptions() map[string]string
	// SocketPath returns the path to the socket file for a given runtime
	SocketPath() string
}

// Config is runtime configuration
type Config struct {
	// Type of runtime to create ("docker, "crio", etc)
	Type string
	// Custom path to a socket file
	Socket string
}

// New returns an appropriately configured runtime
func New(c Config) (Manager, error) {
	switch c.Type {
	case "", "docker":
		return &Docker{config: c}, nil
	case "crio", "cri-o":
		return &CRIO{config: c}, nil
	case "rkt":
		return &Rkt{config: c}, nil
	case "containerd":
		return &Containerd{config: c}, nil
	default:
		return nil, fmt.Errorf("unknown runtime type: %q", c.Type)
	}
}

// disableOthers disables all other runtimes except for me.
func disableOthers(me Manager, cr CommandRunner) error {
	// valid values returned by manager.Name()
	runtimes := []string{"containerd", "crio", "docker", "rkt"}
	for _, name := range runtimes {
		r, err := New(Config{Type: name})
		if err != nil {
			return fmt.Errorf("New(%s): %v", name, err)
		}

		// Don't disable myself.
		if r.Name() == me.Name() {
			continue
		}
		// runtime is already disabled, nothing to do.
		if !r.Active(cr) {
			continue
		}
		if err = r.Disable(cr); err != nil {
			glog.Warningf("disable failed: %v", err)
		}
		// Validate that the runtime really is offline - and that Active & Disable are properly written.
		if r.Active(cr) {
			return fmt.Errorf("%s is still active after being disabled!", r.Name())
		}
	}
	return nil
}

// enableIPForwarding configures IP forwarding, which is handled normally by Docker
// Context: https://github.com/kubernetes/kubeadm/issues/1062
func enableIPForwarding(cr CommandRunner) error {
	if err := cr.Run("sudo modprobe br_netfilter"); err != nil {
		return errors.Wrap(err, "br_netfilter")
	}
	if err := cr.Run("sudo sh -c \"echo 1 > /proc/sys/net/ipv4/ip_forward\""); err != nil {
		return errors.Wrap(err, "ip_forward")
	}
	return nil
}
