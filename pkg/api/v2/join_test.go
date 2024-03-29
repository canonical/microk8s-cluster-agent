package v2_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	v2 "github.com/canonical/microk8s-cluster-agent/pkg/api/v2"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
	utiltest "github.com/canonical/microk8s-cluster-agent/pkg/util/test"
	. "github.com/onsi/gomega"
)

// TestJoin tests responses when joining control plane and worker nodes in an existing cluster.
func TestJoin(t *testing.T) {
	cni := `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`
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
			"kube-apiserver": "--secure-port 16443\n--authorization-mode=Node,RBAC\n--etcd-servers=${SNAP_DATA}/var/kubernetes/backend/kine.sock:12379\n",
			"kube-proxy":     "--cluster-cidr 10.1.0.0/16",
			"cluster-agent":  "--bind=0.0.0.0:25000",
		},
		ClusterTokens:     []string{"worker-token", "control-plane-token", "valid-token-for-auth-test"},
		SelfCallbackToken: "callback-token",
		CNIYaml:           cni,
		KnownTokens: map[string]string{
			"admin": "admin-token-123",
		},
	}
	apiv2 := &v2.API{
		Snap: s,
		LookupIP: func(hostname string) ([]net.IP, error) {
			return map[string][]net.IP{
				"test-no-cert":       {{10, 10, 10, 14}},
				"test-control-plane": {{10, 10, 10, 13}},
				"test-worker":        {{10, 10, 10, 12}},
			}[hostname], nil
		},
		InterfaceAddrs: func() ([]net.Addr, error) {
			return []net.Addr{
				&utiltest.MockCIDR{CIDR: "127.0.0.1/8"},
				&utiltest.MockCIDR{CIDR: "10.0.0.10/16"},
			}, nil
		},
		ListControlPlaneNodeIPs: mockListControlPlaneNodes("10.0.0.1", "10.0.0.2"),
	}
	t.Run("InvalidToken", func(t *testing.T) {
		g := NewWithT(t)
		resp, _, err := apiv2.Join(context.Background(), v2.JoinRequest{ClusterToken: "invalid-token"})
		g.Expect(err).NotTo(BeNil())
		g.Expect(resp).To(BeNil())
		g.Expect(s.ConsumeClusterTokenCalledWith).To(ConsistOf("invalid-token"))
	})

	t.Run("NoCertAuthNoTokensFile", func(t *testing.T) {
		g := NewWithT(t)
		saveArgs := s.ServiceArguments["kube-apiserver"]
		s.ServiceArguments["kube-apiserver"] = "--etcd-servers=${SNAP_DATA}/var/kubernetes/backend/kine.sock:12379\n--secure-port 16443"
		resp, _, err := apiv2.Join(context.Background(), v2.JoinRequest{
			ClusterToken:     "valid-token-for-auth-test",
			ClusterAgentPort: "25000",
			RemoteHostName:   "test-no-cert",
			HostPort:         "10.10.10.10:25000",
			RemoteAddress:    "10.10.10.14:31312",
		})
		g.Expect(resp).To(BeNil())
		g.Expect(err).NotTo(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("requires x509 authentication"))
		s.ServiceArguments["kube-apiserver"] = saveArgs
	})

	t.Run("ControlPlane", func(t *testing.T) {
		g := NewWithT(t)
		s.ConsumeClusterTokenCalledWith = nil
		s.CreateNoCertsReissueLockCalledWith = nil
		resp, _, err := apiv2.Join(context.Background(), v2.JoinRequest{
			ClusterToken:             "control-plane-token",
			RemoteHostName:           "test-control-plane",
			ClusterAgentPort:         "25000",
			HostPort:                 "10.10.10.10:25000",
			RemoteAddress:            "10.10.10.13:41532",
			WorkerOnly:               false,
			CanHandleCertificateAuth: true,
		})
		g.Expect(err).To(BeNil())
		g.Expect(resp).NotTo(BeNil())

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
			DqliteClusterCertificate:   "DQLITE CERTIFICATE DATA",
			DqliteClusterKey:           "DQLITE KEY DATA",
			ClusterCIDR:                "10.1.0.0/16",
		}
		g.Expect(resp).To(Equal(expectedResponse))
		g.Expect(s.ConsumeClusterTokenCalledWith).To(ConsistOf("control-plane-token"))
		g.Expect(s.ApplyCNICalled).To(HaveLen(1))
		g.Expect(s.CreateNoCertsReissueLockCalledWith).To(HaveLen(1))
	})

	t.Run("Worker", func(t *testing.T) {
		g := NewWithT(t)

		// Reset
		s.ConsumeClusterTokenCalledWith = nil
		s.ApplyCNICalled = nil
		s.CNIYaml = cni
		s.CreateNoCertsReissueLockCalledWith = nil

		resp, _, err := apiv2.Join(context.Background(), v2.JoinRequest{
			ClusterToken:     "worker-token",
			RemoteHostName:   "test-worker",
			RemoteAddress:    "10.10.10.12:31451",
			WorkerOnly:       true,
			HostPort:         "10.10.10.10:25000",
			ClusterAgentPort: "25000",
		})
		g.Expect(err).To(BeNil())
		g.Expect(resp).NotTo(BeNil())

		expectedResponse := &v2.JoinResponse{
			CertificateAuthority:       "CA CERTIFICATE DATA",
			CallbackToken:              "callback-token",
			APIServerAuthorizationMode: "Node,RBAC",
			APIServerPort:              "16443",
			KubeletArgs:                "kubelet arguments\n",
			HostNameOverride:           "10.10.10.12",
			ControlPlaneNodes:          []string{"10.0.0.1", "10.0.0.2"},
			ClusterCIDR:                "10.1.0.0/16",
		}

		g.Expect(resp).To(Equal(expectedResponse))
		g.Expect(s.ConsumeClusterTokenCalledWith).To(ConsistOf("worker-token"))
		g.Expect(s.ApplyCNICalled).To(HaveLen(1))
		g.Expect(s.CreateNoCertsReissueLockCalledWith).To(HaveLen(1))
		g.Expect(s.AddCertificateRequestTokenCalledWith).To(ConsistOf("worker-token-kubelet", "worker-token-proxy"))
	})
}

