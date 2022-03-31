package snap

import (
	"context"
)

// Snap is how the cluster agent interacts with the snap.
type Snap interface {
	IsStrict() bool

	EnableAddon(ctx context.Context, addon string) error
	DisableAddon(ctx context.Context, addon string) error
	RestartService(ctx context.Context, serviceName string) error
	RunUpgrade(ctx context.Context, upgrade string, phase string) error

	ReadCA() (string, error)
	ReadCAKey() (string, error)
	ReadServiceAccountKey() (string, error)

	ReadCNIYaml() (string, error)
	WriteCNIYaml([]byte) error
	ApplyCNI(ctx context.Context) error

	ReadDqliteCert() (string, error)
	ReadDqliteKey() (string, error)
	ReadDqliteInfoYaml() (string, error)
	ReadDqliteClusterYaml() (string, error)
	WriteDqliteUpdateYaml(b []byte) error

	GetKubeconfigFile() string

	HasKubeliteLock() bool
	HasDqliteLock() bool
	HasNoCertsReissueLock() bool
	CreateNoCertsReissueLock() error

	ReadServiceArguments(serviceName string) (string, error)
	WriteServiceArguments(serviceName string, b []byte) error

	IsValidClusterToken(token string) bool
	IsValidCertificateRequestToken(token string) bool
	IsValidCallbackToken(clusterAgentEndpoint, token string) bool
	IsValidSelfCallbackToken(token string) bool

	AddCertificateRequestToken(token string) error
	AddCallbackToken(clusterAgentEndpoint, token string) error

	RemoveClusterToken(token string) error
	RemoveCertificateRequestToken(token string) error

	GetOrCreateSelfCallbackToken() (string, error)
	GetOrCreateKubeletToken(hostname string) (string, error)
	GetKnownToken(username string) (string, error)

	SignCertificate(ctx context.Context, csrPEM []byte) ([]byte, error)
}
