package snap

import (
	"context"
	"io"
)

// Snap is how the cluster agent interacts with the snap.
type Snap interface {
	// GetSnapPath returns the path to a file or directory in the snap directory.
	GetSnapPath(parts ...string) string
	// GetSnapDataPath returns the path to a file or directory in the snap's data directory.
	GetSnapDataPath(parts ...string) string
	// GetSnapCommonPath returns the path to a file or directory in the snap's common directory.
	GetSnapCommonPath(parts ...string) string
	// GetCAPIPath returns the path to a file or directory in the CAPI directory.
	GetCAPIPath(parts ...string) string

	// RunCommand runs a shell command.
	RunCommand(ctx context.Context, commands ...string) error

	// GetGroupName is the group microk8s is using.
	// The group name is "microk8s" for classic snaps and "snap_microk8s" for strict snaps.
	GetGroupName() string

	// EnableAddon enables a MicroK8s addon.
	EnableAddon(ctx context.Context, addon string, args ...string) error
	// DisableAddon disables a MicroK8s addon.
	DisableAddon(ctx context.Context, addon string, args ...string) error
	// RestartService restarts a MicroK8s service.
	RestartService(ctx context.Context, serviceName string) error
	// RunUpgrade runs a single phase for an upgrade script. See the upgrade-scripts folder.
	RunUpgrade(ctx context.Context, upgrade string, phase string) error

	// ReadCA returns the CA certificate in PEM format.
	ReadCA() (string, error)
	// ReadCAKey returns the CA private key in PEM format.
	ReadCAKey() (string, error)
	// ReadServiceAccountKey returns the Service Account key in PEM format.
	ReadServiceAccountKey() (string, error)

	// ReadCNIYaml returns the CNI manifest yaml from the snap.
	ReadCNIYaml() (string, error)
	// WriteCNIYaml updates the CNI manifest yaml.
	WriteCNIYaml([]byte) error
	// ApplyCNI applies the current CNI manifest in the MicroK8s cluster.
	ApplyCNI(ctx context.Context) error

	// ReadDqliteCert returns the dqlite certificate in PEM format.
	ReadDqliteCert() (string, error)
	// ReadDqliteKey returns the dqlite private key in PEM format.
	ReadDqliteKey() (string, error)
	// ReadDqliteInfoYaml returns the contents of dqlite's info.yaml file.
	ReadDqliteInfoYaml() (string, error)
	// ReadDqliteClusterYaml returns the contents of dqlite's cluster.yaml file.
	ReadDqliteClusterYaml() (string, error)
	// WriteDqliteUpdateYaml writes a dqlite update.yaml file, used to reconfigure the IP address of the local dqlite node.
	WriteDqliteUpdateYaml(b []byte) error

	// GetKubeconfigFile returns the path to the client kubeconfig file.
	GetKubeconfigFile() string

	// HasKubeliteLock returns true if this MicroK8s instance is running Kubelite.
	HasKubeliteLock() bool
	// HasDqliteLock returns true if this MicroK8s instance is running dqlite.
	HasDqliteLock() bool
	// HasNoCertsReissueLock returns true if the lock file to prevent reissue of the CA certificates is present in this MicroK8s instance.
	HasNoCertsReissueLock() bool
	// CreateNoCertsReissueLock creates the lock file to prevent reissue of CA certificates in this MicroK8s instance.
	CreateNoCertsReissueLock() error

	// ReadServiceArguments reads the arguments file for a particular service.
	ReadServiceArguments(serviceName string) (string, error)
	// WriteServiceArguments updates the arguments file a particular service.
	WriteServiceArguments(serviceName string, b []byte) error

	// ConsumeClusterToken returns true if token is a valid token for authenticating join requests.
	// Tokens with a TTL may be consumed multiple times until they expire. One-time tokens may only be consumed once.
	ConsumeClusterToken(token string) bool
	// ConsumeCertificateRequestToken returns true if token is a valid token for authenticating certificate signing requests.
	// Certificate request tokens may only be consumed once.
	ConsumeCertificateRequestToken(token string) bool
	// ConsumeSelfCallbackToken returns true if token is a valid token for authenticating configure and upgrade requests.
	// Self callback tokens may be consumed multiple times.
	ConsumeSelfCallbackToken(token string) bool

	// AddPersistentClusterToken adds a new persistent token that can be used to authenticate join requests.
	AddPersistentClusterToken(token string) error
	// AddCertificateRequestToken adds a new token that can be used to authenticate certificate signing requests.
	AddCertificateRequestToken(token string) error
	// AddCallbackToken adds a new token that can be used to authenticate requests to a remote cluster agent endpoint.
	AddCallbackToken(clusterAgentEndpoint, token string) error

	// GetOrCreateSelfCallbackToken creates and returns the callback token that can be used for configure and upgrade requests to this cluster agent.
	// Subsequent calls should return the same token.
	GetOrCreateSelfCallbackToken() (string, error)
	// GetOrCreateKubeletToken creates and returns a token used to authenticate a kubelet to the API server.
	// Subsequent calls should return the same token.
	GetOrCreateKubeletToken(hostname string) (string, error)
	// GetKnownToken returns the token for a known user from the known_users.csv file.
	GetKnownToken(username string) (string, error)

	// IsCAPIAuthTokenValid returns true if token is a valid CAPI auth token.
	IsCAPIAuthTokenValid(token string) (bool, error)

	// SignCertificate signs the certificate signing request, and returns the certificate in PEM format.
	SignCertificate(ctx context.Context, csrPEM []byte) ([]byte, error)

	// ImportImage imports an OCI image from raw bytes.
	ImportImage(ctx context.Context, reader io.Reader) error

	// WriteCSRConfig updates the csr.conf.template file on the local node.
	WriteCSRConfig(csrConf []byte) error

	// UpdateContainerdRegistryConfigs writes hosts.toml registry configurations for containerd.
	// Accepts a map where key is the registry (e.g. "docker.io") and the value is the contents of the hosts.toml file.
	UpdateContainerdRegistryConfigs(configs map[string][]byte) error

	// AddAddonsRepository configures an addons repository on the local node, similar to running the 'microk8s addons repo add' command.
	AddAddonsRepository(ctx context.Context, name, url, reference string, force bool) error

	// JoinCluster joins the local node to an existing MicroK8s cluster as a control-plane or worker node.
	JoinCluster(ctx context.Context, url string, worker bool) error

	// ReadEtcdCertificates returns the certificates (CA, certificate, private key) used by kube-apiserver to connect to etcd.
	// Empty values are returned for certificates that are not used.
	// An error is returned if any of the certificates is in use cannot be read.
	// The values returned by this function match the contents of '--etcd-cafile', '--etcd-certfile', '--etcd-keyfile' kube-apiserver arguments respectively.
	ReadEtcdCertificates() (ca string, cert string, key string, err error)
}
