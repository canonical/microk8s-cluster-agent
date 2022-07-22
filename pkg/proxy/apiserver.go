package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"time"
)

type apiServerProxyConfig struct {
	Listen    string   `json:"listen"`
	Endpoints []string `json:"endpoints"`
}

type APIServerProxy struct {
	ConfigFile     string
	KubeconfigFile string
	RefreshCh      <-chan time.Time
}

func (p *APIServerProxy) Run(ctx context.Context) error {
	b, err := os.ReadFile(p.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	cfg := &apiServerProxyConfig{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	if len(cfg.Endpoints) == 0 {
		return fmt.Errorf("fatal error: no known control plane nodes are left")
	}
	sort.Strings(cfg.Endpoints)
	if cfg.Listen == "" {
		cfg.Listen = "127.0.0.1:16443"
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		proxyCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
		nextUpdate:
			for {
				select {
				case <-proxyCtx.Done():
					return
				case <-ctx.Done():
					return
				case <-p.RefreshCh:
				}

				endpoints, err := getKubernetesEndpoints(proxyCtx, p.KubeconfigFile)
				switch {
				case err != nil:
					log.Println(fmt.Errorf("failed to retrieve kubernetes endpoints: %w", err))
					continue nextUpdate
				case len(endpoints) == 0:
					log.Println("warning: empty list of endpoints, skipping update")
					continue nextUpdate
				case len(endpoints) == len(cfg.Endpoints) && reflect.DeepEqual(endpoints, cfg.Endpoints):
					continue nextUpdate
				}
				log.Println("updating endpoints")

				cfg.Endpoints = endpoints
				// first update the configuration file to preserve changes
				if err := updateConfigFile(p.ConfigFile, cfg); err != nil {
					log.Println(fmt.Errorf("warning: failed to update configuration file: %w", err))
				}

				// cancel context in order to restart the proxy
				cancel()
				return
			}
		}()

		if err := startProxy(proxyCtx, cfg.Listen, cfg.Endpoints); err != nil {
			log.Println(fmt.Errorf("apiserver proxy failed: %w", err))
			cancel()
		}
	}
}
