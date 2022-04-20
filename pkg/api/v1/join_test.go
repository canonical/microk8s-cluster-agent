package v1_test

import (
	"context"
	"net"
	"reflect"
	"testing"

	v1 "github.com/canonical/microk8s-cluster-agent/pkg/api/v1"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
)

func TestJoin(t *testing.T) {
	s := &mock.Snap{
		DataStore: snap.SingleNodeEtcdDataStore,
		CA:        "CA CERTIFICATE DATA",
		ServiceArguments: map[string]string{
			"etcd":           "--listen-client-urls=https://0.0.0.0:12379",
			"kube-apiserver": "--secure-port 16443",
			"kubelet":        "kubelet arguments\n",
		},
		ClusterTokens: []string{"valid-cluster-token", "valid-other-token"},
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
		resp, err := apiv1.Join(context.Background(), v1.JoinRequest{
			ClusterToken: "invalid-token",
		})
		if resp != nil {
			t.Fatalf("Expected a nil response due to invalid token, but got %#v\n", resp)
		}
		if err == nil {
			t.Fatal("Expected an error due to invalid token, but did not get any")
		}
		if !reflect.DeepEqual(s.ConsumeClusterTokenCalledWith, []string{"invalid-token"}) {
			t.Fatalf("Expected ConsumeClusterToken to be called with %v, but it was called with %v instead", []string{"invalid-token"}, s.ConsumeClusterTokenCalledWith)
		}
	})

	t.Run("Dqlite", func(t *testing.T) {
		s.DataStore = snap.DqliteDataStore
		resp, err := apiv1.Join(context.Background(), v1.JoinRequest{
			ClusterToken: "valid-other-token",
		})
		if resp != nil {
			t.Fatalf("Expected a nil response due to kubelite lock, but got %#v\n", resp)
		}
		if err == nil {
			t.Fatal("Expected an error due to kubelite lock, but did not get any")
		}
		s.DataStore = snap.SingleNodeEtcdDataStore
	})

	t.Run("Success", func(t *testing.T) {
		s.ConsumeClusterTokenCalledWith = nil
		resp, err := apiv1.Join(context.Background(), v1.JoinRequest{
			ClusterToken:     "valid-cluster-token",
			HostName:         "my-hostname",
			ClusterAgentPort: "25000",
			RemoteAddress:    "10.10.10.10:41422",
			CallbackToken:    "callback-token",
		})
		if err != nil {
			t.Fatalf("Expected no errors, but got %s", err)
		}
		if resp == nil {
			t.Fatal("Expected non-nil response")
		}
		expectedResponse := &v1.JoinResponse{
			CertificateAuthority: "CA CERTIFICATE DATA",
			EtcdEndpoint:         "https://0.0.0.0:12379",
			APIServerPort:        "16443",
			KubeProxyToken:       "kube-proxy-token",
			KubeletArgs:          "kubelet arguments\n\n--hostname-override=10.10.10.10",
			KubeletToken:         resp.KubeletToken,
			HostNameOverride:     "10.10.10.10",
		}
		if *resp != *expectedResponse {
			t.Fatalf("Expected response %#v, but it was %#v", expectedResponse, resp)
		}
		if len(resp.KubeletToken) != 32 {
			t.Fatalf("Expected kubelet token %q to have length 32", resp.KubeletToken)
		}
		if !reflect.DeepEqual(s.ConsumeClusterTokenCalledWith, []string{"valid-cluster-token"}) {
			t.Fatalf("Expected ConsumeClusterToken to be called with %v, but it was called with %v instead", []string{"valid-cluster-token"}, s.ConsumeClusterTokenCalledWith)
		}
		if !reflect.DeepEqual(s.RestartServiceCalledWith, []string{"apiserver"}) {
			t.Fatalf("Expected API server restart command, but got %v instead", s.RestartServiceCalledWith)
		}

		kubeletToken, err := s.GetOrCreateKubeletToken("10.10.10.10")
		if err != nil {
			t.Fatalf("Expected no error when retrieving kubelet token, but received %q", err)
		}
		if kubeletToken != resp.KubeletToken {
			t.Fatalf("Expected kubelet known token to match response, but they do not (%q != %q)", kubeletToken, resp.KubeletToken)
		}

		if !reflect.DeepEqual(s.AddCallbackTokenCalledWith, []string{"10.10.10.10:25000 callback-token"}) {
			t.Fatal("Expected callback-token to be a valid callback token, but it is not")
		}
		if !reflect.DeepEqual(s.AddCertificateRequestTokenCalledWith, []string{"valid-cluster-token"}) {
			t.Fatal("Expected valid-cluster-token to be a valid certificate request token, but it is not")
		}
		if len(s.CreateNoCertsReissueLockCalledWith) != 1 {
			t.Fatal("Expected certificate reissue lock to be in place after successful join, but it is not")
		}
	})
}
