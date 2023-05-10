/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package rpcdriver

import (
	"fmt"
	"net/rpc"
	"os/exec"
	"sync"
	"time"

	"io"

	"k8s.io/klog"
	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/drivers/plugin/localbinary"
	"k8s.io/minikube/pkg/libmachine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/libmachine/state"
	"k8s.io/minikube/pkg/libmachine/libmachine/version"
)

var (
	heartbeatInterval = 5 * time.Second
)

type RPCClientDriverFactory interface {
	NewRPCClientDriver(driverName string, rawDriver []byte) (*RPCClientDriver, error)
	io.Closer
}

type DefaultRPCClientDriverFactory struct {
	openedDrivers     []*RPCClientDriver
	openedDriversLock sync.Locker
}

func NewRPCClientDriverFactory() RPCClientDriverFactory {
	return &DefaultRPCClientDriverFactory{
		openedDrivers:     []*RPCClientDriver{},
		openedDriversLock: &sync.Mutex{},
	}
}

type RPCClientDriver struct {
	plugin          localbinary.DriverPlugin
	heartbeatDoneCh chan bool
	Client          *InternalClient
	exec            runner.Runner
}

type RPCCall struct {
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
}

type InternalClient struct {
	MachineName    string
	RPCClient      *rpc.Client
	rpcServiceName string
}

const (
	RPCServiceNameV0 = `RpcServerDriver`
	RPCServiceNameV1 = `RPCServerDriver`

	HeartbeatMethod          = `.Heartbeat`
	GetVersionMethod         = `.GetVersion`
	CloseMethod              = `.Close`
	GetCreateFlagsMethod     = `.GetCreateFlags`
	SetConfigRawMethod       = `.SetConfigRaw`
	GetConfigRawMethod       = `.GetConfigRaw`
	DriverNameMethod         = `.DriverName`
	SetConfigFromFlagsMethod = `.SetConfigFromFlags`
	GetURLMethod             = `.GetURL`
	GetMachineNameMethod     = `.GetMachineName`
	GetIPMethod              = `.GetIP`
	GetSSHHostnameMethod     = `.GetSSHHostname`
	GetSSHKeyPathMethod      = `.GetSSHKeyPath`
	GetSSHPortMethod         = `.GetSSHPort`
	GetSSHUsernameMethod     = `.GetSSHUsername`
	GetMachineStateMethod    = `.GetMachineState`
	PreCreateCheckMethod     = `.PreCreateCheck`
	CreateMachineMethod      = `.CreateMachine`
	RemoveMachineMethod      = `.RemoveMachine`
	StartMachineMethod       = `.StartMachine`
	StopMachineMethod        = `.StopMachine`
	RestartMachineMethod     = `.RestartMachine`
	KillMachineMethod        = `.KillMachine`
	UpgradeMethod            = `.Upgrade`
	IsManagedMethod          = `.IsManaged`
	IsISOBasedMethod         = `.IsISOBased`
	IsContainerBasedMethod   = `.IsContainerBased`
)

func (ic *InternalClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	if serviceMethod != HeartbeatMethod {
		log.Debugf("(%s) Calling %+v", ic.MachineName, serviceMethod)
	}
	return ic.RPCClient.Call(ic.rpcServiceName+serviceMethod, args, reply)
}

func (ic *InternalClient) switchToV0() {
	ic.rpcServiceName = RPCServiceNameV0
}

func NewInternalClient(rpcclient *rpc.Client) *InternalClient {
	return &InternalClient{
		RPCClient:      rpcclient,
		rpcServiceName: RPCServiceNameV1,
	}
}

func (f *DefaultRPCClientDriverFactory) Close() error {
	f.openedDriversLock.Lock()
	defer f.openedDriversLock.Unlock()

	// for _, openedDriver := range f.openedDrivers {
	// 	if err := openedDriver.close(); err != nil {
	// 		// No need to display an error.
	// 		// There's nothing we can do and it doesn't add value to the user.
	// 		// x7TODO: fix this
	// 		// this is a leak...
	// 		// When closing doesn't work, external driver procs will be
	// 		// left hanging.. that causes issues to the next run
	// 	}
	// }
	f.openedDrivers = []*RPCClientDriver{}

	return nil
}

