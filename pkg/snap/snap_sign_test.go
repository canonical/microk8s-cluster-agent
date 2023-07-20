package snap_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
	utiltest "github.com/canonical/microk8s-cluster-agent/pkg/util/test"
	. "github.com/onsi/gomega"
)

var mockSign = `#!/bin/bash
# this is a mock for the $SNAP/actions/common/utils.sh sign_certificate script, used to
# ensure that cluster-agent is calling it properly

echo $0 $@ > testdata/arguments
cat > testdata/stdin

echo MOCK CERT
`

func TestSignCertificate(t *testing.T) {
	if err := os.MkdirAll("testdata/actions/common", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile("testdata/actions/common/utils.sh", []byte(mockSign), 0755); err != nil {
		t.Fatalf("Failed to initialize mock utils command: %v", err)
	}
	defer func() {
		os.RemoveAll("testdata/actions")
		os.Remove("testdata/stdin")
		os.Remove("testdata/arguments")
	}()
	mockRunner := &utiltest.MockRunner{}
	s := snap.NewSnap("testdata", "testdata", snap.WithCommandRunner(mockRunner.Run))

	g := NewWithT(t)
	b, err := s.SignCertificate(context.Background(), []byte("MOCK CSR"))
	g.Expect(err).To(BeNil())
	g.Expect(strings.TrimSpace(string(b))).To(Equal("MOCK CERT"))

	cmd, err := util.ReadFile("testdata/arguments")
	g.Expect(err).To(BeNil())
	g.Expect(strings.TrimSpace(cmd)).To(Equal("testdata/actions/common/utils.sh sign_certificate"))

	stdin, err := util.ReadFile("testdata/stdin")
	g.Expect(err).To(BeNil())
	g.Expect(strings.TrimSpace(stdin)).To(Equal("MOCK CSR"))
}
