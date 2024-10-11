package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/canonical/microk8s-cluster-agent/pkg/middleware"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
)

type fakeNext struct {
	isCalled bool
}

func (f *fakeNext) next(w http.ResponseWriter, r *http.Request) {
	f.isCalled = true
}

func TestCAPIAuth(t *testing.T) {
	t.Run("NoTokenHeader", func(t *testing.T) {
		r := &http.Request{}
		fake := &fakeNext{}
		fn := middleware.CAPIAuthToken(fake.next, nil)
		w := httptest.NewRecorder()
		fn(w, r)

		g := NewWithT(t)

		g.Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
		g.Expect(fake.isCalled).To(BeFalse())
	})

	t.Run("InvalidToken", func(t *testing.T) {
		r := &http.Request{
			Header: http.Header{
				http.CanonicalHeaderKey(middleware.CAPIAuthTokenHeader): []string{"invalid-token"},
			},
		}
		fake := &fakeNext{}
		snapM := &mock.Snap{
			CAPIAuthTokenValid: false, // explicit
		}
		fn := middleware.CAPIAuthToken(fake.next, snapM)
		w := httptest.NewRecorder()
		fn(w, r)

		g := NewWithT(t)

		g.Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
		g.Expect(fake.isCalled).To(BeFalse())
	})

	t.Run("FailedToValidate", func(t *testing.T) {
		r := &http.Request{
			Header: http.Header{
				http.CanonicalHeaderKey(middleware.CAPIAuthTokenHeader): []string{"invalid-token"},
			},
		}
		fake := &fakeNext{}
		validateErr := errors.New("failed to validate")
		snapM := &mock.Snap{
			CAPIAuthTokenError: validateErr,
		}
		fn := middleware.CAPIAuthToken(fake.next, snapM)
		w := httptest.NewRecorder()
		fn(w, r)

		g := NewWithT(t)

		g.Expect(w.Result().StatusCode).To(Equal(http.StatusInternalServerError))
		g.Expect(fake.isCalled).To(BeFalse())
	})

	t.Run("Success", func(t *testing.T) {
		r := &http.Request{
			Header: http.Header{
				http.CanonicalHeaderKey(middleware.CAPIAuthTokenHeader): []string{"valid-token"},
			},
		}
		fake := &fakeNext{}
		snapM := &mock.Snap{
			CAPIAuthTokenValid: true,
		}
		fn := middleware.CAPIAuthToken(fake.next, snapM)
		w := httptest.NewRecorder()
		fn(w, r)

		g := NewWithT(t)

		g.Expect(fake.isCalled).To(BeTrue())
	})
}
