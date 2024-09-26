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
				RunCommandErr: cmdErr,
			},
		}

		rc, err := apiv2.Remove(context.Background(), v2.RemoveRequest{HostPort: "1.1.1.1:1234"})

		g := NewWithT(t)
		g.Expect(err).To(MatchError(cmdErr))
		g.Expect(rc).To(Equal(http.StatusInternalServerError))
	})

	t.Run("RemovesSuccessfully", func(t *testing.T) {
		apiv2 := &v2.API{
			Snap: &mock.Snap{},
		}

		rc, err := apiv2.Remove(context.Background(), v2.RemoveRequest{HostPort: "1.1.1.1:1234"})

		g := NewWithT(t)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(rc).To(Equal(http.StatusOK))
	})
}
