package snap

import "context"

// Snap is how the cluster agent interacts with the snap.
type Snap interface {
	IsStrict() bool

	EnableAddon(context.Context, string) error
	DisableAddon(context.Context, string) error
	RestartService(context.Context, string) error

	ReadCA() (string, error)
	ReadCAKey() (string, error)
	ReadServiceAccountKey() (string, error)

	ReadCNIYaml() (string, error)
	WriteCNIYaml([]byte) error
	ApplyCNI(context.Context) error

	ReadDqliteInfoYaml() (string, error)
	ReadDqliteClusterYaml() (string, error)
	WriteDqliteUpdateYaml([]byte) error

	GetKubeconfigFile() string

	HasKubeliteLock() bool
	HasDqliteLock() bool
	HasNoCertsReissueLock() bool
	CreateNoCertsReissueLock() error

	ReadServiceArguments(string) (string, error)
	WriteServiceArguments(string, []byte) error
}
