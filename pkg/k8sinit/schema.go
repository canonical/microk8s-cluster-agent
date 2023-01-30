package k8sinit

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/version"
)

var (
	minimumConfigFileVersionRequired  = version.MustParseSemantic("0.1.0")
	maximumConfigFileVersionSupported = version.MustParseSemantic("0.1.0")
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
	// Version is the semantic version of the configuration file format.
	Version string `yaml:"version"`

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

	v, err := version.ParseSemantic(c.Version)
	switch {
	case err != nil:
		return nil, fmt.Errorf("could not parse config file version %q: %w", c.Version, err)
	case maximumConfigFileVersionSupported.LessThan(v):
		return nil, fmt.Errorf("config file version is %v but the maximum version supported is %v", c.Version, maximumConfigFileVersionSupported)
	case v.LessThan(minimumConfigFileVersionRequired):
		return nil, fmt.Errorf("config file version is %v but the minimum version required is %v", c.Version, minimumConfigFileVersionRequired)
	}

	return c, nil
}
