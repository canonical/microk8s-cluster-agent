package v2

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
)

// ImageImportRequest is a request for importing an image to the container runtime.
type ImageImportRequest struct {
	// Token is the certificate request token.
	Token string
	// ImageDataReader is an io.Reader that fetches the OCI image tarball bytes.
	// ImageDataReader is a reader interface instead of a bytes buffer to avoid reading
	// the contents of the image when the token is invalid.
	ImageDataReader io.Reader
}

// ImageImport implements "POST CLUSTER_API_V2/images/import"
func (a *API) ImageImport(ctx context.Context, req *ImageImportRequest) error {
	if !a.Snap.ConsumeSelfCallbackToken(req.Token) {
		return fmt.Errorf("invalid token")
	}

	if req.ImageDataReader == nil {
		return fmt.Errorf("no image data")
	}
	b, err := ioutil.ReadAll(req.ImageDataReader)
	if err != nil {
		return fmt.Errorf("failed to read the image data contents: %w", err)
	}

	// TODO(neoaggelos): we might want to ignore the errors
	if err := a.Snap.ImportImage(ctx, b); err != nil {
		return fmt.Errorf("failed to import the image: %w", err)
	}

	return nil
}
