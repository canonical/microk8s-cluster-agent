package v2_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"

	v2 "github.com/canonical/microk8s-cluster-agent/pkg/api/v2"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
)

func TestRemove(t *testing.T) {
	t.Run("RemoveFails", func(t *testing.T) {
		cmdErr := errors.New("failed to run command")
		apiv2 := &v2.API{
			Snap: &mock.Snap{
				RunCommandErr:      cmdErr,
				CAPIAuthTokenValid: true,
			},
		}

		rc, err := apiv2.RemoveFromDqlite(context.Background(), v2.RemoveFromDqliteRequest{RemoveEndpoint: "1.1.1.1:1234"}, "token")

		g := NewWithT(t)
		g.Expect(err).To(MatchError(cmdErr))
		g.Expect(rc).To(Equal(http.StatusInternalServerError))
	})

	t.Run("InvalidToken", func(t *testing.T) {
		apiv2 := &v2.API{
			Snap: &mock.Snap{
				CAPIAuthTokenValid: false, // explicitly set to false
			},
		}

		rc, err := apiv2.RemoveFromDqlite(context.Background(), v2.RemoveFromDqliteRequest{RemoveEndpoint: "1.1.1.1:1234"}, "token")

		g := NewWithT(t)
		g.Expect(err).To(HaveOccurred())
		g.Expect(rc).To(Equal(http.StatusUnauthorized))
	})

	t.Run("TokenFileNotFound", func(t *testing.T) {
		tokenErr := errors.New("token file not found")
		apiv2 := &v2.API{
			Snap: &mock.Snap{
				CAPIAuthTokenError: tokenErr,
			},
		}

		rc, err := apiv2.RemoveFromDqlite(context.Background(), v2.RemoveFromDqliteRequest{RemoveEndpoint: "1.1.1.1:1234"}, "token")

		g := NewWithT(t)
		g.Expect(err).To(MatchError(tokenErr))
		g.Expect(rc).To(Equal(http.StatusUnauthorized))
	})

	t.Run("RemovesSuccessfully", func(t *testing.T) {
		apiv2 := &v2.API{
			Snap: &mock.Snap{
				CAPIAuthTokenValid: true,
			},
		}

		rc, err := apiv2.RemoveFromDqlite(context.Background(), v2.RemoveFromDqliteRequest{RemoveEndpoint: "1.1.1.1:1234"}, "token")

		g := NewWithT(t)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(rc).To(Equal(http.StatusOK))
	})
}
