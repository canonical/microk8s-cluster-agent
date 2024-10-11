package server

import (
	"fmt"
	"net/http"
	"time"

	v1 "github.com/canonical/microk8s-cluster-agent/pkg/api/v1"
	v2 "github.com/canonical/microk8s-cluster-agent/pkg/api/v2"
	"github.com/canonical/microk8s-cluster-agent/pkg/httputil"
	"github.com/canonical/microk8s-cluster-agent/pkg/middleware"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewServeMux creates a new *http.ServeMux and registers the MicroK8s cluster agent API endpoints.
func NewServeMux(timeout time.Duration, enableMetrics bool, apiv1 *v1.API, apiv2 *v2.API) *http.ServeMux {
	server := http.NewServeMux()

	withMiddleware := func(f http.HandlerFunc) http.HandlerFunc {
		timeoutMiddleware := middleware.Timeout(timeout)
		return middleware.Log(timeoutMiddleware(f))
	}

	capiAuthMiddleWare := func(f http.HandlerFunc, snp snap.Snap) http.HandlerFunc {
		return middleware.CAPIAuthToken(f, snp)
	}

	// Default handler
	server.HandleFunc("/", withMiddleware(func(w http.ResponseWriter, r *http.Request) {
		httputil.Error(w, http.StatusNotFound, fmt.Errorf("not found"))
	}))

	// GET /health
	server.HandleFunc("/health", withMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		httputil.Response(w, map[string]string{"status": "OK"})
	}))

	// Prometheus metrics
	if enableMetrics {
		server.HandleFunc("/metrics", withMiddleware(promhttp.Handler().ServeHTTP))
	}

	// Cluster Agent API
	apiv1.RegisterServer(server, withMiddleware)
	apiv2.RegisterServer(server, withMiddleware, capiAuthMiddleWare)

	return server
}
