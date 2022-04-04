package v1

import (
	"context"
	"fmt"
)

// SignCertRequest is the request message for the sign-cert endpoint.
type SignCertRequest struct {
	// Token is the certificate request token.
	Token string `json:"token"`
	// CertificateSigningRequest is the signing request file contents.
	CertificateSigningRequest string `json:"request"`
}

// SignCertResponse is the response message for the sign-cert endpoint.
type SignCertResponse struct {
	// Certificate is the contents of the signed certificate.
	Certificate string `json:"certificate"`
}

// SignCert implements "POST CLUSTER_API_V1/sign-cert".
func (a *API) SignCert(ctx context.Context, req SignCertRequest) (*SignCertResponse, error) {
	if !a.Snap.IsValidCertificateRequestToken(req.Token) {
		return nil, fmt.Errorf("invalid token")
	}
	if err := a.Snap.RemoveCertificateRequestToken(req.Token); err != nil {
		return nil, fmt.Errorf("failed to remove certificate request token: %w", err)
	}

	cert, err := a.Snap.SignCertificate(ctx, []byte(req.CertificateSigningRequest))
	if err != nil {
		return nil, fmt.Errorf("failed to sign certificate: %w", err)
	}
	return &SignCertResponse{Certificate: string(cert)}, nil
}
