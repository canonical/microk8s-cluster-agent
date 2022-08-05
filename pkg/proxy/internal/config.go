package internal

import (
	"context"
	"fmt"
	"log"
	"sort"
)

// Configuration is configuration for the apiserver proxy
type Configuration struct {
	Listen    string
	Endpoints []string
	ChangedCh chan struct{}
}

// LoadConfiguration loads traefik-compatible configuration for the API server proxy.
func LoadConfiguration(ctx context.Context, traefikConfigFile string) (*Configuration, error) {
	cfg := &Configuration{}
	var traefikConfig TraefikConfiguration
	if err := loadYamlWarnStrict(traefikConfigFile, &traefikConfig); err != nil {
		return nil, fmt.Errorf("failed to load traefik configuration: %w", err)
	}

	// listen address
	cfg.Listen = traefikConfig.EntryPoints.APIServer.Address
	if cfg.Listen == "" {
		cfg.Listen = ":16443"
	}

	// endpoints
	var providerConfig ProviderConfiguration
	if err := loadYamlWarnStrict(traefikConfig.Providers.File.Filename, &providerConfig); err != nil {
		return nil, fmt.Errorf("failed to load provider configuration: %w", err)
	}

	servers := providerConfig.TCP.Services.APIServer.LoadBalancer.Servers
	if len(servers) == 0 {
		return nil, fmt.Errorf("empty list of control plane endpoints")
	}
	cfg.Endpoints = make([]string, 0, len(servers))
	for i := 0; i < len(servers); i++ {
		cfg.Endpoints = append(cfg.Endpoints, servers[i].Address)
	}
	sort.Strings(cfg.Endpoints)

	cfg.ChangedCh = make(chan struct{}, 1)
	if traefikConfig.Providers.File.Watch {
		watcher, err := newWatcher(traefikConfig.Providers.File.Filename)
		if err != nil {
			log.Printf("failed to setup file watch: %v", err)
		} else {
			go func() {
				if err := notifyOnChange(ctx, watcher, cfg.ChangedCh); err != nil {
					log.Printf("error while watching providers file: %v", err)
				}
			}()
		}
	}

	return cfg, nil
}

// UpdateConfiguration updates the list of control plane endpoints in the traefik provider configuration file.
func UpdateConfiguration(endpoints []string, traefikConfigFile string) error {
	// attempt to load existing configuration and see the location of the provider file
	var traefikConfig TraefikConfiguration
	providerConfigFile := ""
	if err := loadYamlWarnStrict(traefikConfigFile, &traefikConfig); err == nil {
		providerConfigFile = traefikConfig.Providers.File.Filename
	}
	if providerConfigFile == "" {
		providerConfigFile = getDefaultProviderFile()
	}

	servers := make([]Server, 0, len(endpoints))
	for i := 0; i < len(endpoints); i++ {
		servers = append(servers, Server{Address: endpoints[i]})
	}

	providerConfig := ProviderConfiguration{
		TCP: TCPProviderConfiguration{
			Services: ExpectedServices{
				APIServer: Service{
					LoadBalancer: LoadBalancerService{
						Servers: servers,
					},
				},
			},
			Routers: ExpectedRouters{
				Router1: Router{
					Rule:    "HostSNI(`*`)",
					Service: "kube-apiserver",
					TLS: RouterTLS{
						Passthrough: true,
					},
				},
			},
		},
	}

	if err := writeYaml(providerConfigFile, providerConfig); err != nil {
		return fmt.Errorf("failed to update provider config: %w", err)
	}
	return nil
}
