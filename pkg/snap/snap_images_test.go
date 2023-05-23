package snap_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
	utiltest "github.com/canonical/microk8s-cluster-agent/pkg/util/test"
	. "github.com/onsi/gomega"
)

var mockCtr = `#!/bin/bash

# this is a mock for the $SNAP/microk8s-ctr.wrapper script, used to
# ensure that cluster-agent is calling it properly

echo $0 $@ > testdata/arguments
cat > testdata/stdin
`

func TestImportImage(t *testing.T) {
	if err := os.WriteFile("testdata/microk8s-ctr.wrapper", []byte(mockCtr), 0755); err != nil {
		t.Fatalf("Failed to initialize mock ctr command: %v", err)
	}
	defer func() {
		os.Remove("testdata/microk8s-ctr.wrapper")
		os.Remove("testdata/stdin")
		os.Remove("testdata/arguments")
	}()
	mockRunner := &utiltest.MockRunner{}
	s := snap.NewSnap("testdata", "testdata", snap.WithCommandRunner(mockRunner.Run))

	g := NewWithT(t)
	err := s.ImportImage(context.Background(), bytes.NewBufferString("IMAGEDATA"))
	g.Expect(err).To(BeNil())

	cmd, err := util.ReadFile("testdata/arguments")
	g.Expect(err).To(BeNil())
	g.Expect(strings.TrimSpace(cmd)).To(Equal(fmt.Sprintf("testdata/microk8s-ctr.wrapper image import --platform %s -", runtime.GOARCH)))

	stdin, err := util.ReadFile("testdata/stdin")
	g.Expect(err).To(BeNil())
	g.Expect(stdin).To(Equal("IMAGEDATA"))
}
