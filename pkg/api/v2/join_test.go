package v2_test

import (
	"context"
	"encoding/json"
	"net"
	"reflect"
	"testing"
	"time"

	v2 "github.com/canonical/microk8s-cluster-agent/pkg/api/v2"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
)

// TestJoin tests responses when joining control plane and worker nodes in an existing cluster.
func TestJoin(t *testing.T) {
	s := &mock.Snap{
		DqliteLock: true,
		DqliteCert: "DQLITE CERTIFICATE DATA",
		DqliteKey:  "DQLITE KEY DATA",
		DqliteInfoYaml: `
Address: 10.10.10.10:19001
ID: 1238719276943521
Role: 0
`,
		DqliteClusterYaml: `
- Address: 10.10.10.10:19001
  ID: 1238719276943521
  Role: 0
- Address: 10.10.10.11:19001
  ID: 12312648746587658
  Role: 0
- Address: 10.10.10.100:19001
  ID: 12312648746587655
  Role: 2
`,
		CA:                "CA CERTIFICATE DATA",
		CAKey:             "CA KEY DATA",
		ServiceAccountKey: "SERVICE ACCOUNT KEY DATA",
		ServiceArguments: map[string]string{
			"kubelet":        "kubelet arguments\n",
			"kube-apiserver": "--secure-port 16443\n--authorization-mode=Node,RBAC",
			"cluster-agent":  "--bind=0.0.0.0:25000",
		},
		ClusterTokens:     []string{"worker-token", "control-plane-token"},
		SelfCallbackToken: "callback-token",
		CNIYaml:           `some random content. "first-found"`,
		KnownTokens: map[string]string{
			"admin": "admin-token-123",
		},
	}
	apiv2 := &v2.API{
		Snap: s,
		LookupIP: func(hostname string) ([]net.IP, error) {
			return map[string][]net.IP{
				"test-control-plane": {{10, 10, 10, 13}},
				"test-worker":        {{10, 10, 10, 12}},
			}[hostname], nil
		},
		ListControlPlaneNodeIPs: mockListControlPlaneNodes("10.0.0.1", "10.0.0.2"),
	}
	t.Run("InvalidToken", func(t *testing.T) {
		resp, _, err := apiv2.Join(context.Background(), v2.JoinRequest{ClusterToken: "invalid-token"})
		if err == nil {
			t.Fatalf("Expected error but did not receive any")
		}
		if resp != nil {
			t.Fatalf("Expected a nil response but received %#v", resp)
		}
		if !reflect.DeepEqual(s.ConsumeClusterTokenCalledWith, []string{"invalid-token"}) {
			t.Fatalf("Expected ConsumeClusterToken to be called with %v, but it was called with %v instead", []string{"invalid-token"}, s.ConsumeClusterTokenCalledWith)
		}
	})

	t.Run("ControlPlane", func(t *testing.T) {
		s.ConsumeClusterTokenCalledWith = nil
		resp, _, err := apiv2.Join(context.Background(), v2.JoinRequest{
			ClusterToken:     "control-plane-token",
			RemoteHostName:   "test-control-plane",
			ClusterAgentPort: "25000",
			HostPort:         "10.10.10.10:25000",
			RemoteAddress:    "10.10.10.13:41532",
			WorkerOnly:       false,
		})
		if err != nil {
			t.Fatalf("Expected no errors, but received %q", err)
		}
		if resp == nil {
			t.Fatal("Expected a response but received nil instead")
		}

		expectedResponse := &v2.JoinResponse{
			CertificateAuthority:       "CA CERTIFICATE DATA",
			CallbackToken:              "callback-token",
			APIServerPort:              "16443",
			APIServerAuthorizationMode: "Node,RBAC",
			KubeletArgs:                "kubelet arguments\n",
			HostNameOverride:           "10.10.10.13",
			DqliteVoterNodes:           []string{"10.10.10.10:19001", "10.10.10.11:19001"},
			ServiceAccountKey:          "SERVICE ACCOUNT KEY DATA",
			CertificateAuthorityKey:    func(s string) *string { return &s }("CA KEY DATA"),
			AdminToken:                 "admin-token-123",
			DqliteClusterCertificate:   "DQLITE CERTIFICATE DATA",
			DqliteClusterKey:           "DQLITE KEY DATA",
		}
		if !reflect.DeepEqual(*resp, *expectedResponse) {
			t.Fatalf("Expected response %#v, but received %#v instead", expectedResponse, resp)
		}
		if !reflect.DeepEqual(s.ConsumeClusterTokenCalledWith, []string{"control-plane-token"}) {
			t.Fatalf("Expected ConsumeClusterToken to be called with %v, but it was called with %v instead", []string{"control-plane-token"}, s.ConsumeClusterTokenCalledWith)
		}

		if len(s.ApplyCNICalled) != 1 {
			t.Fatalf("Expected ApplyCNI to be called once, but it was called %d times instead", len(s.ApplyCNICalled))
		}
		if len(s.CreateNoCertsReissueLockCalledWith) != 1 {
			t.Fatalf("Expected CreateNoCertsReissueLock to be called once, but it was called %d times instead", len(s.ApplyCNICalled))
		}
	})

	t.Run("Worker", func(t *testing.T) {
		// Reset
		s.ConsumeClusterTokenCalledWith = nil
		s.ApplyCNICalled = nil
		s.CreateNoCertsReissueLockCalledWith = nil

		resp, _, err := apiv2.Join(context.Background(), v2.JoinRequest{
			ClusterToken:     "worker-token",
			RemoteHostName:   "test-worker",
			RemoteAddress:    "10.10.10.12:31451",
			WorkerOnly:       true,
			HostPort:         "10.10.10.10:25000",
			ClusterAgentPort: "25000",
		})
		if err != nil {
			t.Fatalf("Expected no errors, but received %q", err)
		}
		if resp == nil {
			t.Fatal("Expected a response but received nil instead")
		}
		expectedResponse := &v2.JoinResponse{
			CertificateAuthority:       "CA CERTIFICATE DATA",
			CallbackToken:              "callback-token",
			APIServerAuthorizationMode: "Node,RBAC",
			APIServerPort:              "16443",
			KubeletArgs:                "kubelet arguments\n",
			HostNameOverride:           "10.10.10.12",
			ControlPlaneNodes:          []string{"10.0.0.1", "10.0.0.2"},
		}

		if !reflect.DeepEqual(*resp, *expectedResponse) {
			t.Fatalf("Expected response %#v, but received %#v instead", expectedResponse, resp)
		}
		if !reflect.DeepEqual(s.ConsumeClusterTokenCalledWith, []string{"worker-token"}) {
			t.Fatalf("Expected ConsumeClusterToken to be called with %v, but it was called with %v instead", []string{"worker-token"}, s.ConsumeClusterTokenCalledWith)
		}
		expectedCertRequestTokens := []string{"worker-token-kubelet", "worker-token-proxy"}
		if !reflect.DeepEqual(s.AddCertificateRequestTokenCalledWith, expectedCertRequestTokens) {
			t.Fatalf("Expected the certificate request tokens %v to be created, but %v were created instead", expectedCertRequestTokens, s.AddCertificateRequestTokenCalledWith)
		}
		if len(s.ApplyCNICalled) != 1 {
			t.Fatalf("Expected ApplyCNI to be called once, but it was called %d times instead", len(s.ApplyCNICalled))
		}
		if len(s.CreateNoCertsReissueLockCalledWith) != 1 {
			t.Fatalf("Expected CreateNoCertsReissueLock to be called once, but it was called %d times instead", len(s.ApplyCNICalled))
		}
	})
}

