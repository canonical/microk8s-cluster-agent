package middleware

import (
	"fmt"
	"net/http"

	"github.com/canonical/microk8s-cluster-agent/pkg/httputil"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

const (
	// CAPIAuthTokenHeader is the header used to pass the CAPI auth token.
	CAPIAuthTokenHeader = "capi-auth-token"
)

func CAPIAuthToken(next http.HandlerFunc, snap snap.Snap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(CAPIAuthTokenHeader)
		fmt.Println(r.Header, "-->", r.Header.Get(CAPIAuthTokenHeader))
		fmt.Println("token", token)
		if token == "" {
			httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("missing CAPI auth token"))
			return
		}

		isValid, err := snap.IsCAPIAuthTokenValid(token)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("failed to validate CAPI auth token: %w", err))
			return
		}

		if !isValid {
			httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("invalid CAPI auth token %q", token))
			return
		}

		next.ServeHTTP(w, r)
	}
}
