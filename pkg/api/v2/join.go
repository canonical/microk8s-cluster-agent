package v2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
)

// WorkerOnlyField is the "worker" field of the JoinRequest message.
type WorkerOnlyField bool

// UnmarshalJSON implements json.Unmarshaler.
// It handles boolean values, as well as the string value "as-worker".
func (v *WorkerOnlyField) UnmarshalJSON(b []byte) error {
	*v = WorkerOnlyField(bytes.Equal(b, []byte("true")) || bytes.Equal(b, []byte(`"as-worker"`)))
	return nil
}

// JoinRequest is the request message for the v2/join API endpoint.
type JoinRequest struct {
	// ClusterToken is the token generated during "microk8s add-node".
	ClusterToken string `json:"token"`
	// RemoteHostName is the hostname of the joining host.
	RemoteHostName string `json:"hostname"`
	// ClusterAgentPort is the port number where the cluster-agent is listening on the joining node.
	ClusterAgentPort string `json:"port"`
	// WorkerOnly is true when joining a worker-only node.
	WorkerOnly WorkerOnlyField `json:"worker"`
	// HostPort is the hostname and port that accepted the request. This is retrieved directly from the *http.Request object.
	HostPort string `json:"-"`
	// RemoteAddress is the remote address from which the join request originates. This is retrieved directly from the *http.Request object.
	RemoteAddress string `json:"-"`
}

// JoinResponse is the response message for the v2/join API endpoint.
type JoinResponse struct {
	// CertificateAuthority is the root CertificateAuthority certificate for the Kubernetes cluster.
	CertificateAuthority string `json:"ca"`
	// CallbackToken is a callback token used to authenticate requests with the cluster agent on the joining node.
	CallbackToken string `json:"callback_token"`
	// APIServerPort is the port where the kube-apiserver is listening.
	APIServerPort string `json:"apiport"`
	// APIServerAuthorizationMode is the AuthorizationMode used by kube-apiserver.
	APIServerAuthorizationMode string `json:"api_authz_mode"`
	// KubeletArgs is a string with arguments for the kubelet service on the joining node.
	KubeletArgs string `json:"kubelet_args"`
	// HostNameOverride is the host name the joining node will be known as in the MicroK8s cluster.
	HostNameOverride string `json:"hostname_override"`
	// DqliteVoterNodes is a list of known dqlite voter nodes. Each voter is identified as "$IP_ADDRESS:$PORT".
	// This is not included in the response when joining worker-only nodes.
	DqliteVoterNodes []string `json:"voters,omitempty"`
	// ServiceAccountKey is the private key used for signing ServiceAccount tokens.
	// This is not included in the response when joining worker-only nodes.
	ServiceAccountKey string `json:"service_account_key"`
	// AdminToken is a static token used to authenticate in the MicroK8s cluster as "admin".
	// This is not included in the response when joining worker-only nodes.
	AdminToken string `json:"admin_token,omitempty"`
	// CertificateAuthorityKey is the private key of the Certificate Authority.
	// Note that this is defined as *string, since the Python code expects this to be explicitly None/nil/null for worker only nodes.
	// This is not included in the response when joining worker-only nodes.
	CertificateAuthorityKey *string `json:"ca_key"`
	// DqliteClusterCertificate is the certificate for connecting to the Dqlite cluster.
	// This is not included in the response when joining worker-only nodes.
	DqliteClusterCertificate string `json:"cluster_cert,omitempty"`
	// DqliteClusterKey is the key for connecting to the Dqlite cluster.
	// This is not included in the response when joining worker-only nodes.
	DqliteClusterKey string `json:"cluster_key,omitempty"`
	// ControlPlaneNodes is a list of known control plane nodes running kube-apiserver.
	// This is only included in the response when joining worker-only nodes.
	ControlPlaneNodes []string `json:"control_plane_nodes"`
	// ClusterCIDR is the cidr that is used by the cluster, defined in kube-proxy args.
	ClusterCIDR string `json:"cluster_cidr,omitempty"`
}

