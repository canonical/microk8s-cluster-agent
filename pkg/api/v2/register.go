package v2

import (
	"fmt"
	"net/http"

	"github.com/canonical/microk8s-cluster-agent/pkg/httputil"
)

// HTTPPrefix is the prefix for all v2 API routes.
const HTTPPrefix = "/cluster/api/v2.0"

// RegisterServer registers the Cluster API v2 endpoints on an HTTP server.
func (a *API) RegisterServer(server *http.ServeMux, middleware func(f http.HandlerFunc) http.HandlerFunc) {
	// POST v2/join
	server.HandleFunc(fmt.Sprintf("%s/join", HTTPPrefix), middleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		req := JoinRequest{}
		if err := httputil.UnmarshalJSON(r, &req); err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		req.RemoteAddress = r.RemoteAddr
		req.HostPort = r.Host

		response, err := a.Join(r.Context(), req)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}
		httputil.Response(w, response)
	}))
}
