/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package service

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1/fake"
	testing_fake "k8s.io/client-go/testing"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/tests"
)

type MockClientGetter struct {
	servicesMap  map[string]typed_core.ServiceInterface
	endpointsMap map[string]typed_core.EndpointsInterface
	secretsMap   map[string]typed_core.SecretInterface
	Fake         fake.FakeCoreV1
}

func init() {
	getCoreClientFail = false
}

// Force GetCoreClient to fail
var getCoreClientFail bool

func (m *MockClientGetter) GetCoreClient() (typed_core.CoreV1Interface, error) {
	if getCoreClientFail {
		return nil, fmt.Errorf("test Error - Get")
	}
	return &MockCoreClient{
		servicesMap:  m.servicesMap,
		endpointsMap: m.endpointsMap,
		secretsMap:   m.secretsMap}, nil
}

func (m *MockClientGetter) GetClientset(timeout time.Duration) (*kubernetes.Clientset, error) {
	return nil, nil
}

// Secrets return fake secret for foo namespace or secrets from mocked client
func (m *MockCoreClient) Secrets(ns string) typed_core.SecretInterface {
	if ns == "foo" {
		return &fake.FakeSecrets{Fake: &fake.FakeCoreV1{Fake: &testing_fake.Fake{}}}
	}
	return m.secretsMap[ns]
}

func (m *MockCoreClient) Services(namespace string) typed_core.ServiceInterface {
	return m.servicesMap[namespace]
}

type MockCoreClient struct {
	fake.FakeCoreV1
	servicesMap  map[string]typed_core.ServiceInterface
	endpointsMap map[string]typed_core.EndpointsInterface
	secretsMap   map[string]typed_core.SecretInterface
}

var secretsNamespaces = map[string]typed_core.SecretInterface{
	"default": defaultNamespaceSecretsInterface,
}

var defaultNamespaceSecretsInterface = &MockSecretInterface{
	SecretsList: &core.SecretList{
		Items: []core.Secret{
			{},
		},
	},
}

var serviceNamespaces = map[string]typed_core.ServiceInterface{
	"default": defaultNamespaceServiceInterface,
}

var defaultNamespaceServiceInterface = &MockServiceInterface{
	ServiceList: &core.ServiceList{
		Items: []core.Service{
			{
				ObjectMeta: meta.ObjectMeta{
					Name:      "mock-dashboard",
					Namespace: "default",
					Labels:    map[string]string{"mock": "mock"},
				},
				Spec: core.ServiceSpec{
					Ports: []core.ServicePort{
						{
							NodePort: int32(1111),
							TargetPort: intstr.IntOrString{
								IntVal: int32(11111),
							},
						},
						{
							NodePort: int32(2222),
							TargetPort: intstr.IntOrString{
								IntVal: int32(22222),
							},
						},
					},
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					Name:      "mock-dashboard-no-ports",
					Namespace: "default",
					Labels:    map[string]string{"mock": "mock"},
				},
				Spec: core.ServiceSpec{
					Ports: []core.ServicePort{},
				},
			},
		},
	},
}

var endpointNamespaces = map[string]typed_core.EndpointsInterface{
	"default": defaultNamespaceEndpointInterface,
}

var defaultNamespaceEndpointInterface = &MockEndpointsInterface{}

func (m *MockCoreClient) Endpoints(namespace string) typed_core.EndpointsInterface {
	return m.endpointsMap[namespace]
}

type MockEndpointsInterface struct {
	fake.FakeEndpoints
	Endpoints *core.Endpoints
}