// Join implements "POST v2/join".
// Join returns the join response on success, otherwise an error and the HTTP status code.
func (a *API) Join(ctx context.Context, req JoinRequest) (*JoinResponse, int, error) {
	if !a.Snap.ConsumeClusterToken(req.ClusterToken) {
		return nil, http.StatusInternalServerError, fmt.Errorf("invalid token")
	}
	if !a.Snap.HasDqliteLock() {
		return nil, http.StatusNotImplemented, fmt.Errorf("not possible to join. This is not an HA MicroK8s cluster")
	}

	// Check cluster agent ports.
	clusterAgentBind := snaputil.GetServiceArgument(a.Snap, "cluster-agent", "--bind")
	_, port, _ := net.SplitHostPort(clusterAgentBind)
	if port != req.ClusterAgentPort {
		return nil, http.StatusBadGateway, fmt.Errorf("the port of the cluster agent port has to be set to %s", port)
	}

	// Prevent joins in the same node.
	remoteIP, _, _ := net.SplitHostPort(req.RemoteAddress)
	if hostIP, _, _ := net.SplitHostPort(req.HostPort); remoteIP == hostIP {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("the joining node has the same IP (%s) as the node we contact", hostIP)
	}

	// Check that hostname resolves to the expected IP address
	if util.GetRemoteHost(a.LookupIP, req.RemoteHostName, req.RemoteAddress) != req.RemoteHostName {
		return nil, http.StatusBadRequest, fmt.Errorf("the hostname (%s) of the joining node does not resolve to the IP %q. Refusing join", req.RemoteHostName, remoteIP)
	}

	// Check node is not in cluster already.
	a.dqliteMu.Lock()
	dqliteCluster, err := snaputil.WaitForDqliteCluster(ctx, a.Snap, func(c snaputil.DqliteCluster) (bool, error) {
		return len(c) >= 1, nil
	})
	if err != nil {
		a.dqliteMu.Unlock()
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to retrieve dqlite cluster nodes: %w", err)
	}
	for _, node := range dqliteCluster {
		if strings.HasPrefix(node.Address, remoteIP+":") {
			a.dqliteMu.Unlock()
			return nil, http.StatusGatewayTimeout, fmt.Errorf("the joining node (%s) is already known to dqlite", remoteIP)
		}
	}

	// Update dqlite cluster if needed
	if len(dqliteCluster) == 1 && strings.HasPrefix(dqliteCluster[0].Address, "127.0.0.1:") {
		requestHost, _, _ := net.SplitHostPort(req.HostPort)
		if err := snaputil.UpdateDqliteIP(ctx, a.Snap, requestHost); err != nil {
			a.dqliteMu.Unlock()
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to update dqlite address to %q: %w", requestHost, err)
		}

		// Wait for dqlite cluster to come up with new address
		dqliteCluster, err = snaputil.WaitForDqliteCluster(ctx, a.Snap, func(c snaputil.DqliteCluster) (bool, error) {
			return len(c) >= 1 && !strings.HasPrefix(c[0].Address, "127.0.0.1:"), nil
		})
		if err != nil {
			a.dqliteMu.Unlock()
			return nil, http.StatusInternalServerError, fmt.Errorf("failed waiting for dqlite cluster to come up: %w", err)
		}
	}
	a.dqliteMu.Unlock()

	callbackToken, err := a.Snap.GetOrCreateSelfCallbackToken()
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("could not retrieve self callback token: %w", err)
	}

	ca, err := a.Snap.ReadCA()
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed reading cluster CA: %w", err)
	}
	kubeletArgs, err := a.Snap.ReadServiceArguments("kubelet")
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to read arguments of kubelet service: %w", err)
	}

	a.calicoMu.Lock()
	if err := snaputil.MaybePatchCalicoAutoDetectionMethod(ctx, a.Snap, remoteIP, true); err != nil {
		log.Printf("WARNING: failed to update cni configuration: %q", err)
	}
	a.calicoMu.Unlock()

	if err := a.Snap.CreateNoCertsReissueLock(); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create lock file to disable certificate reissuing: %w", err)
	}
	response := &JoinResponse{
		CertificateAuthority:       ca,
		CallbackToken:              callbackToken,
		APIServerPort:              snaputil.GetServiceArgument(a.Snap, "kube-apiserver", "--secure-port"),
		APIServerAuthorizationMode: snaputil.GetServiceArgument(a.Snap, "kube-apiserver", "--authorization-mode"),
		HostNameOverride:           remoteIP,
		KubeletArgs:                kubeletArgs,
		ClusterCIDR:                snaputil.GetServiceArgument(a.Snap, "kube-proxy", "--cluster-cidr"),
	}

	if req.WorkerOnly {
		if err := a.Snap.AddCertificateRequestToken(fmt.Sprintf("%s-kubelet", req.ClusterToken)); err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed adding certificate request token for kubelet: %w", err)
		}
		if err := a.Snap.AddCertificateRequestToken(fmt.Sprintf("%s-proxy", req.ClusterToken)); err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed adding certificate request token for kube-proxy: %w", err)
		}

		controlPlaneNodes, err := a.ListControlPlaneNodeIPs(ctx, a.Snap)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to retrieve list of control plane nodes: %w", err)
		}
		response.ControlPlaneNodes = controlPlaneNodes
	} else {
		caKey, err := a.Snap.ReadCAKey()
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to retrieve cluster CA key: %w", err)
		}
		response.CertificateAuthorityKey = &caKey
		response.ServiceAccountKey, err = a.Snap.ReadServiceAccountKey()
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to retrieve service account key: %w", err)
		}
		response.AdminToken, err = a.Snap.GetKnownToken("admin")
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to retrieve token for admin user: %w", err)
		}
		response.DqliteClusterCertificate, err = a.Snap.ReadDqliteCert()
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to retrieve dqlite cluster certificate: %w", err)
		}
		response.DqliteClusterKey, err = a.Snap.ReadDqliteKey()
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to retrieve dqlite cluster key: %w", err)
		}
		voters := make([]string, 0, len(dqliteCluster))
		for _, node := range dqliteCluster {
			if node.NodeRole == 0 {
				voters = append(voters, node.Address)
			}
		}
		response.DqliteVoterNodes = voters
	}

	return response, http.StatusOK, nil
}