// TestJoinFirstNode tests responses when joining a control plane node on a new cluster.
// TestJoinFirstNode mocks the dqlite bind address update and verifies that is is handled properly.
func TestJoinFirstNode(t *testing.T) {
	s := &mock.Snap{
		DqliteLock: true,
		DqliteCert: "DQLITE CERTIFICATE DATA",
		DqliteKey:  "DQLITE KEY DATA",
		DqliteInfoYaml: `
Address: 127.0.0.1:19001
ID: 1238719276943521
Role: 0
`,
		DqliteClusterYaml: `
- Address: 127.0.0.1:19001
  ID: 1238719276943521
  Role: 0
`,
		CA:                "CA CERTIFICATE DATA",
		CAKey:             "CA KEY DATA",
		ServiceAccountKey: "SERVICE ACCOUNT KEY DATA",
		ServiceArguments: map[string]string{
			"kubelet":        "kubelet arguments\n",
			"kube-apiserver": "--secure-port 16443\n--authorization-mode=Node",
			"cluster-agent":  "--bind=0.0.0.0:25000",
		},
		ClusterTokens:     []string{"control-plane-token"},
		SelfCallbackToken: "callback-token",
		CNIYaml:           `some random content. "first-found"`,
		KnownTokens: map[string]string{
			"admin": "admin-token-123",
		},
	}
	apiv2 := &v2.API{
		Snap: s,
		LookupIP: func(hostname string) ([]net.IP, error) {
			return []net.IP{{10, 10, 10, 13}}, nil
		},
	}

	go func() {
		// update cluster with new address
		<-time.After(500 * time.Millisecond)
		s.DqliteClusterYaml = `
- Address: 10.10.10.10:19001
  ID: 1238719276943521
  Role: 0`
	}()

	resp, _, err := apiv2.Join(context.Background(), v2.JoinRequest{
		ClusterToken:     "control-plane-token",
		RemoteHostName:   "test-worker-nohostname",
		ClusterAgentPort: "25000",
		HostPort:         "10.10.10.10:25000",
		RemoteAddress:    "10.10.10.13:41532",
		WorkerOnly:       false,
	})
	if err != nil {
		t.Fatalf("Expected no errors, but received %q", err)
	}
	if resp == nil {
		t.Fatal("Expected a response but received nil instead")
	}

	expectedResponse := &v2.JoinResponse{
		CertificateAuthority:       "CA CERTIFICATE DATA",
		CallbackToken:              "callback-token",
		APIServerPort:              "16443",
		APIServerAuthorizationMode: "Node",
		KubeletArgs:                "kubelet arguments\n",
		HostNameOverride:           "10.10.10.13",
		DqliteVoterNodes:           []string{"10.10.10.10:19001"},
		ServiceAccountKey:          "SERVICE ACCOUNT KEY DATA",
		CertificateAuthorityKey:    func(s string) *string { return &s }("CA KEY DATA"),
		AdminToken:                 "admin-token-123",
		DqliteClusterCertificate:   "DQLITE CERTIFICATE DATA",
		DqliteClusterKey:           "DQLITE KEY DATA",
	}
	if !reflect.DeepEqual(*resp, *expectedResponse) {
		t.Fatalf("Expected response %#v, but received %#v instead", expectedResponse, resp)
	}
	if !reflect.DeepEqual(s.ConsumeClusterTokenCalledWith, []string{"control-plane-token"}) {
		t.Fatalf("Expected ConsumeClusterToken to be called with %v, but it was called with %v instead", []string{"control-plane-token"}, s.ConsumeClusterTokenCalledWith)
	}
	expectedUpdate := []string{"Address: 10.10.10.10:19001\n"}
	if !reflect.DeepEqual(s.WriteDqliteUpdateYamlCalledWith, expectedUpdate) {
		t.Fatalf("Expected WriteDqliteUpdateYaml to be called with %v, but it was called with %v instead", expectedUpdate, s.WriteDqliteUpdateYamlCalledWith)
	}

	expectedRestart := []string{"k8s-dqlite"}
	if !reflect.DeepEqual(s.RestartServiceCalledWith, expectedRestart) {
		t.Fatalf("Expected the services %v to restart, but %v were restarted instead", expectedRestart, s.RestartServiceCalledWith)
	}
	if len(s.ApplyCNICalled) != 1 {
		t.Fatalf("Expected ApplyCNI to be called once, but it was called %d times instead", len(s.ApplyCNICalled))
	}
	if len(s.CreateNoCertsReissueLockCalledWith) != 1 {
		t.Fatalf("Expected CreateNoCertsReissueLock to be called once, but it was called %d times instead", len(s.ApplyCNICalled))
	}
}

func TestUnmarshalWorkerOnlyField(t *testing.T) {
	for _, tc := range []struct {
		b             string
		expectedValue v2.WorkerOnlyField
	}{
		{b: "true", expectedValue: true},
		{b: "false", expectedValue: false},
		{b: "null", expectedValue: false},
		{b: `"as-worker"`, expectedValue: true},
		{b: `"as-controlplane"`, expectedValue: false},
	} {
		t.Run(tc.b, func(t *testing.T) {
			var v v2.WorkerOnlyField
			if err := json.Unmarshal([]byte(tc.b), &v); err != nil {
				t.Fatalf("Expected no error but received %q", err)
			}
			if v != tc.expectedValue {
				t.Fatalf("Expected value to be %v, but it was %v", tc.expectedValue, v)
			}
		})
	}
}
