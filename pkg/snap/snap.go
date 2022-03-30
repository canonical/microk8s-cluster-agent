package snap

import (
	"context"
	"os"
	"path/filepath"

	"github.com/canonical/microk8s-cluster-agent/pkg/util"
	"gopkg.in/yaml.v2"
)

// snap implements the Snap interface.
type snap struct {
	snapDir     string
	snapDataDir string
	runCommand  func(context.Context, ...string) error
}

func NewSnap(snapDir, snapDataDir string, runCommand func(context.Context, ...string) error) Snap {
	return &snap{snapDir, snapDataDir, runCommand}
}

func (s *snap) snapPath(parts ...string) string {
	return filepath.Join(append([]string{s.snapDir}, parts...)...)
}

func (s *snap) snapDataPath(parts ...string) string {
	return filepath.Join(append([]string{s.snapDataDir}, parts...)...)
}

func (s *snap) EnableAddon(ctx context.Context, addon string) error {
	return s.runCommand(ctx, s.snapPath("microk8s-enable.wrapper", addon))
}

func (s *snap) DisableAddon(ctx context.Context, addon string) error {
	return s.runCommand(ctx, s.snapPath("microk8s-disable.wrapper", addon))
}

type snapcraftYml struct {
	Confinement string `yaml:"confinement"`
}

func (s *snap) IsStrict() bool {
	var meta snapcraftYml
	contents, err := util.ReadFile(s.snapPath("meta", "snapcraft.yaml"))
	if err != nil {
		return false
	}
	if err := yaml.Unmarshal([]byte(contents), &meta); err != nil {
		return false
	}
	return meta.Confinement == "strict"
}

// snapctlServiceName infers the name of the snapctl daemon from the service name.
func snapctlServiceName(serviceName string, hasKubelite bool) string {
	switch serviceName {
	case "kube-apiserver", "kube-proxy", "kube-scheduler", "kube-controller-manager":
		// drop kube- prefix
		serviceName = serviceName[5:]
	}
	if hasKubelite {
		switch serviceName {
		case "apiserver", "proxy", "kubelet", "scheduler", "controller-manager":
			serviceName = "kubelite"
		}
	}
	if strings.HasPrefix(serviceName, "microk8s.daemon-") {
		return serviceName
	}
	return fmt.Sprintf("microk8s.daemon-%s", serviceName)
}

func (s *snap) RestartService(ctx context.Context, serviceName string) error {
	return s.runCommand(ctx, "snapctl", "restart", snapctlServiceName(serviceName, s.HasKubeliteLock()))
}

func (s *snap) ReadCA() (string, error) {
	return util.ReadFile(s.snapDataPath("certs", "ca.crt"))
}

func (s *snap) ReadCAKey() (string, error) {
	return util.ReadFile(s.snapDataPath("certs", "ca.key"))
}

func (s *snap) ReadServiceAccountKey() (string, error) {
	return util.ReadFile(s.snapDataPath("certs", "serviceaccount.key"))
}

func (s *snap) GetCNIYamlPath() string {
	return s.snapDataPath("args", "cni-network", "cni.yaml")
}

func (s *snap) ReadCNIYaml() (string, error) {
	return util.ReadFile(s.snapDataPath("args", "cni-network", "cni.yaml"))
}

func (s *snap) WriteCNIYaml(cniManifest []byte) error {
	return os.WriteFile(s.snapDataPath("args", "cni-network", "cni.yaml"), []byte(cniManifest), 0660)
}

func (s *snap) ApplyCNI(ctx context.Context) error {
	return s.runCommand(ctx, s.snapPath("microk8s-kubectl.wrapper"), "apply", "-f", s.GetCNIYamlPath())
}

func (s *snap) ReadDqliteInfoYaml() (string, error) {
	return util.ReadFile(s.snapDataPath("var", "kubernetes", "backend", "info.yaml"))
}

func (s *snap) ReadDqliteClusterYaml() (string, error) {
	return util.ReadFile(s.snapDataPath("var", "kubernetes", "backend", "cluster.yaml"))
}

func (s *snap) WriteDqliteUpdateYaml(updateYaml []byte) error {
	return os.WriteFile(s.snapDataPath("var", "kubernetes", "backend", "update.yaml"), updateYaml, 0660)
}

func (s *snap) GetKubeconfigFile() string {
	return s.snapDataPath("credentials", "client.config")
}

func (s *snap) HasKubeliteLock() bool {
	return util.FileExists(s.snapDataPath("var", "lock", "lite.lock"))
}

func (s *snap) HasDqliteLock() bool {
	return util.FileExists(s.snapDataPath("var", "lock", "ha-cluster"))
}

func (s *snap) HasNoCertsReissueLock() bool {
	return util.FileExists(s.snapDataPath("var", "lock", "no-cert-reissue"))
}

func (s *snap) CreateNoCertsReissueLock() error {
	_, err := os.OpenFile(s.snapDataPath("var", "lock", "no-cert-reissue"), os.O_CREATE, 0600)
	return err
}

func (s *snap) ReadServiceArguments(serviceName string) (string, error) {
	return util.ReadFile(s.snapDataPath("args", serviceName))
}

func (s *snap) WriteServiceArguments(serviceName string, arguments []byte) error {
	return os.WriteFile(s.snapDataPath("args", serviceName), arguments, 0660)
}

var _ Snap = &snap{}
