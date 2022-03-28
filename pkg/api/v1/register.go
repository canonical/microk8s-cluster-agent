package v1

import (
	"fmt"
	"net/http"

	"github.com/canonical/microk8s-cluster-agent/pkg/httputil"
)

// HTTPPrefix is the prefix for all v1 API routes.
const HTTPPrefix = "/cluster/api/1.0"

// RegisterServer registers the Cluster API v2 endpoints on an HTTP server.
func (a *API) RegisterServer(server *http.ServeMux, middleware func(f http.HandlerFunc) http.HandlerFunc) {
	// POST /v1/join
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

		// Set remote address from request object.
		req.RemoteAddress = r.RemoteAddr

		resp, err := a.Join(r.Context(), req)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Response(w, resp)
	}))

	// POST v1/sign-cert
	server.HandleFunc(fmt.Sprintf("%s/sign-cert", HTTPPrefix), middleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		req := SignCertRequest{}
		if err := httputil.UnmarshalJSON(r, &req); err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		resp, err := a.SignCert(r.Context(), req)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Response(w, resp)
	}))

	// POST v1/configure
	server.HandleFunc(fmt.Sprintf("%s/configure", HTTPPrefix), middleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		req := ConfigureRequest{}
		if err := httputil.UnmarshalJSON(r, &req); err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		err := a.Configure(r.Context(), req)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}
		httputil.Response(w, map[string]string{"result": "ok"})
	}))

	// POST v1/upgrade
	server.HandleFunc(fmt.Sprintf("%s/upgrade", HTTPPrefix), middleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		req := UpgradeRequest{}
		if err := httputil.UnmarshalJSON(r, &req); err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		err := a.Upgrade(r.Context(), req)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}
		httputil.Response(w, map[string]string{"result": "ok"})
	}))
}
