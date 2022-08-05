package proxy

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	internal "github.com/canonical/microk8s-cluster-agent/pkg/proxy/internal"
)

// APIServerProxy is a TCP proxy that forwards requests to the API Servers of the cluster.
type APIServerProxy struct {
	// TraefikConfigFile is the path to the traefik configuration file.
	// Note that this is only to stay backwards-compatible with the initial implementation of
	// worker nodes that was using Traefik for proxying requests to the control plane.
	// Only a subset of Traefik configuration flags are supported.
	TraefikConfigFile string
	// KubeconfigFile is the path to the kubeconfig file to use for updating the list of known apiservers.
	// The known apiservers are retrieved from `kubectl get endpoints kubernetes`.
	KubeconfigFile string
	// RefreshCh is used to check for updates in the list of control plane nodes in the cluster.
	RefreshCh <-chan time.Time
}

// Run starts the proxy.
func (p *APIServerProxy) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// create proxy context
		proxyCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		cfg, err := internal.LoadConfiguration(proxyCtx, p.TraefikConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		go p.startProxy(proxyCtx, cancel, cfg)
		go p.watchForNewEndpoints(proxyCtx, cancel, cfg)
		go p.watchForConfigFileChanges(proxyCtx, cancel, cfg)

		<-proxyCtx.Done()
	}
}

func (p *APIServerProxy) startProxy(ctx context.Context, cancel func(), cfg *internal.Configuration) {
	if err := startProxy(ctx, cfg.Listen, cfg.Endpoints); err != nil {
		log.Println(fmt.Errorf("apiserver proxy failed: %w", err))
		cancel()
	}
}

func (p *APIServerProxy) watchForNewEndpoints(ctx context.Context, cancel func(), cfg *internal.Configuration) {
	if p.RefreshCh == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.RefreshCh:
		}

		endpoints, err := getKubernetesEndpoints(ctx, p.KubeconfigFile)
		switch {
		case err != nil:
			log.Println(fmt.Errorf("failed to retrieve kubernetes endpoints: %w", err))
			continue
		case len(endpoints) == 0:
			log.Println("warning: empty list of endpoints, skipping update")
			continue
		case len(endpoints) == len(cfg.Endpoints) && reflect.DeepEqual(endpoints, cfg.Endpoints):
			continue
		}
		log.Println("updating endpoints")

		if err := internal.UpdateConfiguration(endpoints, p.TraefikConfigFile); err != nil {
			log.Printf("could not update configuration file with new endpoints: %q", err)
		}

		// cancel context in order to restart the proxy
		cancel()
		return
	}
}

func (p *APIServerProxy) watchForConfigFileChanges(ctx context.Context, cancel func(), cfg *internal.Configuration) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-cfg.ChangedCh:
			log.Println("Config file changed on disk, will restart proxy")
			cancel()
			return
		}
	}
}