func (f *DefaultRPCClientDriverFactory) NewRPCClientDriver(driverName string, rawDriver []byte) (*RPCClientDriver, error) {
	mcnName := ""

	p, err := localbinary.NewPlugin(driverName)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := p.Serve(); err != nil {
			// TODO: Is this best approach?
			log.Warn(err)
			return
		}
	}()

	addr, err := p.Address()
	if err != nil {
		return nil, fmt.Errorf("error attempting to get plugin server address for RPC: %s", err)
	}

	rpcclient, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, err
	}

	c := &RPCClientDriver{
		Client:          NewInternalClient(rpcclient),
		heartbeatDoneCh: make(chan bool),
	}

	f.openedDriversLock.Lock()
	f.openedDrivers = append(f.openedDrivers, c)
	f.openedDriversLock.Unlock()

	var serverVersion int
	if err := c.Client.Call(GetVersionMethod, struct{}{}, &serverVersion); err != nil {
		// this is the first call we make to the server. We try to play nice with old pre 0.5.1 client,
		// by gracefully trying old RPCServiceName, we do this only once, and keep the result for future calls.
		log.Debugf(err.Error())
		log.Debugf("Client (%s) with %s does not work, re-attempting with %s", c.Client.MachineName, RPCServiceNameV1, RPCServiceNameV0)
		c.Client.switchToV0()
		if err := c.Client.Call(GetVersionMethod, struct{}{}, &serverVersion); err != nil {
			return nil, err
		}
	}

	if serverVersion != version.APIVersion {
		return nil, fmt.Errorf("driver binary uses an incompatible API version (%d)", serverVersion)
	}
	log.Debug("Using API Version ", serverVersion)

	go func(c *RPCClientDriver) {
		for {
			select {
			case <-c.heartbeatDoneCh:
				return
			case <-time.After(heartbeatInterval):
				if err := c.Client.Call(HeartbeatMethod, struct{}{}, nil); err != nil {
					log.Warnf("Wrapper Docker Machine process exiting due to closed plugin server (%s)", err)
					if err := c.close(); err != nil {
						log.Warn(err)
					}
				}
			}
		}
	}(c)

	if err := c.SetConfigRaw(rawDriver); err != nil {
		return nil, err
	}

	mcnName = c.GetMachineName()
	p.MachineName = mcnName
	c.Client.MachineName = mcnName
	c.plugin = p

	return c, nil
}

func (c *RPCClientDriver) MarshalJSON() ([]byte, error) {
	return c.GetConfigRaw()
}

func (c *RPCClientDriver) UnmarshalJSON(data []byte) error {
	return c.SetConfigRaw(data)
}

func (c *RPCClientDriver) close() error {
	c.heartbeatDoneCh <- true
	close(c.heartbeatDoneCh)

	log.Debug("Making call to close driver server")

	if err := c.Client.Call(CloseMethod, struct{}{}, nil); err != nil {
		log.Debugf("Failed to make call to close driver server: %s", err)
	} else {
		log.Debug("Successfully made call to close driver server")
	}

	log.Debug("Making call to close connection to plugin binary")

	return c.plugin.Close()
}

// Helper method to make requests which take no arguments and return simply a
// string, e.g. "GetIP".
func (c *RPCClientDriver) rpcStringCall(method string) (string, error) {
	var info string

	if err := c.Client.Call(method, struct{}{}, &info); err != nil {
		return "", err
	}

	return info, nil
}

func (c *RPCClientDriver) GetCreateFlags() []mcnflag.Flag {
	var flags []mcnflag.Flag

	if err := c.Client.Call(GetCreateFlagsMethod, struct{}{}, &flags); err != nil {
		log.Warnf("Error attempting call to get create flags: %s", err)
	}

	return flags
}

func (c *RPCClientDriver) SetConfigRaw(data []byte) error {
	return c.Client.Call(SetConfigRawMethod, data, nil)
}

