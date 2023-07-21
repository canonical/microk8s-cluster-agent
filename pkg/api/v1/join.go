package v1

import (
	"context"
	"fmt"
	"net"

	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
)

// JoinRequest is the request data for the join API endpoint.
type JoinRequest struct {
	// ClusterToken is the token generated during "microk8s add-node".
	ClusterToken string `json:"token"`
	// HostName is the hostname of the joining host.
	HostName string `json:"hostname"`
	// ClusterAgentPort is the port number where the cluster-agent is listening on the joining node.
	ClusterAgentPort string `json:"port"`
	// CallbackToken is a token that this node can use to authenticate with the cluster-agent on the joining node.
	CallbackToken string `json:"callback"`
	// RemoteAddress is the remote address from which the join request originates. This is retrieved directly from the *http.Request object.
	RemoteAddress string `json:"-"`
}

// JoinResponse is the response data for the join API endpoint.
type JoinResponse struct {
	// CertificateAuthority is the root CertificateAuthority certificate for the Kubernetes cluster.
	CertificateAuthority string `json:"ca"`
	// EtcdEndpoint is the endpoint where the etcd server is running, typically https://0.0.0.0:12379.
	// Note that "0.0.0.0" in the response will be replaced with the IP used to join the new node.
	EtcdEndpoint string `json:"etcd"`
	// APIServerAuthenticationMode is the AuthenticationMode used by kube-apiserver.
	APIServerAuthenticationMode string `json:"api_authn_mode"`
	// APIServerPort is the port where the kube-apiserver is listening.
	APIServerPort string `json:"apiport"`
	// KubeProxyToken is a token used to authenticate kube-proxy on the joining node.
	KubeProxyToken string `json:"kubeproxy"`
	// KubeletToken is a token used to authenticate kubelet on the joining node.
	KubeletToken string `json:"kubelet"`
	// KubeletArgs is a string with arguments for the kubelet service on the joining node.
	KubeletArgs string `json:"kubelet_args"`
	// HostNameOverride is the host name the joining node will be known as in the MicroK8s cluster.
	HostNameOverride string `json:"hostname_override"`
	// ClusterCIDR is the cidr that is used by the cluster, defined in kube-proxy args.
	ClusterCIDR string `json:"cluster_cidr,omitempty"`
}

// Join implements "POST /CLUSTER_API_V1/join".
func (a *API) Join(ctx context.Context, request JoinRequest) (*JoinResponse, error) {

	response := &JoinResponse{
		EtcdEndpoint:  snaputil.GetServiceArgument(a.Snap, "etcd", "--listen-client-urls"),
		APIServerPort: snaputil.GetServiceArgument(a.Snap, "kube-apiserver", "--secure-port"),
		ClusterCIDR:   snaputil.GetServiceArgument(a.Snap, "kube-proxy", "--cluster-cidr"),
	}

	if !a.Snap.ConsumeClusterToken(request.ClusterToken) {
		return nil, fmt.Errorf("invalid token")
	}

	if a.Snap.HasDqliteLock() {
		return nil, fmt.Errorf("failed to join the cluster. This is an HA MicroK8s cluster.\nPlease retry after enabling HA on this joining node with 'microk8s enable ha-cluster'")
	}

	if err := a.Snap.AddCertificateRequestToken(request.ClusterToken); err != nil {
		return nil, fmt.Errorf("failed to add certificate request token: %w", err)
	}
	hostname := util.GetRemoteHost(a.LookupIP, request.HostName, request.RemoteAddress)
	clusterAgentEndpoint := net.JoinHostPort(hostname, request.ClusterAgentPort)

	if err := a.Snap.AddCallbackToken(clusterAgentEndpoint, request.CallbackToken); err != nil {
		return nil, fmt.Errorf("failed to add callback token for %s: %w", clusterAgentEndpoint, err)
	}

	ca, err := a.Snap.ReadCA()
	if err != nil {
		return nil, fmt.Errorf("failed to read cluster CA: %w", err)
	}
	response.CertificateAuthority = ca

	if snaputil.GetServiceArgument(a.Snap, "kube-apiserver", "--token-auth-file") != "" {
		response.APIServerAuthenticationMode = "Token"
		response.KubeProxyToken, err = a.Snap.GetKnownToken("system:kube-proxy")
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve kube-proxy token: %w", err)
		}
		response.KubeletToken, err = a.Snap.GetOrCreateKubeletToken(hostname)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve kubelet token: %w", err)
		}
	} else {
		response.APIServerAuthenticationMode = "Cert"
		response.KubeProxyToken = request.ClusterToken
		if err := a.Snap.AddCertificateRequestToken(fmt.Sprintf("%s-proxy", request.ClusterToken)); err != nil {
			return nil, fmt.Errorf("failed adding certificate request token for kube-proxy: %w", err)
		}
		response.KubeletToken = request.ClusterToken
		if err := a.Snap.AddCertificateRequestToken(fmt.Sprintf("%s-kubelet", request.ClusterToken)); err != nil {
			return nil, fmt.Errorf("failed adding certificate request token for kubelet: %w", err)
		}
	}
	if err := a.Snap.RestartService(ctx, "apiserver"); err != nil {
		return nil, fmt.Errorf("failed to restart apiserver service: %w", err)
	}
	response.KubeletArgs, err = a.Snap.ReadServiceArguments("kubelet")
	if err != nil {
		return nil, fmt.Errorf("failed to read arguments of kubelet service: %w", err)
	}
	if hostname != request.HostName {
		response.KubeletArgs = fmt.Sprintf("%s\n--hostname-override=%s", response.KubeletArgs, hostname)
	}
	response.HostNameOverride = hostname
	if err := a.Snap.CreateNoCertsReissueLock(); err != nil {
		return nil, fmt.Errorf("failed to create lock file to disable certificate reissuing: %w", err)
	}
	return response, nil
}
