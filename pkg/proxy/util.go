package proxy

import (
	"encoding/json"
	"fmt"
	"os"
)

func updateConfigFile(configFile string, config *apiServerProxyConfig) error {
	b, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.WriteFile(configFile, b, 0644); err != nil {
		return fmt.Errorf("failed to update configuration file %q: %w", configFile, err)
	}
	return nil
}
