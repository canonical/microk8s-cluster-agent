package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func getDefaultProviderFile() string {
	snapData := os.Getenv("SNAP_DATA")
	if snapData == "" {
		snapData = "/var/snap/microk8s/current"
	}
	return filepath.Join(filepath.Dir(snapData), "current", "args", "traefik", "provider.yaml")
}

// loadYamlWarnStrict will load a yaml file and unmarshal it to dest.
// loadYamlWarnStrict will first attempt to yaml.UnmarshalStrict() the file contents. if that
// fails, an error is logged and then yaml.Unmarshal() will be attempted.
// loadYamlWarnStrict is useful to us since we only support a subset of Traefik configs, and
// we must warn users in case they want to use other configs.
func loadYamlWarnStrict(file string, dest interface{}) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	err = yaml.UnmarshalStrict(b, dest)
	if err != nil {
		log.Printf("WARNING: Unmarshaling yaml file failed, error was %q. This might due to using unsupported configuration fields. Note that only a subset of traefik configuration fields are supported by the API server proxy. Refer to the documentation for more details.\n", err)
	}

	// Proceed with unmarshaling
	err = yaml.Unmarshal(b, dest)
	if err != nil {
		return fmt.Errorf("unmarshal configuration failed: %w", err)
	}
	return nil
}

func writeYaml(file string, data interface{}) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}
	if err := os.WriteFile(file, b, 0664); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
