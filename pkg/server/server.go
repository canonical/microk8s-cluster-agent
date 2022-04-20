package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/canonical/microk8s-cluster-agent/pkg/httputil"
	"github.com/canonical/microk8s-cluster-agent/pkg/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Registrar func(server *http.ServeMux, middleware func(f http.HandlerFunc) http.HandlerFunc)

// NewServer creates a new *http.ServeMux and registers the MicroK8s cluster agent API endpoints.
func NewServer(timeout time.Duration, enableMetrics bool, registrars ...Registrar) *http.ServeMux {
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

	// Cluster Agent API
	for _, register := range registrars {
		register(server, withMiddleware)
	}

	return server
}
