package proxy

type apiServerProxyConfig struct {
	Listen    string   `json:"listen"`
	Endpoints []string `json:"endpoints"`
}
