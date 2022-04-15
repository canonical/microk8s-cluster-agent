package v1_test

import (
	"context"
	"reflect"
	"testing"

	v1 "github.com/canonical/microk8s-cluster-agent/pkg/api/v1"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
)

func TestSignCert(t *testing.T) {
	s := &mock.Snap{
		CertificateRequestTokens: []string{"valid-token"},
		SignedCertificate:        "CERT DATA",
	}
	apiv1 := &v1.API{Snap: s}
	t.Run("InvalidToken", func(t *testing.T) {
		resp, err := apiv1.SignCert(context.Background(), v1.SignCertRequest{
			Token: "invalid-token",
		})
		if err == nil {
			t.Fatal("Expected an error but did not receive any")
		}
		if resp != nil {
			t.Fatalf("Expected a nil response but received %#v", resp)
		}
		expectedConsumeCertificateRequestTokenCalledWith := []string{"invalid-token"}
		if !reflect.DeepEqual(expectedConsumeCertificateRequestTokenCalledWith, s.ConsumeCertificateRequestTokenCalledWith) {
			t.Fatalf("Expected ConsumeCertificateRequestToken to be called with %v, but it was called with %v instead", expectedConsumeCertificateRequestTokenCalledWith, s.ConsumeCertificateRequestTokenCalledWith)
		}
	})

	t.Run("Success", func(t *testing.T) {
		s.ConsumeCertificateRequestTokenCalledWith = nil

		resp, err := apiv1.SignCert(context.Background(), v1.SignCertRequest{
			Token:                     "valid-token",
			CertificateSigningRequest: "CSR DATA",
		})
		if err != nil {
			t.Fatalf("Expected no error but received %q", err)
		}
		if resp == nil {
			t.Fatal("Expected a non-nil response")
		}
		expectedCalledWith := []string{"CSR DATA"}
		if !reflect.DeepEqual(expectedCalledWith, s.SignCertificateCalledWith) {
			t.Fatalf("Expected SignCertificate called with %v, but it was called with %v instead", expectedCalledWith, s.SignCertificateCalledWith)
		}
		expectedConsumeCertificateRequestTokenCalledWith := []string{"valid-token"}
		if !reflect.DeepEqual(expectedConsumeCertificateRequestTokenCalledWith, s.ConsumeCertificateRequestTokenCalledWith) {
			t.Fatalf("Expected ConsumeCertificateRequestToken to be called with %v, but it was called with %v instead", expectedConsumeCertificateRequestTokenCalledWith, s.ConsumeCertificateRequestTokenCalledWith)
		}
		if resp.Certificate != s.SignedCertificate {
			t.Fatalf("Expected ceritificate to be %q, but it was %q instead", s.SignedCertificate, resp.Certificate)
		}
	})
}
