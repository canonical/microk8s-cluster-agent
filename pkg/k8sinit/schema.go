package k8sinit

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/version"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

var (
	minimumConfigFileVersionRequired  = version.MustParseSemantic("0.1.0")
	maximumConfigFileVersionSupported = version.MustParseSemantic("0.1.0")

	// errEmptyConfig is an ignorable error when parsing empty YAML documents
	errEmptyConfig = fmt.Errorf("empty configuration object")
)

// JoinConfiguration is configuration to join the local node to an existing MicroK8s cluster.
type JoinConfiguration struct {
	// URL is the URL passed to the microk8s join command.
	URL string `yaml:"url"`

	// Worker is true when joining the cluster as a worker-only node.
	Worker bool `yaml:"worker"`
}

// AddonConfiguration specifies an addon to be enabled or disabled.
type AddonConfiguration struct {
	// Name of the addon to configure.
	Name string `yaml:"name"`

	// Disable the addon.
	Disable bool `yaml:"disable"`

	// Arguments is optional arguments passed to the addon enable or disable operation.
	Arguments []string `yaml:"args"`
}

// AddonRepositoryConfiguration specifies an addon repository to be added.
type AddonRepositoryConfiguration struct {
	// Name of the addon repository.
	Name string `yaml:"name"`
	// URL of the addon repository.
	URL string `yaml:"url"`
	// Reference is an optional reference to check out instead of the default branch.
	Reference string `yaml:"reference"`
}

// MultiPartConfiguration is a configuration split into multiple parts.
type MultiPartConfiguration struct {
	// Parts are configuration objects that are meant to be applied in order.
	Parts []*Configuration
}