// TestJoinFirstNode tests responses when joining a control plane node on a new cluster.
// TestJoinFirstNode mocks the dqlite bind address update and verifies that is is handled properly.
func TestJoinFirstNode(t *testing.T) {
	g := NewWithT(t)

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
			"kube-apiserver": "--secure-port 16443\n--authorization-mode=Node\n--token-auth-file=known_tokens.csv\n--etcd-servers=${SNAP_DATA}/var/kubernetes/backend/kine.sock:12379\n",
			"kube-proxy":     "--cluster-cidr 10.1.0.0/16",
			"cluster-agent":  "--bind=0.0.0.0:25000",
		},
		ClusterTokens:     []string{"control-plane-token"},
		SelfCallbackToken: "callback-token",
		CNIYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
		KnownTokens: map[string]string{
			"admin": "admin-token-123",
		},
	}
	apiv2 := &v2.API{
		Snap: s,
		LookupIP: func(hostname string) ([]net.IP, error) {
			return []net.IP{{10, 10, 10, 13}}, nil
		},
		InterfaceAddrs: func() ([]net.Addr, error) {
			return []net.Addr{
				&utiltest.MockCIDR{CIDR: "127.0.0.1/8"},
				&utiltest.MockCIDR{CIDR: "10.10.10.10/16"},
			}, nil
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
	g.Expect(err).To(BeNil())
	g.Expect(resp).NotTo(BeNil())

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
		ClusterCIDR:                "10.1.0.0/16",
	}
	g.Expect(resp).To(Equal(expectedResponse))
	g.Expect(s.ConsumeClusterTokenCalledWith).To(ConsistOf("control-plane-token"))
	g.Expect(s.ApplyCNICalled).To(HaveLen(1))
	g.Expect(s.CreateNoCertsReissueLockCalledWith).To(HaveLen(1))
	g.Expect(s.WriteDqliteUpdateYamlCalledWith).To(ConsistOf("Address: 10.10.10.10:19001\n"))
	g.Expect(s.RestartServiceCalledWith).To(ConsistOf("k8s-dqlite"))
}

// TestJoinWithoutDNSResolution tests that node joins are not rejected when the remote hostname does not resolve, but InternalIP is used for kubelet communication.
func TestJoinWithoutDNSResolution(t *testing.T) {
	g := NewWithT(t)

	s := &mock.Snap{
		DqliteLock: true,
		DqliteCert: "DQLITE CERTIFICATE DATA",
		DqliteKey:  "DQLITE KEY DATA",
		DqliteInfoYaml: `
Address: 10.10.10.10:19001
ID: 1238719276943521
Role: 0`,
		DqliteClusterYaml: `
- Address: 10.10.10.10:19001
  ID: 1238719276943521
  Role: 0`,
		CA:                "CA CERTIFICATE DATA",
		CAKey:             "CA KEY DATA",
		ServiceAccountKey: "SERVICE ACCOUNT KEY DATA",
		ServiceArguments: map[string]string{
			"kubelet":        "kubelet arguments\n",
			"kube-apiserver": "--secure-port 16443\n--authorization-mode=Node\n--kubelet-preferred-address-types=InternalIP,Hostname\n--token-auth-file=known_tokens.csv\n--etcd-servers=${SNAP_DATA}/var/kubernetes/backend/kine.sock:12379\n",
			"kube-proxy":     "--cluster-cidr 10.1.0.0/16",
			"cluster-agent":  "--bind=0.0.0.0:25000",
		},
		ClusterTokens:     []string{"control-plane-token"},
		SelfCallbackToken: "callback-token",
		CNIYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
		KnownTokens: map[string]string{
			"admin": "admin-token-123",
		},
	}
	apiv2 := &v2.API{
		Snap: s,
		LookupIP: func(hostname string) ([]net.IP, error) {
			return nil, fmt.Errorf("no DNS resolution")
		},
		InterfaceAddrs: func() ([]net.Addr, error) {
			return []net.Addr{
				&utiltest.MockCIDR{CIDR: "127.0.0.1/8"},
				&utiltest.MockCIDR{CIDR: "10.10.10.10/16"},
			}, nil
		},
	}

	resp, _, err := apiv2.Join(context.Background(), v2.JoinRequest{
		ClusterToken:     "control-plane-token",
		RemoteHostName:   "test-worker-nohostname",
		ClusterAgentPort: "25000",
		HostPort:         "10.10.10.10:25000",
		RemoteAddress:    "10.10.10.13:41532",
		WorkerOnly:       false,
	})
	g.Expect(err).To(BeNil())
	g.Expect(resp).NotTo(BeNil())
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
		ClusterCIDR:                "10.1.0.0/16",
	}
	g.Expect(resp).To(Equal(expectedResponse))
	g.Expect(s.ConsumeClusterTokenCalledWith).To(ConsistOf("control-plane-token"))
	g.Expect(s.ApplyCNICalled).To(HaveLen(1))
	g.Expect(s.CreateNoCertsReissueLockCalledWith).To(HaveLen(1))
}

