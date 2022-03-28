package server

import (
	"fmt"
	"net/http"
	"time"

	v1 "github.com/canonical/microk8s-cluster-agent/pkg/api/v1"
	v2 "github.com/canonical/microk8s-cluster-agent/pkg/api/v2"
	"github.com/canonical/microk8s-cluster-agent/pkg/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewServer creates a new *http.ServeMux and registers the MicroK8s cluster agent API endpoints.
func NewServer(timeout time.Duration, enableMetrics bool, apiv1 *v1.API, apiv2 *v2.API) *http.ServeMux {
	server := http.NewServeMux()

	withMiddleware := func(f http.HandlerFunc) http.HandlerFunc {
		timeoutMiddleware := middleware.Timeout(timeout)
		return middleware.Log(timeoutMiddleware(f))
	}

	// Default handler
	server.HandleFunc("/", withMiddleware(func(w http.ResponseWriter, r *http.Request) {
		httputil.Error(w, http.StatusNotFound, fmt.Errorf("not found"))
	}))

	// Prometheus metrics
	if enableMetrics {
		server.HandleFunc("/metrics", withMiddleware(promhttp.Handler().ServeHTTP))
	}

	// POST /v1/join
	server.HandleFunc(fmt.Sprintf("%s/join", ClusterAPIV1), withMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		req := v1.JoinRequest{}
		if err := UnmarshalJSON(r, &req); err != nil {
			HTTPError(w, http.StatusBadRequest, err)
			return
		}

	// Cluster Agent API
	apiv1.RegisterServer(server, withMiddleware)
	apiv2.RegisterServer(server, withMiddleware)

	return server
}