var endpointMap = map[string]*core.Endpoints{
	"no-subsets": {},
	"not-ready": {
		Subsets: []core.EndpointSubset{
			{
				Addresses: []core.EndpointAddress{},
				NotReadyAddresses: []core.EndpointAddress{
					{IP: "1.1.1.1"},
					{IP: "2.2.2.2"},
				},
			},
		},
	},
	"one-ready": {
		Subsets: []core.EndpointSubset{
			{
				Addresses: []core.EndpointAddress{
					{IP: "1.1.1.1"},
				},
				NotReadyAddresses: []core.EndpointAddress{
					{IP: "2.2.2.2"},
				},
			},
		},
	},
	"mock-dashboard": {
		Subsets: []core.EndpointSubset{
			{
				Ports: []core.EndpointPort{
					{
						Name: "port1",
						Port: int32(11111),
					},
					{
						Name: "port2",
						Port: int32(22222),
					},
				},
			},
		},
	},
}

func (e MockEndpointsInterface) Get(name string, _ meta.GetOptions) (*core.Endpoints, error) {
	endpoint, ok := endpointMap[name]
	if !ok {
		return nil, errors.New("Endpoint not found")
	}
	return endpoint, nil
}

type MockServiceInterface struct {
	fake.FakeServices
	ServiceList *core.ServiceList
}

type MockSecretInterface struct {
	fake.FakeSecrets
	SecretsList *core.SecretList
}

func (s MockServiceInterface) List(opts meta.ListOptions) (*core.ServiceList, error) {
	serviceList := &core.ServiceList{
		Items: []core.Service{},
	}
	if opts.LabelSelector != "" {
		keyValArr := strings.Split(opts.LabelSelector, "=")

		for _, service := range s.ServiceList.Items {
			if service.Spec.Selector[keyValArr[0]] == keyValArr[1] {
				serviceList.Items = append(serviceList.Items, service)
			}
		}

		return serviceList, nil
	}

	return s.ServiceList, nil
}

func (s MockServiceInterface) Get(name string, _ meta.GetOptions) (*core.Service, error) {
	for _, svc := range s.ServiceList.Items {
		if svc.ObjectMeta.Name == name {
			return &svc, nil
		}
	}

	return nil, nil
}

