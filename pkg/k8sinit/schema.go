package k8sinit

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"
)

const (
	minimumConfigFileVersionRequired  = 1
	maximumConfigFileVersionSupported = 1
)

// AddonConfiguration specifies an addon to be enabled or disabled.
type AddonConfiguration struct {
	// Name of the addon to configure.
	Name string `yaml:"name"`

	// Enable controls whether we will enable the addon.
	Enable bool `yaml:"enable"`

	// Arguments is optional arguments passed to the addon enable or disable operation.
	Arguments []string `yaml:"args"`
}

// Configuration is the top-level definition for MicroK8s configuration files.
type Configuration struct {
	// Version of the configuration file format. Reserved for backwards-compatibility.
	Version int `yaml:"version"`

	// Addons is a list of addons to enable and/or disable.
	Addons []AddonConfiguration `yaml:"addons"`
}

// ParseConfiguration tries to parse a Configuration object from YAML data.
func ParseConfiguration(input []byte) (*Configuration, error) {
	c := &Configuration{}

	if strictParseErr := yaml.UnmarshalStrict(input, c); strictParseErr != nil {
		// If non-strict parsing also fails, then raise the error
		if err := yaml.Unmarshal(input, c); err != nil {
			return nil, fmt.Errorf("could not parse configuration: %w", err)
		}

		log.Printf("WARNING: configuration may contain unknown fields (error was %q).", strictParseErr)
		log.Printf("Any unknown fields will be ignored")
	}

	return c, nil
}
