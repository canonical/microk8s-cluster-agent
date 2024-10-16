package mock

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
)

// AddonRepository is a mock for the current state of an addon repository.
type AddonRepository struct {
	URL       string
	Reference string
	Force     bool
}

// JoinClusterCall is a mock for the join cluster call.
type JoinClusterCall struct {
	URL    string
	Worker bool
}

// RunCommandCall contains the arguments passed to a specific call of the RunCommand method.
type RunCommandCall struct {
	Commands []string
}

// Snap is a generic mock for the snap.Snap interface.
type Snap struct {
	SnapDir       string
	SnapDataDir   string
	SnapCommonDir string
	CAPIDir       string

	RunCommandCalledWith []RunCommandCall
	RunCommandErr        error

	GroupName string

	EnableAddonCalledWith    []string
	DisableAddonCalledWith   []string
	RestartServiceCalledWith []string
	RunUpgradeCalledWith     []string // "{upgrade} {phase}"

	CA                string
	CAKey             string
	ServiceAccountKey string

	CNIYaml                string
	WriteCNIYamlCalledWith [][]byte
	ApplyCNICalled         []struct{}

	DqliteCert        string
	DqliteKey         string
	DqliteClusterYaml string
	DqliteInfoYaml    string

	WriteDqliteUpdateYamlCalledWith []string

	KubeconfigFile string

	KubeliteLock                       bool
	DqliteLock                         bool
	NoCertsReissueLock                 bool
	CreateNoCertsReissueLockCalledWith []struct{}

	ServiceArguments            map[string]string
	WriteServiceArgumentsCalled bool

	ClusterTokens            []string
	CertificateRequestTokens []string
	SelfCallbackTokens       []string

	AddPersistentClusterTokenCalledWith  []string
	AddCertificateRequestTokenCalledWith []string
	AddCallbackTokenCalledWith           []string // "{clusterAgentEndpoint} {token}"

	ConsumeClusterTokenCalledWith            []string
	ConsumeCertificateRequestTokenCalledWith []string

	SelfCallbackToken string
	KubeletTokens     map[string]string // map hostname to token
	KnownTokens       map[string]string // map username to token

	CAPIAuthTokenValid bool
	CAPIAuthTokenError error

	SignCertificateCalledWith []string // string(csrPEM)
	SignedCertificate         string

	ImportImageCalledWith []string // string(io.ReadAll(reader))

	CSRConfig string

	ContainerdRegistryConfigs map[string]string // map registry name to hosts.toml contents

	AddonRepositories map[string]AddonRepository

	JoinClusterCalledWith []JoinClusterCall

	EtcdCA, EtcdCert, EtcdKey string
}

// GetSnapPath is a mock implementation for the snap.Snap interface.
func (s *Snap) GetSnapPath(parts ...string) string {
	return filepath.Join(append([]string{s.SnapDir}, parts...)...)
}

// GetSnapDataPath is a mock implementation for the snap.Snap interface.
func (s *Snap) GetSnapDataPath(parts ...string) string {
	return filepath.Join(append([]string{s.SnapDataDir}, parts...)...)
}

// GetSnapCommonPath is a mock implementation for the snap.Snap interface.
func (s *Snap) GetSnapCommonPath(parts ...string) string {
	return filepath.Join(append([]string{s.SnapCommonDir}, parts...)...)
}

// GetCAPIPath is a mock implementation for the snap.Snap interface.
func (s *Snap) GetCAPIPath(parts ...string) string {
	return filepath.Join(append([]string{s.CAPIDir}, parts...)...)
}

// RunCommand is a mock implementation for the snap.Snap interface.
func (s *Snap) RunCommand(_ context.Context, commands ...string) error {
	s.RunCommandCalledWith = append(s.RunCommandCalledWith, RunCommandCall{Commands: commands})
	return s.RunCommandErr
}

// GetGroupName is a mock implementation for the snap.Snap interface.
func (s *Snap) GetGroupName() string {
	return s.GroupName
}

// EnableAddon is a mock implementation for the snap.Snap interface.
func (s *Snap) EnableAddon(_ context.Context, addon string, args ...string) error {
	s.EnableAddonCalledWith = append(s.EnableAddonCalledWith, strings.TrimSpace(fmt.Sprintf("%s %s", addon, strings.Join(args, " "))))
	return nil
}

// DisableAddon is a mock implementation for the snap.Snap interface.
func (s *Snap) DisableAddon(_ context.Context, addon string, args ...string) error {
	s.DisableAddonCalledWith = append(s.DisableAddonCalledWith, strings.TrimSpace(fmt.Sprintf("%s %s", addon, strings.Join(args, " "))))
	return nil
}

// RestartService is a mock implementation for the snap.Snap interface.
func (s *Snap) RestartService(_ context.Context, service string) error {
	s.RestartServiceCalledWith = append(s.RestartServiceCalledWith, service)
	return nil
}

