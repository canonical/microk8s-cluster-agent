package mock

import "context"

// Snap is a generic mock for the snap.Snap interface.
type Snap struct {
	Strict bool

	EnableAddonCalledWith    []string
	DisableAddonCalledWith   []string
	RestartServiceCalledWith []string

	CA                string
	CAKey             string
	ServiceAccountKey string

	CNIYaml                string
	WriteCNIYamlCalledWith [][]byte
	ApplyCNICalled         []struct{}

	DqliteClusterYaml               string
	DqliteInfoYaml                  string
	WriteDqliteUpdateYamlCalledWith [][]byte

	KubeconfigFile string

	KubeliteLock                       bool
	DqliteLock                         bool
	NoCertsReissueLock                 bool
	CreateNoCertsReissueLockCalledWith []struct{}

	ServiceArguments map[string]string
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

func (s *Snap) ReadDqliteInfoYaml() (string, error) {
	return s.DqliteInfoYaml, nil
}

func (s *Snap) ReadDqliteClusterYaml() (string, error) {
	return s.DqliteClusterYaml, nil
}

func (s *Snap) WriteDqliteUpdateYaml(b []byte) error {
	s.WriteDqliteUpdateYamlCalledWith = append(s.WriteDqliteUpdateYamlCalledWith, b)
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
	return nil
}
