package snaputil

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"gopkg.in/yaml.v2"
)

// DqliteCluster is the format of the dqlite cluster.yaml file.
type DqliteCluster []DqliteClusterNode

// DqliteClusterNode is a node in the dqlite cluster.
type DqliteClusterNode struct {
	// Address is the address of the node in the cluster.
	Address string `yaml:"Address"`
	// ID is the unique identifier of the node in the cluster.
	ID uint64 `yaml:"ID,omitempty"`
	// NodeRole is the role of the node in the cluster.
	// 0 -- Voter
	// 1 -- StandBy
	// 2 -- Spare
	NodeRole int `yaml:"Role,omitempty"`
}

// GetDqliteCluster a list of all currently known dqlite cluster nodes.
func GetDqliteCluster(s snap.Snap) (DqliteCluster, error) {
	clusterYaml, err := s.ReadDqliteClusterYaml()
	if err != nil {
		return DqliteCluster{}, fmt.Errorf("failed to read list of dqlite nodes: %w", err)
	}

	cluster := DqliteCluster{}
	if err := yaml.Unmarshal([]byte(clusterYaml), &cluster); err != nil {
		return DqliteCluster{}, fmt.Errorf("failed to parse list of dqlite nodes: %w", err)
	}

	return cluster, nil
}

// UpdateDqliteIP sets the local dqlite cluster node to bind to a new IP address.
func UpdateDqliteIP(ctx context.Context, s snap.Snap, host string) error {
	infoYaml, err := s.ReadDqliteInfoYaml()
	if err != nil {
		return fmt.Errorf("failed to retrieve current node info: %w", err)
	}
	var node DqliteClusterNode
	if err := yaml.Unmarshal([]byte(infoYaml), &node); err != nil {
		return fmt.Errorf("invalid format for current node info: %w", err)
	}

	_, port, _ := net.SplitHostPort(node.Address)
	nodeUpdate := DqliteClusterNode{
		Address: net.JoinHostPort(host, port),
	}
	b, err := yaml.Marshal(nodeUpdate)
	if err != nil {
		return fmt.Errorf("failed to marshal current node info update: %w", err)
	}

	if err := s.WriteDqliteUpdateYaml(b); err != nil {
		return fmt.Errorf("failed to create dqlite update file: %w", err)
	}

	if err := s.RestartService(ctx, "k8s-dqlite"); err != nil {
		return fmt.Errorf("failed to restart k8s-dqlite service: %w", err)
	}
	return nil
}

// WaitForDqliteCluster queries the dqlite cluster nodes repeatedly until f(cluster) becomes true.
func WaitForDqliteCluster(ctx context.Context, s snap.Snap, f func(DqliteCluster) (bool, error)) (DqliteCluster, error) {
	interval := time.NewTicker(time.Second)
	for {
		cluster, err := GetDqliteCluster(s)
		if err != nil {
			return DqliteCluster{}, err
		}

		ok, err := f(cluster)
		if err != nil {
			return DqliteCluster{}, fmt.Errorf("failed check for cluster condition: %w", err)
		}
		if ok {
			return cluster, nil
		}

		select {
		case <-ctx.Done():
			return DqliteCluster{}, fmt.Errorf("timed out waiting for cluster condition: %w", ctx.Err())
		case <-interval.C:
		}
	}
}

// RemoveNodeFromDqlite uses the Dqlite binary to remove a node from the Dqlite cluster.
func RemoveNodeFromDqlite(ctx context.Context, snap snap.Snap, removeEp string) error {
	binPath := snap.GetSnapPath("bin", "dqlite")
	clusterYamlPath := snap.GetSnapDataPath("var", "kubernetes", "backend", "cluster.yaml")
	clusterCrtPath := snap.GetSnapDataPath("var", "kubernetes", "backend", "cluster.crt")
	clusterKeyPath := snap.GetSnapDataPath("var", "kubernetes", "backend", "cluster.key")

	// NOTE(Hue): The last two arguments (.remove <address>) should be a single string. Otherwise Dqlite throws an error.
	if err := snap.RunCommand(ctx, binPath, "-s", "file://"+clusterYamlPath, "-c", clusterCrtPath, "-k", clusterKeyPath, "-f", "json", "k8s", fmt.Sprintf(".remove %s", removeEp)); err != nil {
		return fmt.Errorf("failed to run remove command: %w", err)
	}

	return nil
}
