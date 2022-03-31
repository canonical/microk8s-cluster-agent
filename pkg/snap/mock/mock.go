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
	CallbackTokens           []string // "{clusterAgentEndpoint} {token}"
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

func (s *Snap) IsStrict() bool {
	return s.Strict
}

func (s *Snap) EnableAddon(_ context.Context, addon string) error {
	s.EnableAddonCalledWith = append(s.EnableAddonCalledWith, addon)
	return nil
}

func (s *Snap) DisableAddon(_ context.Context, addon string) error {
	s.DisableAddonCalledWith = append(s.DisableAddonCalledWith, addon)
	return nil
}

func (s *Snap) RestartService(_ context.Context, service string) error {
	s.RestartServiceCalledWith = append(s.RestartServiceCalledWith, service)
	return nil
}

func (s *Snap) ReadCA() (string, error) {
	return s.CA, nil
}

func (s *Snap) ReadCAKey() (string, error) {
	return s.CAKey, nil
}

func (s *Snap) ReadServiceAccountKey() (string, error) {
	return s.ServiceAccountKey, nil
}

func (s *Snap) ReadCNIYaml() (string, error) {
	return s.CNIYaml, nil
}

func (s *Snap) WriteCNIYaml(b []byte) error {
	s.CNIYaml = string(b)
	return nil
}

func (s *Snap) ApplyCNI(_ context.Context) error {
	s.ApplyCNICalled = append(s.ApplyCNICalled, struct{}{})
	return nil
}

func (s *Snap) ReadDqliteCert() (string, error) {
	return s.DqliteCert, nil
}

func (s *Snap) ReadDqliteKey() (string, error) {
	return s.DqliteKey, nil
}

func (s *Snap) ReadDqliteInfoYaml() (string, error) {
	return s.DqliteInfoYaml, nil
}

func (s *Snap) ReadDqliteClusterYaml() (string, error) {
	return s.DqliteClusterYaml, nil
}

func (s *Snap) WriteDqliteUpdateYaml(b []byte) error {
	s.WriteDqliteUpdateYamlCalledWith = append(s.WriteDqliteUpdateYamlCalledWith, string(b))
	return nil
}

func (s *Snap) GetKubeconfigFile() string {
	return s.KubeconfigFile
}

func (s *Snap) HasKubeliteLock() bool {
	return s.KubeliteLock
}

func (s *Snap) HasDqliteLock() bool {
	return s.DqliteLock
}

func (s *Snap) HasNoCertsReissueLock() bool {
	return s.NoCertsReissueLock
}

func (s *Snap) CreateNoCertsReissueLock() error {
	s.NoCertsReissueLock = true
	s.CreateNoCertsReissueLockCalledWith = append(s.CreateNoCertsReissueLockCalledWith, struct{}{})
	return nil
}

func (s *Snap) ReadServiceArguments(service string) (string, error) {
	if s.ServiceArguments == nil {
		s.ServiceArguments = make(map[string]string)
	}
	return s.ServiceArguments[service], nil
}

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

func (s *Snap) IsValidClusterToken(token string) bool {
	return contains(s.ClusterTokens, token)
}

func (s *Snap) IsValidCertificateRequestToken(token string) bool {
	return contains(s.CertificateRequestTokens, token)
}

func (s *Snap) IsValidCallbackToken(clusterAgentEndpoint string, token string) bool {
	return contains(s.CallbackTokens, fmt.Sprintf("%s %s", clusterAgentEndpoint, token))
}

func (s *Snap) IsValidSelfCallbackToken(token string) bool {
	return contains(s.SelfCallbackTokens, token)
}

func (s *Snap) AddCertificateRequestToken(token string) error {
	if s.AddCertificateRequestTokenCalledWith == nil {
		s.AddCertificateRequestTokenCalledWith = make([]string, 0, 1)
	}

	s.AddCertificateRequestTokenCalledWith = append(s.AddCertificateRequestTokenCalledWith, token)
	return nil
}

func (s *Snap) AddCallbackToken(clusterAgentEndpoint string, token string) error {
	if s.AddCallbackTokenCalledWith == nil {
		s.AddCallbackTokenCalledWith = make([]string, 0, 1)
	}

	s.AddCallbackTokenCalledWith = append(s.AddCallbackTokenCalledWith, fmt.Sprintf("%s %s", clusterAgentEndpoint, token))
	return nil
}

func (s *Snap) RemoveClusterToken(token string) error {
	if s.RemoveClusterTokenCalledWith == nil {
		s.RemoveClusterTokenCalledWith = make([]string, 0, 1)
	}

	s.RemoveClusterTokenCalledWith = append(s.RemoveClusterTokenCalledWith, token)
	return nil
}

func (s *Snap) RemoveCertificateRequestToken(token string) error {
	if s.RemoveCertificateRequestTokenCalledWith == nil {
		s.RemoveCertificateRequestTokenCalledWith = make([]string, 0, 1)
	}

	s.RemoveCertificateRequestTokenCalledWith = append(s.RemoveCertificateRequestTokenCalledWith, token)
	return nil
}

func (s *Snap) GetOrCreateSelfCallbackToken() (string, error) {
	if s.SelfCallbackToken == "" {
		s.SelfCallbackToken = "callback-token"
	}
	return s.SelfCallbackToken, nil
}

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

func (s *Snap) GetKnownToken(username string) (string, error) {
	if t, ok := s.KnownTokens[username]; ok {
		return t, nil
	}
	return "", fmt.Errorf("no known token for user %s", username)
}

func (s *Snap) RunUpgrade(ctx context.Context, upgrade string, phase string) error {
	if s.RunUpgradeCalledWith == nil {
		s.RunUpgradeCalledWith = make([]string, 0, 1)
	}
	s.RunUpgradeCalledWith = append(s.RunUpgradeCalledWith, fmt.Sprintf("%s %s", upgrade, phase))
	return nil
}

func (s *Snap) SignCertificate(ctx context.Context, csrPEM []byte) ([]byte, error) {
	if s.SignCertificateCalledWith == nil {
		s.SignCertificateCalledWith = make([]string, 0, 1)
	}
	s.SignCertificateCalledWith = append(s.SignCertificateCalledWith, string(csrPEM))
	return []byte(s.SignedCertificate), nil
}

var _ snap.Snap = &Snap{}
