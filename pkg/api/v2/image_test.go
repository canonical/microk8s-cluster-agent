package v2_test

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	v2 "github.com/canonical/microk8s-cluster-agent/pkg/api/v2"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
)

func TestImageImport(t *testing.T) {
	s := &mock.Snap{
		SelfCallbackTokens: []string{"valid-token"},
	}

	apiv2 := &v2.API{Snap: s}

	t.Run("InvalidToken", func(t *testing.T) {
		reader := &recordingReader{
			Reader: bytes.NewBufferString("IMAGEDATA"),
		}
		err := apiv2.ImageImport(context.Background(), &v2.ImageImportRequest{
			Token:           "invalid-token",
			ImageDataReader: reader,
		})
		if err == nil {
			t.Fatal("Expected an error but did not receive any")
		}
		if len(reader.CountReadBytes) > 0 {
			t.Fatalf("Expected no Read calls for the image contents, but got one")
		}
	})

	t.Run("ValidToken", func(t *testing.T) {
		reader := &recordingReader{
			Reader: bytes.NewBufferString("IMAGEDATA"),
		}
		err := apiv2.ImageImport(context.Background(), &v2.ImageImportRequest{
			Token:           "valid-token",
			ImageDataReader: reader,
		})
		if err != nil {
			t.Fatalf("Expected no errors but received %q", err)
		}
		if len(reader.CountReadBytes) == 0 {
			t.Fatalf("Expected Read calls for the image contents, but did not get any")
		}

		expectedImportImage := []string{"IMAGEDATA"}
		if importedImages := s.ImportImageCalledWith; !reflect.DeepEqual(expectedImportImage, importedImages) {
			t.Fatalf("Expected ImportImage to be called with %q, but it was called with %q instead", expectedImportImage, importedImages)
		}
	})
}
