package v1_test

import (
	"context"
	"net"
	"testing"

	v1 "github.com/canonical/microk8s-cluster-agent/pkg/api/v1"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestJoin(t *testing.T) {
	s := &mock.Snap{
		CA: "CA CERTIFICATE DATA",
		ServiceArguments: map[string]string{
			"etcd":           "--listen-client-urls=https://0.0.0.0:12379",
			"kube-apiserver": "--secure-port 16443\n--token-auth-file tokens-file",
			"kube-proxy":     "--cluster-cidr 10.1.0.0/16",
			"kubelet":        "kubelet arguments\n",
		},
		ClusterTokens: []string{"valid-cluster-token-cert", "valid-cluster-token-auth", "valid-other-token", "valid-token-for-auth-test"},
		KnownTokens: map[string]string{
			"admin":             "admin-token",
			"system:kube-proxy": "kube-proxy-token",
		},
	}
	apiv1 := &v1.API{
		Snap:     s,
		LookupIP: net.LookupIP,
	}

	t.Run("InvalidToken", func(t *testing.T) {
		g := NewWithT(t)
		resp, err := apiv1.Join(context.Background(), v1.JoinRequest{
			ClusterToken: "invalid-token",
		})
		g.Expect(resp).To(BeNil())
		g.Expect(err).NotTo(BeNil())
		g.Expect(s.ConsumeClusterTokenCalledWith).To(ConsistOf("invalid-token"))
	})

	t.Run("Dqlite", func(t *testing.T) {
		g := NewWithT(t)
		s.DqliteLock = true
		resp, err := apiv1.Join(context.Background(), v1.JoinRequest{
			ClusterToken: "valid-other-token",
		})
		g.Expect(resp).To(BeNil())
		g.Expect(err).NotTo(BeNil())
		s.DqliteLock = false
	})

	t.Run("NoCertAuthNoTokensFile", func(t *testing.T) {
		g := NewWithT(t)
		saveArgs := s.ServiceArguments["kube-apiserver"]
		s.ServiceArguments["kube-apiserver"] = "--secure-port 16443"
		resp, err := apiv1.Join(context.Background(), v1.JoinRequest{
			ClusterToken: "valid-token-for-auth-test",
		})
		g.Expect(resp).To(BeNil())
		g.Expect(err).NotTo(BeNil())
		s.ServiceArguments["kube-apiserver"] = saveArgs
	})

	t.Run("Success", func(t *testing.T) {
		t.Run("CertAuth", func(t *testing.T) {
			g := NewWithT(t)
			s.ConsumeClusterTokenCalledWith = nil
			s.CreateNoCertsReissueLockCalledWith = nil
			s.AddCallbackTokenCalledWith = nil
			s.AddCertificateRequestTokenCalledWith = nil
			resp, err := apiv1.Join(context.Background(), v1.JoinRequest{
				ClusterToken:             "valid-cluster-token-cert",
				HostName:                 "my-hostname",
				ClusterAgentPort:         "25000",
				RemoteAddress:            "10.10.10.10:41422",
				CallbackToken:            "callback-token",
				CanHandleCertificateAuth: true,
			})
			g.Expect(err).To(BeNil())
			g.Expect(resp).To(Equal(&v1.JoinResponse{
				CertificateAuthority: "CA CERTIFICATE DATA",
				EtcdEndpoint:         "https://0.0.0.0:12379",
				APIServerAuthMode:    v1.APIServerAuthModeCert,
				APIServerPort:        "16443",
				KubeProxyToken:       "valid-cluster-token-cert",
				KubeletArgs:          "kubelet arguments\n\n--hostname-override=10.10.10.10",
				KubeletToken:         "valid-cluster-token-cert",
				HostNameOverride:     "10.10.10.10",
				ClusterCIDR:          "10.1.0.0/16",
			}))
			g.Expect(s.ConsumeClusterTokenCalledWith).To(ConsistOf("valid-cluster-token-cert"))
			g.Expect(s.RestartServiceCalledWith).To(BeEmpty())
			g.Expect(s.AddCallbackTokenCalledWith).To(ConsistOf("10.10.10.10:25000 callback-token"))
			g.Expect(s.AddCertificateRequestTokenCalledWith).To(ConsistOf("valid-cluster-token-cert", "valid-cluster-token-cert-kubelet", "valid-cluster-token-cert-proxy"))
			g.Expect(s.CreateNoCertsReissueLockCalledWith).To(HaveLen(1))
		})

		t.Run("TokenAuth", func(t *testing.T) {
			g := NewWithT(t)
			s.ConsumeClusterTokenCalledWith = nil
			s.CreateNoCertsReissueLockCalledWith = nil
			s.AddCallbackTokenCalledWith = nil
			s.AddCertificateRequestTokenCalledWith = nil
			resp, err := apiv1.Join(context.Background(), v1.JoinRequest{
				ClusterToken:             "valid-cluster-token-auth",
				HostName:                 "my-hostname",
				ClusterAgentPort:         "25000",
				RemoteAddress:            "10.10.10.10:41422",
				CallbackToken:            "callback-token",
				CanHandleCertificateAuth: false,
			})
			g.Expect(err).To(BeNil())
			kubeletToken, err := s.GetOrCreateKubeletToken("10.10.10.10")
			g.Expect(err).To(BeNil())
			g.Expect(resp).To(Equal(&v1.JoinResponse{
				CertificateAuthority: "CA CERTIFICATE DATA",
				EtcdEndpoint:         "https://0.0.0.0:12379",
				APIServerAuthMode:    v1.APIServerAuthModeToken,
				APIServerPort:        "16443",
				KubeProxyToken:       "kube-proxy-token",
				KubeletArgs:          "kubelet arguments\n\n--hostname-override=10.10.10.10",
				KubeletToken:         kubeletToken,
				HostNameOverride:     "10.10.10.10",
				ClusterCIDR:          "10.1.0.0/16",
			}))
			g.Expect(s.ConsumeClusterTokenCalledWith).To(ConsistOf("valid-cluster-token-auth"))
			g.Expect(s.RestartServiceCalledWith).To(ConsistOf("apiserver"))
			g.Expect(s.AddCallbackTokenCalledWith).To(ConsistOf("10.10.10.10:25000 callback-token"))
			g.Expect(s.AddCertificateRequestTokenCalledWith).To(ConsistOf("valid-cluster-token-auth"))
			g.Expect(s.CreateNoCertsReissueLockCalledWith).To(HaveLen(1))
		})
	})
}