func TestGetServiceListFromServicesByLabel(t *testing.T) {
	serviceList := &core.ServiceList{
		Items: []core.Service{
			{
				Spec: core.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
	}
	serviceIface := MockServiceInterface{
		ServiceList: serviceList,
	}
	if _, err := getServiceListFromServicesByLabel(&serviceIface, "nothing", "nothing"); err != nil {
		t.Fatalf("Service had no label match, but getServiceListFromServicesByLabel returned an error")
	}

	if _, err := getServiceListFromServicesByLabel(&serviceIface, "foo", "bar"); err != nil {
		t.Fatalf("Endpoint was ready with at least one Address, but getServiceListFromServicesByLabel returned an error")
	}
}

func TestPrintURLsForService(t *testing.T) {
	defaultTemplate := template.Must(template.New("svc-template").Parse("http://{{.IP}}:{{.Port}}"))
	client := &MockCoreClient{
		servicesMap:  serviceNamespaces,
		endpointsMap: endpointNamespaces,
	}
	var tests = []struct {
		description    string
		serviceName    string
		namespace      string
		tmpl           *template.Template
		expectedOutput []string
		err            bool
	}{
		{
			description:    "should get all node ports",
			serviceName:    "mock-dashboard",
			namespace:      "default",
			tmpl:           defaultTemplate,
			expectedOutput: []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
		},
		{
			description:    "should get all node ports with arbitrary format",
			serviceName:    "mock-dashboard",
			namespace:      "default",
			tmpl:           template.Must(template.New("svc-arbitrary-template").Parse("{{.IP}}:{{.Port}}")),
			expectedOutput: []string{"127.0.0.1:1111", "127.0.0.1:2222"},
		},
		{
			description:    "should get the name of all target ports with arbitrary format",
			serviceName:    "mock-dashboard",
			namespace:      "default",
			tmpl:           template.Must(template.New("svc-arbitrary-template").Parse("{{.Name}}={{.IP}}:{{.Port}}")),
			expectedOutput: []string{"port1=127.0.0.1:1111", "port2=127.0.0.1:2222"},
		},
		{
			description:    "empty slice for no node ports",
			serviceName:    "mock-dashboard-no-ports",
			namespace:      "default",
			tmpl:           defaultTemplate,
			expectedOutput: []string{},
		},
		{
			description: "throw error without template",
			err:         true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			urls, err := printURLsForService(client, "127.0.0.1", test.serviceName, test.namespace, test.tmpl)
			if err != nil && !test.err {
				t.Errorf("Error: %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Expected error but got none")
			}
			if !reflect.DeepEqual(urls, test.expectedOutput) {
				t.Errorf("\nExpected %v \nActual: %v \n\n", test.expectedOutput, urls)
			}
		})
	}
}

func TestOptionallyHttpsFormattedUrlString(t *testing.T) {

	var tests = []struct {
		description                     string
		bareURLString                   string
		https                           bool
		expectedHTTPSFormattedURLString string
		expectedIsHTTPSchemedURL        bool
	}{
		{
			description:                     "no https for http schemed with no https option",
			bareURLString:                   "http://192.168.99.100:30563",
			https:                           false,
			expectedHTTPSFormattedURLString: "http://192.168.99.100:30563",
			expectedIsHTTPSchemedURL:        true,
		},
		{
			description:                     "no https for non-http schemed with no https option",
			bareURLString:                   "xyz.http.myservice:30563",
			https:                           false,
			expectedHTTPSFormattedURLString: "xyz.http.myservice:30563",
			expectedIsHTTPSchemedURL:        false,
		},
		{
			description:                     "https for http schemed with https option",
			bareURLString:                   "http://192.168.99.100:30563",
			https:                           true,
			expectedHTTPSFormattedURLString: "https://192.168.99.100:30563",
			expectedIsHTTPSchemedURL:        true,
		},
		{
			description:                     "no https for non-http schemed with https option and http substring",
			bareURLString:                   "xyz.http.myservice:30563",
			https:                           true,
			expectedHTTPSFormattedURLString: "xyz.http.myservice:30563",
			expectedIsHTTPSchemedURL:        false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			httpsFormattedURLString, isHTTPSchemedURL := OptionallyHTTPSFormattedURLString(test.bareURLString, test.https)

			if httpsFormattedURLString != test.expectedHTTPSFormattedURLString {
				t.Errorf("\nhttpsFormattedURLString, Expected %v \nActual: %v \n\n", test.expectedHTTPSFormattedURLString, httpsFormattedURLString)
			}

			if isHTTPSchemedURL != test.expectedIsHTTPSchemedURL {
				t.Errorf("\nisHTTPSchemedURL, Expected %v \nActual: %v \n\n",
					test.expectedHTTPSFormattedURLString, httpsFormattedURLString)
			}
		})
	}
}

func TestGetServiceURLs(t *testing.T) {
	defaultAPI := &tests.MockAPI{
		FakeStore: tests.FakeStore{
			Hosts: map[string]*host.Host{
				config.GetMachineName(): {
					Name:   config.GetMachineName(),
					Driver: &tests.MockDriver{},
				},
			},
		},
	}
	defaultTemplate := template.Must(template.New("svc-template").Parse("http://{{.IP}}:{{.Port}}"))

	var tests = []struct {
		description string
		api         libmachine.API
		namespace   string
		expected    URLs
		err         bool
	}{
		{
			description: "no host",
			api: &tests.MockAPI{
				FakeStore: tests.FakeStore{
					Hosts: make(map[string]*host.Host),
				},
			},
			err: true,
		},
		{
			description: "correctly return serviceURLs",
			namespace:   "default",
			api:         defaultAPI,
			expected: []URL{
				{
					Namespace: "default",
					Name:      "mock-dashboard",
					URLs:      []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
				},
				{
					Namespace: "default",
					Name:      "mock-dashboard-no-ports",
					URLs:      []string{},
				},
			},
		},
	}

	defer revertK8sClient(K8s)
	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()

			K8s = &MockClientGetter{
				servicesMap:  serviceNamespaces,
				endpointsMap: endpointNamespaces,
			}
			urls, err := GetServiceURLs(test.api, test.namespace, defaultTemplate)
			if err != nil && !test.err {
				t.Errorf("Error GetServiceURLs %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Test should have failed, but didn't")
			}
			if !reflect.DeepEqual(urls, test.expected) {
				t.Errorf("URLs did not match, expected %v \n\n got %v", test.expected, urls)
			}
		})
	}
}

