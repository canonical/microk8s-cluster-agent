package snaputil

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

// MaybePatchCalicoAutoDetectionMethod attempts to update the calico cni.yaml to
// update the value of the IP_AUTODETECTION_METHOD option.
//
// The default value is "first-found", which works well for single-node clusters.
// However, after adding a new node, we want to change this to `can-reach=canReachHost`
// to mitigate issues with multiple NICs.
//
// Optionally, the new manifest may be applied using the microk8s-kubectl.wrapper script.
func MaybePatchCalicoAutoDetectionMethod(ctx context.Context, s snap.Snap, canReachHost string, apply bool) error {
	config, err := s.ReadCNIYaml()
	var newConfig string
	var after, target, replacement string

	if err != nil {
		return fmt.Errorf("failed to read existing cni configuration: %w", err)
	}

	if ip := net.ParseIP(canReachHost); ip.To4() == nil {
		// Address is in IPv6
		after, target, replacement = "IP6_AUTODETECTION_METHOD", "first-found", "can-reach=%s"
	} else {
		// Address is in IPv4
		after, target, replacement = "IP_AUTODETECTION_METHOD", "first-found", "can-reach=%s"
	}

	newConfig, err = ReplaceAfter(config, after, target, fmt.Sprintf(replacement, canReachHost))
	if err != nil {
		return fmt.Errorf("failed to update %s in CNI configuration: %w", after, err)
	}

	if newConfig != config {
		if err := s.WriteCNIYaml([]byte(newConfig)); err != nil {
			return fmt.Errorf("failed to update cni configuration: %w", err)
		}
	}
	if apply {
		if err := s.ApplyCNI(ctx); err != nil {
			return fmt.Errorf("failed to apply cni configuration: %w", err)
		}
	}
	return nil
}

// ReplaceAfter replaces the first occurrence of target after the first occurrence of "after" with replacement.
func ReplaceAfter(s, after, target, replacement string) (string, error) {
	if !strings.Contains(s, after) {
		return "", fmt.Errorf("after string not found: '%s'", after)
	}

	afterIndex := strings.Index(s, after)
	if afterIndex == -1 {
		return "", fmt.Errorf("string not found: '%s'", after)
	}

	targetIndex := strings.Index(s[afterIndex:], target)
	if targetIndex == -1 {
		return "", fmt.Errorf("string not found after '%s': '%s'", after, target)
	}

	endIndex := afterIndex + targetIndex

	result := s[:endIndex] + replacement + s[len(target)+endIndex:]
	return result, nil
}
