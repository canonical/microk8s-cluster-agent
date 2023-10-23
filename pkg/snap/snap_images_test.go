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
)

var mockCtr = `#!/bin/bash

# this is a mock for the $SNAP/microk8s-ctr.wrapper script, used to
# ensure that cluster-agent is calling it properly

echo $0 $@ > testdata/arguments
cat > testdata/stdin
`

func TestImportImage(t *testing.T) {
	if err := os.MkdirAll("testdata/bin", 0700); err != nil {
		t.Fatalf("Failed to intialize mock bin dir: %v", err)
	}
	if err := os.WriteFile("testdata/bin/ctr", []byte(mockCtr), 0755); err != nil {
		t.Fatalf("Failed to initialize mock ctr command: %v", err)
	}
	defer func() {
		os.RemoveAll("testdata/bin")
		os.Remove("testdata/stdin")
		os.Remove("testdata/arguments")
	}()
	mockRunner := &utiltest.MockRunner{}
	s := snap.NewSnap("testdata", "testdata", "testdata/common", snap.WithCommandRunner(mockRunner.Run))

	err := s.ImportImage(context.Background(), bytes.NewBufferString("IMAGEDATA"))
	if err != nil {
		t.Fatalf("Expected no error but got %q instead", err)
	}

	expectedCmd := fmt.Sprintf("testdata/bin/ctr --namespace k8s.io --address testdata/common/run/containerd.sock image import --platform %s -", runtime.GOARCH)
	if cmd, err := util.ReadFile("testdata/arguments"); err != nil || strings.TrimSpace(cmd) != expectedCmd {
		t.Fatalf("Expected command to be %q but it was %q instead", expectedCmd, cmd)
	}

	expectedStdin := "IMAGEDATA"
	if stdin, err := util.ReadFile("testdata/stdin"); err != nil || stdin != "IMAGEDATA" {
		t.Fatalf("Expected stdin to be %q but it was %q instead", expectedStdin, stdin)
	}
}
