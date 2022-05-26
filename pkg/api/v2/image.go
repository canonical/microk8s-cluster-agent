package v2

import (
	"context"
	"fmt"
	"io"
	"net/http"
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
func (a *API) ImageImport(ctx context.Context, req *ImageImportRequest) (int, error) {
	if !a.Snap.ConsumeSelfCallbackToken(req.Token) {
		return http.StatusUnauthorized, fmt.Errorf("invalid token")
	}

	if req.ImageDataReader == nil {
		return http.StatusBadRequest, fmt.Errorf("no image data")
	}

	// TODO(neoaggelos): we might want to ignore the errors
	if err := a.Snap.ImportImage(ctx, req.ImageDataReader); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to import the image: %w", err)
	}

	return http.StatusOK, nil
}