func TestGetServiceURLsForService(t *testing.T) {
	defaultAPI := &tests.MockAPI{
		FakeStore: tests.FakeStore{
			Hosts: map[string]*host.Host{
				config.GetMachineName(): {
					Name:   config.GetMachineName(),
					Driver: &tests.MockDriver{},
				},
			},
		},
	}
	defaultTemplate := template.Must(template.New("svc-template").Parse("http://{{.IP}}:{{.Port}}"))

	var tests = []struct {
		description string
		api         libmachine.API
		namespace   string
		service     string
		expected    []string
		err         bool
	}{
		{
			description: "no host",
			api: &tests.MockAPI{
				FakeStore: tests.FakeStore{
					Hosts: make(map[string]*host.Host),
				},
			},
			err: true,
		},
		{
			description: "correctly return serviceURLs",
			namespace:   "default",
			service:     "mock-dashboard",
			api:         defaultAPI,
			expected:    []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
		},
		{
			description: "correctly return empty serviceURLs",
			namespace:   "default",
			service:     "mock-dashboard-no-ports",
			api:         defaultAPI,
			expected:    []string{},
		},
	}

	defer revertK8sClient(K8s)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			K8s = &MockClientGetter{
				servicesMap:  serviceNamespaces,
				endpointsMap: endpointNamespaces,
			}
			urls, err := GetServiceURLsForService(test.api, test.namespace, test.service, defaultTemplate)
			if err != nil && !test.err {
				t.Errorf("Error GetServiceURLsForService %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Test should have failed, but didn't")
			}
			if !reflect.DeepEqual(urls, test.expected) {
				t.Errorf("URLs did not match, expected %+v \n\n got %+v", test.expected, urls)
			}
		})
	}
}

func revertK8sClient(k K8sClient) {
	K8s = k
	getCoreClientFail = false
}

func TestGetCoreClient(t *testing.T) {
	k8s := K8sClientGetter{}
	_, err := k8s.GetCoreClient()
	if err != nil {
		t.Fatalf("GetCoreClient returned unexpected error: %v", err)
	}
}

func TestPrintServiceList(t *testing.T) {
	var buf bytes.Buffer
	out := &buf
	input := [][]string{{"foo", "bar", "baz"}}
	PrintServiceList(out, input)
	expected := `|-----------|------|-----|
| NAMESPACE | NAME | URL |
|-----------|------|-----|
| foo       | bar  | baz |
|-----------|------|-----|
`
	got := out.String()
	if got != expected {
		t.Fatalf("PrintServiceList(%v) expected to return %v but got \n%v", input, expected, got)
	}
}