func TestJoinCustomEtcdEndpoints(t *testing.T) {
	s := &mock.Snap{
		DqliteLock: true,
		EtcdCA:     "ETCD CA DATA",
		EtcdCert:   "ETCD CERTIFICATE DATA",
		EtcdKey:    "ETCD KEY DATA",
		ServiceArguments: map[string]string{
			"kubelet":        "kubelet arguments\n",
			"kube-apiserver": "--secure-port 16443\n--etcd-servers=https://etcd1:2379,https://etcd2:2379\n",
			"kube-proxy":     "--cluster-cidr 10.1.0.0/16",
			"cluster-agent":  "--bind=0.0.0.0:25000",
		},
		ClusterTokens: []string{"token-1", "token-2"},
	}

	apiv2 := &v2.API{
		Snap: s,
		LookupIP: func(hostname string) ([]net.IP, error) {
			return map[string][]net.IP{
				"test-1": {{10, 10, 10, 11}},
				"test-2": {{10, 10, 10, 12}},
			}[hostname], nil
		},
		InterfaceAddrs: func() ([]net.Addr, error) {
			return []net.Addr{
				&utiltest.MockCIDR{CIDR: "127.0.0.1/8"},
				&utiltest.MockCIDR{CIDR: "10.10.10.10/16"},
			}, nil
		},
	}

	t.Run("NotSupportedByClient", func(t *testing.T) {
		g := NewWithT(t)
		resp, _, err := apiv2.Join(context.Background(), v2.JoinRequest{
			ClusterToken:             "token-1",
			RemoteHostName:           "test-1",
			ClusterAgentPort:         "25000",
			HostPort:                 "10.10.10.10:25000",
			RemoteAddress:            "10.10.10.11:41532",
			WorkerOnly:               false,
			CanHandleCertificateAuth: true,
		})
		g.Expect(err).NotTo(BeNil())
		g.Expect(resp).To(BeNil())
	})

	t.Run("SupportedByClient", func(t *testing.T) {
		g := NewWithT(t)
		resp, _, err := apiv2.Join(context.Background(), v2.JoinRequest{
			ClusterToken:             "token-2",
			RemoteHostName:           "test-2",
			ClusterAgentPort:         "25000",
			HostPort:                 "10.10.10.10:25000",
			RemoteAddress:            "10.10.10.12:41532",
			WorkerOnly:               false,
			CanHandleCustomEtcd:      true,
			CanHandleCertificateAuth: true,
		})
		g.Expect(err).To(BeNil())
		g.Expect(resp.EtcdServers).To(Equal("https://etcd1:2379,https://etcd2:2379"))
		g.Expect(resp.EtcdCertificateAuthority).To(Equal("ETCD CA DATA"))
		g.Expect(resp.EtcdClientCertificate).To(Equal("ETCD CERTIFICATE DATA"))
		g.Expect(resp.EtcdClientKey).To(Equal("ETCD KEY DATA"))
		g.Expect(resp.DqliteClusterCertificate).To(BeEmpty())
		g.Expect(resp.DqliteClusterKey).To(BeEmpty())
		g.Expect(resp.DqliteVoterNodes).To(BeEmpty())
	})
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
			g := NewWithT(t)
			var v v2.WorkerOnlyField
			err := json.Unmarshal([]byte(tc.b), &v)
			g.Expect(err).To(BeNil())
			g.Expect(v).To(Equal(tc.expectedValue))
		})
	}
}