// ReadCA is a mock implementation for the snap.Snap interface.
func (s *Snap) ReadCA() (string, error) {
	return s.CA, nil
}

// ReadCAKey is a mock implementation for the snap.Snap interface.
func (s *Snap) ReadCAKey() (string, error) {
	return s.CAKey, nil
}

// ReadServiceAccountKey is a mock implementation for the snap.Snap interface.
func (s *Snap) ReadServiceAccountKey() (string, error) {
	return s.ServiceAccountKey, nil
}

// ReadCNIYaml is a mock implementation for the snap.Snap interface.
func (s *Snap) ReadCNIYaml() (string, error) {
	return s.CNIYaml, nil
}

// WriteCNIYaml is a mock implementation for the snap.Snap interface.
func (s *Snap) WriteCNIYaml(b []byte) error {
	s.CNIYaml = string(b)
	return nil
}

// ApplyCNI is a mock implementation for the snap.Snap interface.
func (s *Snap) ApplyCNI(_ context.Context) error {
	s.ApplyCNICalled = append(s.ApplyCNICalled, struct{}{})
	return nil
}

// ReadDqliteCert is a mock implementation for the snap.Snap interface.
func (s *Snap) ReadDqliteCert() (string, error) {
	return s.DqliteCert, nil
}

// ReadDqliteKey is a mock implementation for the snap.Snap interface.
func (s *Snap) ReadDqliteKey() (string, error) {
	return s.DqliteKey, nil
}

// ReadDqliteInfoYaml is a mock implementation for the snap.Snap interface.
func (s *Snap) ReadDqliteInfoYaml() (string, error) {
	return s.DqliteInfoYaml, nil
}

// ReadDqliteClusterYaml is a mock implementation for the snap.Snap interface.
func (s *Snap) ReadDqliteClusterYaml() (string, error) {
	return s.DqliteClusterYaml, nil
}

// WriteDqliteUpdateYaml is a mock implementation for the snap.Snap interface.
func (s *Snap) WriteDqliteUpdateYaml(b []byte) error {
	s.WriteDqliteUpdateYamlCalledWith = append(s.WriteDqliteUpdateYamlCalledWith, string(b))
	return nil
}

// GetKubeconfigFile is a mock implementation for the snap.Snap interface.
func (s *Snap) GetKubeconfigFile() string {
	return s.KubeconfigFile
}

// HasKubeliteLock is a mock implementation for the snap.Snap interface.
func (s *Snap) HasKubeliteLock() bool {
	return s.KubeliteLock
}

// HasDqliteLock is a mock implementation for the snap.Snap interface.
func (s *Snap) HasDqliteLock() bool {
	return s.DqliteLock
}

// HasNoCertsReissueLock is a mock implementation for the snap.Snap interface.
func (s *Snap) HasNoCertsReissueLock() bool {
	return s.NoCertsReissueLock
}

// CreateNoCertsReissueLock is a mock implementation for the snap.Snap interface.
func (s *Snap) CreateNoCertsReissueLock() error {
	s.NoCertsReissueLock = true
	s.CreateNoCertsReissueLockCalledWith = append(s.CreateNoCertsReissueLockCalledWith, struct{}{})
	return nil
}

// ReadServiceArguments is a mock implementation for the snap.Snap interface.
func (s *Snap) ReadServiceArguments(service string) (string, error) {
	if s.ServiceArguments == nil {
		s.ServiceArguments = make(map[string]string)
	}
	return s.ServiceArguments[service], nil
}

// WriteServiceArguments is a mock implementation for the snap.Snap interface.
func (s *Snap) WriteServiceArguments(service string, b []byte) error {
	if s.ServiceArguments == nil {
		s.ServiceArguments = make(map[string]string)
	}
	s.ServiceArguments[service] = string(b)
	s.WriteServiceArgumentsCalled = true
	return nil
}

func contains(list []string, item string) bool {
	for _, i := range list {
		if item == i {
			return true
		}
	}
	return false
}

// ConsumeClusterToken is a mock implementation for the snap.Snap interface.
func (s *Snap) ConsumeClusterToken(token string) bool {
	s.ConsumeClusterTokenCalledWith = append(s.ConsumeClusterTokenCalledWith, token)
	return contains(s.ClusterTokens, token)
}

// ConsumeCertificateRequestToken is a mock implementation for the snap.Snap interface.
func (s *Snap) ConsumeCertificateRequestToken(token string) bool {
	s.ConsumeCertificateRequestTokenCalledWith = append(s.ConsumeCertificateRequestTokenCalledWith, token)
	return contains(s.CertificateRequestTokens, token)
}

// ConsumeSelfCallbackToken is a mock implementation for the snap.Snap interface.
func (s *Snap) ConsumeSelfCallbackToken(token string) bool {
	return contains(s.SelfCallbackTokens, token)
}

