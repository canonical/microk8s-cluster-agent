package snap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/canonical/microk8s-cluster-agent/pkg/util"
	"gopkg.in/yaml.v2"
)

// snap implements the Snap interface.
type snap struct {
	snapDir     string
	snapDataDir string
	runCommand  func(context.Context, ...string) error
}

// NewSnap creates a new interface with the MicroK8s snap.
// NewSnap accepts the $SNAP and $SNAP_DATA directories, as well as a function that executes shell commands.
func NewSnap(snapDir, snapDataDir string, runCommand func(context.Context, ...string) error) Snap {
	return &snap{snapDir, snapDataDir, runCommand}
}

func (s *snap) snapPath(parts ...string) string {
	return filepath.Join(append([]string{s.snapDir}, parts...)...)
}

func (s *snap) snapDataPath(parts ...string) string {
	return filepath.Join(append([]string{s.snapDataDir}, parts...)...)
}

func (s *snap) GetGroupName() string {
	if s.isStrict() {
		return "snap_microk8s"
	}
	return "microk8s"
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

func (s *snap) isStrict() bool {
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

func (s *snap) RunUpgrade(ctx context.Context, upgrade string, phase string) error {
	switch phase {
	case "prepare", "commit", "rollback":
	default:
		return fmt.Errorf("unknown upgrade phase %q", phase)
	}
	scriptName := s.snapPath("upgrade-scripts", upgrade, fmt.Sprintf("%s-node.sh", phase))
	if !util.FileExists(scriptName) {
		return fmt.Errorf("could not find script %s", scriptName)
	}
	if err := s.runCommand(ctx, scriptName); err != nil {
		return fmt.Errorf("failed to execute %s: %q", scriptName, err)
	}
	return nil
}

func (s *snap) ReadCA() (string, error) {
	return util.ReadFile(s.snapDataPath("certs", "ca.crt"))
}

func (s *snap) ReadCAKey() (string, error) {
	return util.ReadFile(s.snapDataPath("certs", "ca.key"))
}

func (s *snap) GetCAPath() string {
	return s.snapDataPath("certs", "ca.crt")
}

func (s *snap) GetCAKeyPath() string {
	return s.snapDataPath("certs", "ca.key")
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

func (s *snap) ReadDqliteCert() (string, error) {
	return util.ReadFile(s.snapDataPath("var", "kubernetes", "backend", "cluster.crt"))
}

func (s *snap) ReadDqliteKey() (string, error) {
	return util.ReadFile(s.snapDataPath("var", "kubernetes", "backend", "cluster.key"))
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

func (s *snap) IsValidClusterToken(token string) bool {
	return util.IsValidToken(token, s.snapDataPath("credentials", "cluster-tokens.txt"))
}

func (s *snap) IsValidCertificateRequestToken(token string) bool {
	return util.IsValidToken(token, s.snapDataPath("credentials", "certs-request-tokens.txt"))
}

func (s *snap) IsValidSelfCallbackToken(token string) bool {
	return util.IsValidToken(token, s.snapDataPath("credentials", "callback-token.txt"))
}

func (s *snap) AddCertificateRequestToken(token string) error {
	return util.AppendToken(token, s.snapDataPath("credentials", "certs-request-tokens.txt"), s.GetGroupName())
}

func (s *snap) AddCallbackToken(clusterAgentEndpoint string, token string) error {
	return util.AppendToken(fmt.Sprintf("%s %s", clusterAgentEndpoint, token), s.snapDataPath("credentials", "callback-tokens.txt"), s.GetGroupName())
}

func (s *snap) RemoveClusterToken(token string) error {
	return util.RemoveToken(token, s.snapDataPath("credentials", "cluster-tokens.txt"), s.GetGroupName())
}

func (s *snap) RemoveCertificateRequestToken(token string) error {
	return util.RemoveToken(token, s.snapDataPath("credentials", "certs-request-tokens.txt"), s.GetGroupName())
}

func (s *snap) GetOrCreateSelfCallbackToken() (string, error) {
	callbackTokenFile := s.snapDataPath("credentials", "callback-token.txt")
	c, err := util.ReadFile(callbackTokenFile)
	if err != nil {
		token := util.NewRandomString(util.Alpha, 64)
		if err := os.WriteFile(callbackTokenFile, []byte(fmt.Sprintf("%s\n", token)), 0600); err != nil {
			return "", fmt.Errorf("failed to create callback token file: %w", err)
		}
		return token, nil
	}
	return strings.TrimSpace(c), nil
}

func (s *snap) GetOrCreateKubeletToken(hostname string) (string, error) {
	user := fmt.Sprintf("system:node:%s", hostname)
	existingToken, err := s.GetKnownToken(user)
	if err == nil {
		return existingToken, nil
	}

	token := util.NewRandomString(util.Alpha, 32)
	uid := util.NewRandomString(util.Digits, 8)

	if err := util.AppendToken(fmt.Sprintf("%s,%s,kubelet-%s,\"system:nodes\"", token, user, uid), s.snapDataPath("credentials", "known_tokens.csv"), s.GetGroupName()); err != nil {
		return "", fmt.Errorf("failed to add new kubelet token for %s: %w", user, err)
	}

	return token, nil
}

func (s *snap) GetKnownToken(username string) (string, error) {
	allTokens, err := util.ReadFile(s.snapDataPath("credentials", "known_tokens.csv"))
	if err != nil {
		return "", fmt.Errorf("failed to retrieve known token for user %s: %w", username, err)
	}
	for _, line := range strings.Split(allTokens, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, ",", 3)
		if len(parts) >= 2 && parts[1] == username {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("no known token found for user %s", username)
}

func (s *snap) SignCertificate(ctx context.Context, csrPEM []byte) ([]byte, error) {
	// TODO: consider using crypto/x509 for this instead of relying on openssl commands.
	// NOTE(neoaggelos): x509.CreateCertificate() has some hardcoded fields that are incompatible with MicroK8s.
	signCmd := exec.CommandContext(ctx,
		"openssl", "x509", "-sha256", "-req",
		"-CA", s.snapDataPath("certs", "ca.crt"), "-CAkey", s.snapDataPath("certs", "ca.key"),
		"-CAcreateserial", "-days", "3650",
	)
	stdin, err := signCmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("could not create stdin pipe for sign command: %w", err)
	}
	if _, err := stdin.Write(csrPEM); err != nil {
		return nil, fmt.Errorf("could not write certificate request to openssl command: %w", err)
	}
	stdin.Close()
	certificateBytes, err := signCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("openssl failed: %w", err)
	}

	return certificateBytes, nil
}

var _ Snap = &snap{}
