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

	// NOTE(Hue): This is technically not a good thing to do here,
	// because for whatever reason we might *want* to write an empty file.
	// And we're also making this function context aware (read coupled)
	// which is not a good thing.
	// However, since this function is **only** used to write `provider.yaml` file and
	// we know we don't want that file to be empty, we can safely return an error here.
	// Make sure to remove this check if the above statement is no longer true.
	if data == nil || len(b) == 0 {
		return fmt.Errorf("empty yaml data")
	}

	dir := filepath.Dir(file)
	tmpFile, err := os.CreateTemp(dir, "tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(b); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), file); err != nil {
		return fmt.Errorf("failed to rename temp file to target file: %w", err)
	}

	return nil
}