func (c *RPCClientDriver) GetConfigRaw() ([]byte, error) {
	var data []byte

	if err := c.Client.Call(GetConfigRawMethod, struct{}{}, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// DriverName returns the name of the driver
func (c *RPCClientDriver) DriverName() string {
	driverName, err := c.rpcStringCall(DriverNameMethod)
	if err != nil {
		log.Warnf("Error attempting call to get driver name: %s", err)
	}

	return driverName
}

func (c *RPCClientDriver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return c.Client.Call(SetConfigFromFlagsMethod, &flags, nil)
}

func (c *RPCClientDriver) GetURL() (string, error) {
	return c.rpcStringCall(GetURLMethod)
}

func (c *RPCClientDriver) GetMachineName() string {
	name, err := c.rpcStringCall(GetMachineNameMethod)
	if err != nil {
		log.Warnf("Error attempting call to get machine name: %s", err)
	}

	return name
}

func (c *RPCClientDriver) GetIP() (string, error) {
	return c.rpcStringCall(GetIPMethod)
}

func (c *RPCClientDriver) GetSSHHostname() (string, error) {
	return c.rpcStringCall(GetSSHHostnameMethod)
}

// GetSSHKeyPath returns the key path
// TODO:  This method doesn't even make sense to have with RPC.
// x7NOTE: don't you worry pal. We're taking care of that.
func (c *RPCClientDriver) GetSSHKeyPath() string {
	path, err := c.rpcStringCall(GetSSHKeyPathMethod)
	if err != nil {
		log.Warnf("Error attempting call to get SSH key path: %s", err)
	}

	return path
}

func (c *RPCClientDriver) GetSSHPort() (int, error) {
	var port int

	if err := c.Client.Call(GetSSHPortMethod, struct{}{}, &port); err != nil {
		return 0, err
	}

	return port, nil
}

func (c *RPCClientDriver) GetSSHUsername() string {
	username, err := c.rpcStringCall(GetSSHUsernameMethod)
	if err != nil {
		log.Warnf("Error attempting call to get SSH username: %s", err)
	}

	return username
}

func (c *RPCClientDriver) GetMachineState() (state.State, error) {
	var s state.State

	if err := c.Client.Call(GetMachineStateMethod, struct{}{}, &s); err != nil {
		return state.Error, err
	}

	return s, nil
}

func (c *RPCClientDriver) PreCreateCheck() error {
	return c.Client.Call(PreCreateCheckMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) CreateMachine() error {
	return c.Client.Call(CreateMachineMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) RemoveMachine() error {
	return c.Client.Call(RemoveMachineMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) StartMachine() error {
	return c.Client.Call(StartMachineMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) StopMachine() error {
	return c.Client.Call(StopMachineMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) RestartMachine() error {
	return c.Client.Call(RestartMachineMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) KillMachine() error {
	return c.Client.Call(KillMachineMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) RunCmd(cmd *exec.Cmd) (*runner.RunResult, error) {
	if c.exec == nil {
		rnr, err := c.GetRunner()
		if err != nil {
			return nil, err
		}

		c.exec = rnr
	}

	return c.exec.RunCmd(cmd)
}

func (c *RPCClientDriver) GetRunner() (runner.Runner, error) {
	ip, err := c.GetIP()
	if err != nil {
		return nil, err
	}

	port, err := c.GetSSHPort()
	if err != nil {
		return nil, err
	}

	sshKPath := c.GetSSHKeyPath()
	sshUsrName := c.GetSSHUsername()

	return runner.NewSSHRunner(ip, sshKPath, sshUsrName, port), nil
}

func (c *RPCClientDriver) IsContainerBased() bool {
	var resp bool
	err := c.Client.Call(IsContainerBasedMethod, &struct{}{}, &resp)
	if err != nil {
		klog.Fatalf("failed to contact external driver to check machine type: %v", err)
	}
	return resp
}

func (c *RPCClientDriver) IsISOBased() bool {
	var resp bool
	err := c.Client.Call(IsISOBasedMethod, &struct{}{}, &resp)
	if err != nil {
		klog.Fatalf("failed to contact external driver to check machine type: %v", err)
	}
	return resp
}

func (c *RPCClientDriver) IsManaged() bool {
	var resp bool
	err := c.Client.Call(IsManagedMethod, &struct{}{}, &resp)
	if err != nil {
		klog.Fatalf("failed to contact external driver to check machine type: %v", err)
	}
	return resp
}
