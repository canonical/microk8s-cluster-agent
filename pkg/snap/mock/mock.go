package mock

import (
	"context"
	"fmt"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
)

// Snap is a generic mock for the snap.Snap interface.
type Snap struct {
	Strict bool

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

	AddCertificateRequestTokenCalledWith []string
	AddCallbackTokenCalledWith           []string // "{clusterAgentEndpoint} {token}"

	RemoveClusterTokenCalledWith            []string
	RemoveCertificateRequestTokenCalledWith []string

	SelfCallbackToken string
	KubeletTokens     map[string]string // map hostname to token
	KnownTokens       map[string]string // map username to token

	SignCertificateCalledWith []string // string(csrPEM)
	SignedCertificate         string
}

// IsStrict is a mock implementation for the snap.Snap interface.
func (s *Snap) IsStrict() bool {
	return s.Strict
}

// EnableAddon is a mock implementation for the snap.Snap interface.
func (s *Snap) EnableAddon(_ context.Context, addon string) error {
	s.EnableAddonCalledWith = append(s.EnableAddonCalledWith, addon)
	return nil
}

// DisableAddon is a mock implementation for the snap.Snap interface.
func (s *Snap) DisableAddon(_ context.Context, addon string) error {
	s.DisableAddonCalledWith = append(s.DisableAddonCalledWith, addon)
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

// IsValidClusterToken is a mock implementation for the snap.Snap interface.
func (s *Snap) IsValidClusterToken(token string) bool {
	return contains(s.ClusterTokens, token)
}

// IsValidCertificateRequestToken is a mock implementation for the snap.Snap interface.
func (s *Snap) IsValidCertificateRequestToken(token string) bool {
	return contains(s.CertificateRequestTokens, token)
}

// IsValidSelfCallbackToken is a mock implementation for the snap.Snap interface.
func (s *Snap) IsValidSelfCallbackToken(token string) bool {
	return contains(s.SelfCallbackTokens, token)
}

// AddCertificateRequestToken is a mock implementation for the snap.Snap interface.
func (s *Snap) AddCertificateRequestToken(token string) error {
	if s.AddCertificateRequestTokenCalledWith == nil {
		s.AddCertificateRequestTokenCalledWith = make([]string, 0, 1)
	}

	s.AddCertificateRequestTokenCalledWith = append(s.AddCertificateRequestTokenCalledWith, token)
	return nil
}

// AddCallbackToken is a mock implementation for the snap.Snap interface.
func (s *Snap) AddCallbackToken(clusterAgentEndpoint string, token string) error {
	if s.AddCallbackTokenCalledWith == nil {
		s.AddCallbackTokenCalledWith = make([]string, 0, 1)
	}

	s.AddCallbackTokenCalledWith = append(s.AddCallbackTokenCalledWith, fmt.Sprintf("%s %s", clusterAgentEndpoint, token))
	return nil
}

// RemoveClusterToken is a mock implementation for the snap.Snap interface.
func (s *Snap) RemoveClusterToken(token string) error {
	if s.RemoveClusterTokenCalledWith == nil {
		s.RemoveClusterTokenCalledWith = make([]string, 0, 1)
	}

	s.RemoveClusterTokenCalledWith = append(s.RemoveClusterTokenCalledWith, token)
	return nil
}

// RemoveCertificateRequestToken is a mock implementation for the snap.Snap interface.
func (s *Snap) RemoveCertificateRequestToken(token string) error {
	if s.RemoveCertificateRequestTokenCalledWith == nil {
		s.RemoveCertificateRequestTokenCalledWith = make([]string, 0, 1)
	}

	s.RemoveCertificateRequestTokenCalledWith = append(s.RemoveCertificateRequestTokenCalledWith, token)
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
		fmt.Println("AA")
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

// RunUpgrade is a mock implementation for the snap.Snap interface.
func (s *Snap) RunUpgrade(ctx context.Context, upgrade string, phase string) error {
	if s.RunUpgradeCalledWith == nil {
		s.RunUpgradeCalledWith = make([]string, 0, 1)
	}
	s.RunUpgradeCalledWith = append(s.RunUpgradeCalledWith, fmt.Sprintf("%s %s", upgrade, phase))
	return nil
}

// SignCertificate is a mock implementation for the snap.Snap interface.
func (s *Snap) SignCertificate(ctx context.Context, csrPEM []byte) ([]byte, error) {
	if s.SignCertificateCalledWith == nil {
		s.SignCertificateCalledWith = make([]string, 0, 1)
	}
	s.SignCertificateCalledWith = append(s.SignCertificateCalledWith, string(csrPEM))
	return []byte(s.SignedCertificate), nil
}

var _ snap.Snap = &Snap{}
