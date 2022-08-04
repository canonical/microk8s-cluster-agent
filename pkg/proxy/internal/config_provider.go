package internal

// ProviderConfiguration is traefik-compatible configuration.
type ProviderConfiguration struct {
	TCP TCPProviderConfiguration `yaml:"tcp"`
}

// TCPProviderConfiguration is traefik-compatible configuration.
type TCPProviderConfiguration struct {
	Routers  ExpectedRouters  `yaml:"routers"`
	Services ExpectedServices `yaml:"services"`
}

// ExpectedServices is traefik-compatible configuration.
type ExpectedServices struct {
	APIServer Service `yaml:"kube-apiserver"`
}

// Service is traefik-compatible configuration.
type Service struct {
	LoadBalancer LoadBalancerService `yaml:"loadBalancer"`
}

// LoadBalancerService is traefik-compatible configuration.
type LoadBalancerService struct {
	Servers []Server `yaml:"servers"`
}

// Server is traefik-compatible configuration.
type Server struct {
	Address string `yaml:"address"`
}

// ExpectedRouters is traefik-compatible configuration.
type ExpectedRouters struct {
	Router1 Router `yaml:"Router-1"`
}

// Router is traefik-compatible configuration.
type Router struct {
	Rule    string    `yaml:"rule"`
	Service string    `yaml:"service"`
	TLS     RouterTLS `yaml:"tls"`
}

// RouterTLS is traefik-compatible configuration.
type RouterTLS struct {
	Passthrough bool `yaml:"passthrough"`
}