// Configuration is the top-level definition for MicroK8s configuration files.
type Configuration struct {
	// Version is the semantic version of the configuration file format.
	Version string `yaml:"version"`

	// AddonRepositories is extra addon repositories to configure on the local node.
	AddonRepositories []AddonRepositoryConfiguration `yaml:"addonRepositories"`

	// Addons is a list of addons to enable and/or disable.
	Addons []AddonConfiguration `yaml:"addons"`

	// ExtraKubeletArgs is a list of extra arguments to add to the local node kubelet.
	// Set a value to null to remove it from the arguments.
	ExtraKubeletArgs map[string]*string `yaml:"extraKubeletArgs"`

	// ExtraKubeAPIServerArgs is a list of extra arguments to add to the local node kube-apiserver.
	// Set a value to null to remove it from the arguments.
	ExtraKubeAPIServerArgs map[string]*string `yaml:"extraKubeAPIServerArgs"`

	// ExtraKubeProxyArgs is a list of extra arguments to add to the local node kube-proxy.
	// Set a value to null to remove it from the arguments.
	ExtraKubeProxyArgs map[string]*string `yaml:"extraKubeProxyArgs"`

	// ExtraKubeControllerManagerArgs is a list of extra arguments to add to the local node kube-controller-manager.
	// Set a value to null to remove it from the arguments.
	ExtraKubeControllerManagerArgs map[string]*string `yaml:"extraKubeControllerManagerArgs"`

	// ExtraKubeSchedulerArgs is a list of extra arguments to add to the local node kube-scheduler.
	// Set a value to null to remove it from the arguments.
	ExtraKubeSchedulerArgs map[string]*string `yaml:"extraKubeSchedulerArgs"`

	// ExtraKubeliteEnv is extra environment variables (e.g. GOFIPS) for the local node Kubernetes services.
	ExtraKubeliteEnv map[string]*string `yaml:"extraKubeliteEnv"`

	// ExtraSANs are a list of extra Subject Alternate Names to add to the local API server.
	ExtraSANs *[]string `yaml:"extraSANs"`

	// ContainerdRegistryConfigs is containerd hosts.toml configurations to configure registries.
	ContainerdRegistryConfigs map[string]string `yaml:"containerdRegistryConfigs"`

	// ExtraContainerdArgs is a list of extra arguments to add to the local node containerd.
	// Set a value to null to remove it from the arguments.
	ExtraContainerdArgs map[string]*string `yaml:"extraContainerdArgs"`

	// ExtraContainerdEnv is extra environment variables (e.g. proxy configuration) for the local node containerd.
	// Set a value to null to remove it from the environment.
	ExtraContainerdEnv map[string]*string `yaml:"extraContainerdEnv"`

	// ExtraDqliteArgs is a list of extra arguments to add to the local node Dqlite.
	// Set a value to null to remove it from the arguments.
	ExtraDqliteArgs map[string]*string `yaml:"extraDqliteArgs"`

	// ExtraDqliteEnv is extra environment variables (e.g. dqlite debug flags) for the local node dqlite.
	// Set a value to null to remove it from the environment.
	ExtraDqliteEnv map[string]*string `yaml:"extraDqliteEnv"`

	// ExtraMicroK8sClusterAgentArgs is a list of extra arguments to add to the local node cluster-agent.
	// Set a value to null to remove it from the arguments.
	ExtraMicroK8sClusterAgentArgs map[string]*string `yaml:"extraMicroK8sClusterAgentArgs"`

	// ExtraMicroK8sClusterAgentEnv is extra environment variables (e.g. GOFIPS) for the local node cluster-agent.
	// Set a value to null to remove it from the environment.
	ExtraMicroK8sClusterAgentEnv map[string]*string `yaml:"extraMicroK8sClusterAgentEnv"`

	// ExtraMicroK8sAPIServerProxyArgs is a list of extra arguments (e.g. --refresh-interval) to add to the local node apiserver-proxy used by worker nodes.
	// Set a value to null to remove it from the arguments.
	ExtraMicroK8sAPIServerProxyArgs map[string]*string `yaml:"extraMicroK8sAPIServerProxyArgs"`

	// ExtraMicroK8sAPIServerProxyEnv is extra environment variables (e.g. GOFIPS) for the local node apiserver-proxy.
	// Set a value to null to remove it from the environment.
	ExtraMicroK8sAPIServerProxyEnv map[string]*string `yaml:"extraMicroK8sAPIServerProxyEnv"`

	// ExtraEtcdArgs is a list of extra arguments to add to the local node etcd.
	// Set a value to null to remove it from the arguments.
	ExtraEtcdArgs map[string]*string `yaml:"extraEtcdArgs"`

	// ExtraEtcdEnv is extra environment variables (e.g. GOFIPS) for the local node etcd.
	// Set a value to null to remove it from the environment.
	ExtraEtcdEnv map[string]*string `yaml:"extraEtcdEnv"`

	// ExtraFlanneldArgs is a list of extra arguments to add to the local node flanneld.
	// Set a value to null to remove it from the arguments.
	ExtraFlanneldArgs map[string]*string `yaml:"extraFlanneldArgs"`

	// ExtraFlanneldEnv is extra environment variables (e.g. GOFIPS) for the local node flanneld.
	// Set a value to null to remove it from the environment.
	ExtraFlanneldEnv map[string]*string `yaml:"extraFlanneldEnv"`

	// ExtraConfigFiles is extra service configuration files to create (e.g. for configuring kube-apiserver encryption at rest).
	// These files will be written at $SNAP_DATA/args/<filename>.
	ExtraConfigFiles map[string]string `yaml:"extraConfigFiles"`

	// PersistentClusterToken is a token that may be used to authentication join requests to the local node.
	PersistentClusterToken string `yaml:"persistentClusterToken"`

	// Join configuration. Setting this will attempt to join the local node to an already existing MicroK8s cluster.
	Join JoinConfiguration `yaml:"join"`

	// ExtraCNIEnv is configuration of network such us IPv4/v6 cluster and service CIDRs.
	ExtraCNIEnv map[string]*string `yaml:"extraCNIEnv"`

	// ExtraFIPSEnv is configuration for MicroK8s to run in FIPS mode.
	ExtraFIPSEnv map[string]*string `yaml:"extraFIPSEnv"`
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

	if c.isZero() {
		return nil, errEmptyConfig
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

// ParseMultiPartConfiguration parses a multiple YAML configuration objects into a MultiPartConfiguration.
func ParseMultiPartConfiguration(b []byte) (MultiPartConfiguration, error) {
	reader := k8syaml.NewYAMLReader(bufio.NewReader(bytes.NewBuffer(b)))

	cfg := MultiPartConfiguration{}
	for {
		doc, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else if err != nil {
				return MultiPartConfiguration{}, err
			}
		}

		part, err := ParseConfiguration(doc)
		if err != nil {
			if errors.Is(err, errEmptyConfig) {
				continue
			}
			return MultiPartConfiguration{}, err
		}
		cfg.Parts = append(cfg.Parts, part)
	}

	return cfg, nil
}

// isZero returns true if all configuration values are zero/empty.
// NOTE(neoaggelos): this needs to be updated when new fields are added to the Configuration struct.
func (c *Configuration) isZero() bool {
	switch {
	case c.Version != "":
		return false
	case c.PersistentClusterToken != "":
		return false
	case c.Join.URL != "":
		return false
	case c.Join.Worker:
		return false
	case len(c.AddonRepositories) > 0:
		return false
	case len(c.Addons) > 0:
		return false
	case len(c.ExtraKubeletArgs) > 0:
		return false
	case len(c.ExtraKubeAPIServerArgs) > 0:
		return false
	case len(c.ExtraKubeProxyArgs) > 0:
		return false
	case len(c.ExtraKubeControllerManagerArgs) > 0:
		return false
	case len(c.ExtraKubeSchedulerArgs) > 0:
		return false
	case len(c.ExtraKubeliteEnv) > 0:
		return false
	case c.ExtraSANs != nil && len(*c.ExtraSANs) > 0:
		return false
	case len(c.ContainerdRegistryConfigs) > 0:
		return false
	case len(c.ExtraContainerdArgs) > 0:
		return false
	case len(c.ExtraContainerdEnv) > 0:
		return false
	case len(c.ExtraDqliteArgs) > 0:
		return false
	case len(c.ExtraDqliteEnv) > 0:
		return false
	case len(c.ExtraMicroK8sClusterAgentArgs) > 0:
		return false
	case len(c.ExtraMicroK8sClusterAgentEnv) > 0:
		return false
	case len(c.ExtraMicroK8sAPIServerProxyArgs) > 0:
		return false
	case len(c.ExtraMicroK8sAPIServerProxyEnv) > 0:
		return false
	case len(c.ExtraEtcdArgs) > 0:
		return false
	case len(c.ExtraEtcdEnv) > 0:
		return false
	case len(c.ExtraFlanneldArgs) > 0:
		return false
	case len(c.ExtraFlanneldEnv) > 0:
		return false
	case len(c.ExtraConfigFiles) > 0:
		return false
	case len(c.ExtraCNIEnv) > 0:
		return false
	case len(c.ExtraFIPSEnv) > 0:
		return false
	}
	return true
}