func TestGetServiceListByLabel(t *testing.T) {
	t.Run("failed Get Client", func(t *testing.T) {
		K8s = &MockClientGetter{
			servicesMap:  serviceNamespaces,
			endpointsMap: endpointNamespaces,
		}
		getCoreClientFail = true
		defer revertK8sClient(K8s)
		ns := "foo"
		_, err := GetServiceListByLabel(ns, "mock-dashboard", "foo")
		if err == nil {
			t.Fatal("GetServiceListByLabel expected to fail when GetCoreClient fails")
		}
	})
	t.Run("bad namespace", func(t *testing.T) {
		K8s = &MockClientGetter{
			servicesMap:  serviceNamespaces,
			endpointsMap: endpointNamespaces,
		}
		defer revertK8sClient(K8s)
		ns := "no-such-namespace"
		_, err := GetServiceListByLabel(ns, "mock-dashboard", "foo")
		if err == nil {
			t.Fatalf("GetServiceListByLabel expected to fail for bad namespace: %v", ns)
		}
	})
	t.Run("ok no matches", func(t *testing.T) {
		K8s = &MockClientGetter{
			servicesMap:  serviceNamespaces,
			endpointsMap: endpointNamespaces,
		}
		defer revertK8sClient(K8s)
		ns, svc, label := "default", "mock-dashboard", "foo"
		svcs, err := GetServiceListByLabel(ns, svc, label)
		if err != nil {
			t.Fatalf("GetServiceListByLabel not expected to fail but got err: %v. Namespace %v Service: %v, Label: %v", err, ns, svc, label)
		}
		if len(svcs.Items) == 1 {
			t.Fatalf("GetServiceListByLabel for no matches should return empty list but got: %v", svcs.Items)
		}
	})

	t.Run("ok", func(t *testing.T) {
		K8s = &MockClientGetter{
			servicesMap:  serviceNamespaces,
			endpointsMap: endpointNamespaces,
		}
		defer revertK8sClient(K8s)
		ns, svc, label := "default", "mock-dashboard", "mock"
		svcs, err := GetServiceListByLabel(ns, svc, label)
		if err != nil {
			t.Fatalf("GetServiceListByLabel not expected to fail but got err: %v. Namespace %v Service: %v, Label: %v", err, ns, svc, label)
		}
		if len(svcs.Items) >= 1 {
			t.Fatalf("GetServiceListByLabel for matching data should return at least 1 item but got: %v", svcs.Items)
		}
	})
}

func TestCheckService(t *testing.T) {
	t.Run("svc-no-ports", func(t *testing.T) {
		if err := CheckService("default", "mock-dashboard-no-ports"); err == nil {
			t.Fatal("Error expected for service without ports")
		}
	})
	t.Run("bad namespace", func(t *testing.T) {
		if err := CheckService("no-such-namespace", "foo"); err == nil {
			t.Fatal("Error expected for not existing namespace")
		}
	})
	t.Run("failed Get Client", func(t *testing.T) {
		K8s = &MockClientGetter{
			servicesMap:  serviceNamespaces,
			endpointsMap: endpointNamespaces,
		}
		getCoreClientFail = true
		defer revertK8sClient(K8s)
		if err := CheckService("default", "mock-dashboard"); err == nil {
			t.Fatal("Expected to fail for failed GetCoreClient")
		}
	})
	t.Run("ok", func(t *testing.T) {
		if err := CheckService("default", "mock-dashboard"); err != nil {
			t.Fatalf("Unexpected error: %v while checking existing service", err)
		}
	})
}

func TestDeleteSecret(t *testing.T) {
	t.Run("failed Get Client", func(t *testing.T) {
		ns := "foo"
		K8s = &MockClientGetter{
			servicesMap:  serviceNamespaces,
			endpointsMap: endpointNamespaces,
			secretsMap:   secretsNamespaces,
		}
		getCoreClientFail = true
		defer revertK8sClient(K8s)
		if err := DeleteSecret(ns, "foo"); err == nil {
			t.Fatal("CreateSecret expected to fail for failed GetCoreClient")
		}
	})
	t.Run("ok", func(t *testing.T) {
		ns := "foo"
		secret := "mock-secret"
		K8s = &MockClientGetter{
			servicesMap:  serviceNamespaces,
			endpointsMap: endpointNamespaces,
			secretsMap:   secretsNamespaces,
		}
		defer revertK8sClient(K8s)
		if err := DeleteSecret(ns, secret); err != nil {
			t.Fatalf("DeleteSecret not expected to fail for namespace: %v and secret %v but got err: %v", ns, secret, err)
		}
	})
	t.Run("bad namespace", func(t *testing.T) {
		ns := "no-such-namespace"
		secret := "mock-secret"
		K8s = &MockClientGetter{
			servicesMap:  serviceNamespaces,
			endpointsMap: endpointNamespaces,
			secretsMap:   secretsNamespaces,
		}
		defer revertK8sClient(K8s)
		if err := DeleteSecret(ns, secret); err == nil {
			t.Fatalf("DeleteSecret expected to fail for namespace: %v which doesn't exist", ns)
		}
	})
}