// AddPersistentClusterToken is a mock implementation for the snap.Snap interface.
func (s *Snap) AddPersistentClusterToken(token string) error {
	s.AddPersistentClusterTokenCalledWith = append(s.AddPersistentClusterTokenCalledWith, token)
	return nil
}

// AddCertificateRequestToken is a mock implementation for the snap.Snap interface.
func (s *Snap) AddCertificateRequestToken(token string) error {
	s.AddCertificateRequestTokenCalledWith = append(s.AddCertificateRequestTokenCalledWith, token)
	return nil
}

// AddCallbackToken is a mock implementation for the snap.Snap interface.
func (s *Snap) AddCallbackToken(clusterAgentEndpoint string, token string) error {
	s.AddCallbackTokenCalledWith = append(s.AddCallbackTokenCalledWith, fmt.Sprintf("%s %s", clusterAgentEndpoint, token))
	return nil
}

// GetOrCreateSelfCallbackToken is a mock implementation for the snap.Snap interface.
func (s *Snap) GetOrCreateSelfCallbackToken() (string, error) {
	if s.SelfCallbackToken == "" {
		s.SelfCallbackToken = "callback-token"
	}
	return s.SelfCallbackToken, nil
}

// GetOrCreateKubeletToken is a mock implementation for the snap.Snap interface.
func (s *Snap) GetOrCreateKubeletToken(hostname string) (string, error) {
	if s.KubeletTokens == nil {
		s.KubeletTokens = make(map[string]string, 1)
	}
	if t, ok := s.KubeletTokens[hostname]; ok {
		return t, nil
	}
	s.KubeletTokens[hostname] = util.NewRandomString(util.Alpha, 32)
	return s.KubeletTokens[hostname], nil
}

// GetKnownToken is a mock implementation for the snap.Snap interface.
func (s *Snap) GetKnownToken(username string) (string, error) {
	if t, ok := s.KnownTokens[username]; ok {
		return t, nil
	}
	return "", fmt.Errorf("no known token for user %s", username)
}

// IsCAPIAuthTokenValid is a mock implementation for the snap.Snap interface.
func (s *Snap) IsCAPIAuthTokenValid(token string) (bool, error) {
	return s.CAPIAuthTokenValid, s.CAPIAuthTokenError
}

// RunUpgrade is a mock implementation for the snap.Snap interface.
func (s *Snap) RunUpgrade(ctx context.Context, upgrade string, phase string) error {
	s.RunUpgradeCalledWith = append(s.RunUpgradeCalledWith, fmt.Sprintf("%s %s", upgrade, phase))
	return nil
}

// SignCertificate is a mock implementation for the snap.Snap interface.
func (s *Snap) SignCertificate(ctx context.Context, csrPEM []byte) ([]byte, error) {
	s.SignCertificateCalledWith = append(s.SignCertificateCalledWith, string(csrPEM))
	return []byte(s.SignedCertificate), nil
}

// ImportImage is a mock implementation for the snap.Snap interface.
func (s *Snap) ImportImage(ctx context.Context, reader io.Reader) error {
	b, _ := io.ReadAll(reader)
	s.ImportImageCalledWith = append(s.ImportImageCalledWith, string(b))
	return nil
}

// WriteCSRConfig is a mock implementation for the snap.Snap interface.
func (s *Snap) WriteCSRConfig(b []byte) error {
	s.CSRConfig = string(b)
	return nil
}

// UpdateContainerdRegistryConfigs is a mock implementation for the snap.Snap interface.
func (s *Snap) UpdateContainerdRegistryConfigs(cfgs map[string][]byte) error {
	if s.ContainerdRegistryConfigs == nil {
		s.ContainerdRegistryConfigs = make(map[string]string)
	}
	for k, v := range cfgs {
		s.ContainerdRegistryConfigs[k] = string(v)
	}
	return nil
}

// AddAddonsRepository is a mock implementation for the snap.Snap interface.
func (s *Snap) AddAddonsRepository(ctx context.Context, name, url, reference string, force bool) error {
	if s.AddonRepositories == nil {
		s.AddonRepositories = make(map[string]AddonRepository)
	}

	s.AddonRepositories[name] = AddonRepository{
		URL:       url,
		Reference: reference,
		Force:     force,
	}
	return nil
}

// JoinCluster is a mock implementation for the snap.Snap interface.
func (s *Snap) JoinCluster(ctx context.Context, url string, worker bool) error {
	s.JoinClusterCalledWith = append(s.JoinClusterCalledWith, JoinClusterCall{url, worker})
	return nil
}

// ReadEtcdCertificates is a mock implementation for the snap.Snap interface.
func (s *Snap) ReadEtcdCertificates() (string, string, string, error) {
	return s.EtcdCA, s.EtcdCert, s.EtcdKey, nil
}

var _ snap.Snap = &Snap{}
