package snaputil

import (
	"context"
	"fmt"
	"net"
	"regexp"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

var (
	// https://regex101.com/r/VHTsvE/1
	ipAutodetectionMethodRe  = regexp.MustCompile(`(?m)(IP_AUTODETECTION_METHOD(.*\n.*)?)first-found`)
	ip6AutodetectionMethodRe = regexp.MustCompile(`(?m)(IP6_AUTODETECTION_METHOD(.*\n.*)?)first-found`)
)

// MaybePatchCalicoAutoDetectionMethod attempts to update the calico cni.yaml to
// update the value of the IP_AUTODETECTION_METHOD option.
//
// The default value is "first-found", which works well for single-node clusters.
// However, after adding a new node, we want to change this to `can-reach=canReachHost`
// to mitigate issues with multiple NICs.
//
// If canReachHost is an IPv6 address, IP6_AUTODETECTION_METHOD is updated instead.
//
// Optionally, the new manifest may be applied using the microk8s-kubectl.wrapper script.
func MaybePatchCalicoAutoDetectionMethod(ctx context.Context, s snap.Snap, canReachHost string, apply bool) error {
	config, err := s.ReadCNIYaml()
	if err != nil {
		return fmt.Errorf("failed to read existing cni configuration: %w", err)
	}

	re := ipAutodetectionMethodRe
	if ip := net.ParseIP(canReachHost); ip != nil && ip.To4() == nil {
		// Address is in IPv6
		re = ip6AutodetectionMethodRe
	}

	newConfig := re.ReplaceAllString(config, fmt.Sprintf("${1}can-reach=%s", canReachHost))
	if newConfig == config {
		return nil
	}
	if err := s.WriteCNIYaml([]byte(newConfig)); err != nil {
		return fmt.Errorf("failed to update cni configuration: %w", err)
	}
	if apply {
		if err := s.ApplyCNI(ctx); err != nil {
			return fmt.Errorf("failed to apply cni configuration: %w", err)
		}
	}
	return nil
}