func TestCreateSecret(t *testing.T) {
	t.Run("failed Get Client", func(t *testing.T) {
		ns := "foo"
		K8s = &MockClientGetter{
			servicesMap:  serviceNamespaces,
			endpointsMap: endpointNamespaces,
			secretsMap:   secretsNamespaces,
		}
		getCoreClientFail = true
		defer revertK8sClient(K8s)
		if err := CreateSecret(ns, "foo", map[string]string{"ns": "secret"}, map[string]string{"ns": "baz"}); err == nil {
			t.Fatal("CreateSecret expected to fail for failed GetCoreClient")
		}
	})
	t.Run("bad namespace", func(t *testing.T) {
		ns := "no-such-namespace"
		K8s = &MockClientGetter{
			servicesMap:  serviceNamespaces,
			endpointsMap: endpointNamespaces,
			secretsMap:   secretsNamespaces,
		}
		defer revertK8sClient(K8s)
		if err := CreateSecret(ns, "foo", map[string]string{"ns": "secret"}, map[string]string{"ns": "baz"}); err == nil {
			t.Fatalf("CreateSecret expected to fail for namespace: %v", ns)
		}
	})
	t.Run("ok", func(t *testing.T) {
		ns := "foo"
		K8s = &MockClientGetter{
			servicesMap:  serviceNamespaces,
			endpointsMap: endpointNamespaces,
			secretsMap:   secretsNamespaces,
		}
		defer revertK8sClient(K8s)
		if err := CreateSecret(ns, "foo", map[string]string{"ns": "secret"}, map[string]string{"ns": "baz"}); err != nil {
			t.Fatalf("CreateSecret not expected to fail but got: %v", err)
		}
	})
}

func TestWaitAndMaybeOpenService(t *testing.T) {
	defaultAPI := &tests.MockAPI{
		FakeStore: tests.FakeStore{
			Hosts: map[string]*host.Host{
				config.GetMachineName(): {
					Name:   config.GetMachineName(),
					Driver: &tests.MockDriver{},
				},
			},
		},
	}
	defaultTemplate := template.Must(template.New("svc-template").Parse("http://{{.IP}}:{{.Port}}"))

	var tests = []struct {
		description string
		api         libmachine.API
		namespace   string
		service     string
		expected    []string
		err         bool
	}{
		{
			description: "no host",
			api: &tests.MockAPI{
				FakeStore: tests.FakeStore{
					Hosts: make(map[string]*host.Host),
				},
			},
			err: true,
		},
		{
			description: "correctly return serviceURLs",
			namespace:   "default",
			service:     "mock-dashboard",
			api:         defaultAPI,
			expected:    []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
		},
		{
			description: "correctly return empty serviceURLs",
			namespace:   "default",
			service:     "mock-dashboard-no-ports",
			api:         defaultAPI,
			expected:    []string{},
			err:         true,
		},
	}
	defer revertK8sClient(K8s)
	for _, mode := range []bool{true, false} {
		t.Run(fmt.Sprintf("url mode %v", mode), func(t *testing.T) {
			for _, https := range []bool{true, false} {
				t.Run(fmt.Sprintf("https %v", mode), func(t *testing.T) {

					for _, test := range tests {
						t.Run(test.description, func(t *testing.T) {
							K8s = &MockClientGetter{
								servicesMap:  serviceNamespaces,
								endpointsMap: endpointNamespaces,
							}

							err := WaitAndMaybeOpenService(defaultAPI, test.namespace, test.service, defaultTemplate, mode, https, 1, 0)
							if test.err && err == nil {
								t.Fatalf("WaitAndMaybeOpenService expected to fail for test: %v", test)
							}
							if !test.err && err != nil {
								t.Fatalf("WaitAndMaybeOpenService not expected to fail but got err: %v", err)
							}

						})
					}
				})
			}
		})
	}
}
