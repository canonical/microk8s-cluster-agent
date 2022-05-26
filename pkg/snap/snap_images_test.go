package snap_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
	utiltest "github.com/canonical/microk8s-cluster-agent/pkg/util/test"
)

func TestImportImage(t *testing.T) {
	mockRunner := &utiltest.MockRunner{}
	s := snap.NewSnap("testdata", "testdata", mockRunner.Run)

	err := s.ImportImage(context.Background(), []byte(`IMAGEDATA`))
	if err != nil {
		t.Fatalf("Expected no error but got %q instead", err)
	}

	// hex of SHA256("IMAGEDATA")
	expectedFilename := "testdata/image-494d41474544415441e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.tar"

	expectedCmd := []string{fmt.Sprintf("testdata/microk8s-ctr.wrapper image import %s", expectedFilename)}
	if cmd := mockRunner.CalledWithCommand; !reflect.DeepEqual(cmd, expectedCmd) {
		t.Fatalf("Expected command %q but received %q instead", expectedCmd, cmd)
	}

	if util.FileExists(expectedFilename) {
		t.Fatalf("Expected file %q to not exist after image import, but it was not", expectedFilename)
	}
}
