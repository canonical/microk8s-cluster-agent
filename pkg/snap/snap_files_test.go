package snap_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
	. "github.com/onsi/gomega"
)

var (
	dqliteCrt         = "DQLITE CERTIFICATE DATA"
	dqliteKey         = "DQLITE KEY DATA"
	dqliteInfoYaml    = "DQLITE INFO YAML"
	dqliteClusterYaml = "DQLITE CLUSTER YAML"
	caCrt             = "CA CERTIFICATE DATA"
	caKey             = "CA KEY DATA"
	saKey             = "SERVICE ACCOUNT KEY DATA"
	cniYaml           = "CNI DATA"
	etcdCa            = "ETCD CA DATA"
	etcdCrt           = "ETCD CERT DATA"
	etcdKey           = "ETCD KEY DATA"
)

func TestFiles(t *testing.T) {
	g := NewWithT(t)

	// Create test data
	for file, contents := range map[string]string{
		"testdata/var/kubernetes/backend/cluster.crt":  dqliteCrt,
		"testdata/var/kubernetes/backend/cluster.key":  dqliteKey,
		"testdata/var/kubernetes/backend/info.yaml":    dqliteInfoYaml,
		"testdata/var/kubernetes/backend/cluster.yaml": dqliteClusterYaml,
		"testdata/certs/ca.crt":                        caCrt,
		"testdata/certs/ca.key":                        caKey,
		"testdata/certs/etcd-ca.crt":                   etcdCa,
		"testdata/certs/etcd-client.crt":               etcdCrt,
		"testdata/certs/etcd-client.key":               etcdKey,
		"testdata/certs/serviceaccount.key":            saKey,
		"testdata/args/cni-network/cni.yaml":           cniYaml,
	} {
		g.Expect(os.MkdirAll(filepath.Dir(file), 0755)).To(BeNil())
		g.Expect(os.WriteFile(file, []byte(contents), 0660)).To(BeNil())
		defer os.RemoveAll(filepath.Dir(file))
	}

	s := snap.NewSnap("testdata", "testdata", "testdata")

	for _, tc := range []struct {
		name     string
		retrieve func() (string, error)
		expected string
	}{
		{name: "CA", retrieve: s.ReadCA, expected: caCrt},
		{name: "CAKey", retrieve: s.ReadCAKey, expected: caKey},
		{name: "ServiceAccountKey", retrieve: s.ReadServiceAccountKey, expected: saKey},
		{name: "DqliteInfoYaml", retrieve: s.ReadDqliteInfoYaml, expected: dqliteInfoYaml},
		{name: "DqliteClusterYaml", retrieve: s.ReadDqliteClusterYaml, expected: dqliteClusterYaml},
		{name: "DqliteCrt", retrieve: s.ReadDqliteCert, expected: dqliteCrt},
		{name: "DqliteKey", retrieve: s.ReadDqliteKey, expected: dqliteKey},
		{name: "CNI", retrieve: s.ReadCNIYaml, expected: cniYaml},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			v, err := tc.retrieve()
			g.Expect(err).To(BeNil())
			g.Expect(v).To(Equal(tc.expected))
		})
	}

	for _, tc := range []struct {
		name  string
		write func([]byte) error
		file  string
	}{
		{name: "CNI", write: s.WriteCNIYaml, file: "testdata/args/cni-network/cni.yaml"},
		{name: "DqliteUpdateYaml", write: s.WriteDqliteUpdateYaml, file: "testdata/var/kubernetes/backend/update.yaml"},
		{name: "WriteCSRConfig", write: s.WriteCSRConfig, file: "testdata/certs/csr.conf.template"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			magicVal := util.NewRandomString(util.Alpha, 10)
			err := tc.write([]byte(magicVal))
			g.Expect(err).To(BeNil())

			v, err := util.ReadFile(tc.file)
			g.Expect(err).To(BeNil())
			g.Expect(v).To(Equal(magicVal))
		})
	}

	t.Run("EtcdCertificates", func(t *testing.T) {
		for _, tc := range []struct {
			name           string
			apiserverArgs  []string
			expectEtcdCA   string
			expectEtcdCert string
			expectEtcdKey  string
			expectErr      bool
		}{
			{
				name:           "all",
				apiserverArgs:  []string{"--etcd-cafile=$SNAP_DATA/certs/etcd-ca.crt", "--etcd-certfile=${SNAP_DATA}/certs/etcd-client.crt", `--etcd-keyfile="testdata/certs/etcd-client.key"`},
				expectEtcdCA:   etcdCa,
				expectEtcdCert: etcdCrt,
				expectEtcdKey:  etcdKey,
			},
			{
				name: "none",
			},
			{
				name:          "only-ca",
				apiserverArgs: []string{"--etcd-cafile=$SNAP_DATA/certs/etcd-ca.crt"},
				expectEtcdCA:  etcdCa,
			},
			{
				name:          "missing-file",
				apiserverArgs: []string{"--etcd-cafile=$SNAP_DATA/certs/etcd-caa.crt"},
				expectErr:     true,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)
				g.Expect(os.WriteFile("testdata/args/kube-apiserver", []byte(strings.Join(tc.apiserverArgs, "\n")), 0755)).To(BeNil())
				defer os.RemoveAll("testdata/args/kube-apiserver")

				ca, crt, key, err := s.ReadEtcdCertificates()
				g.Expect(ca).To(Equal(tc.expectEtcdCA))
				g.Expect(crt).To(Equal(tc.expectEtcdCert))
				g.Expect(key).To(Equal(tc.expectEtcdKey))
				if tc.expectErr {
					g.Expect(err).NotTo(BeNil())
				} else {
					g.Expect(err).To(BeNil())
				}
			})
		}
	})
}
