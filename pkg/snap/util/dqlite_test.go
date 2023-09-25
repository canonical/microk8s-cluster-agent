package snaputil_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
	. "github.com/onsi/gomega"
)

func TestUpdateDqliteIP(t *testing.T) {
	s := &mock.Snap{
		DqliteClusterYaml: `
- Address: 127.0.0.1:19001
  ID: 1236189235178654365
  Role: 0`,
		DqliteInfoYaml: `
Address: 127.0.0.1:19001
ID: 1236189235178654365
Role: 0`,
	}

	if err := snaputil.UpdateDqliteIP(context.Background(), s, "10.10.10.10"); err != nil {
		t.Fatalf("Expected no errors but received %q", err)
	}
	expectedCalledWith := []string{"Address: 10.10.10.10:19001\n"}
	if !reflect.DeepEqual(s.WriteDqliteUpdateYamlCalledWith, expectedCalledWith) {
		t.Fatalf("Expected WriteDqliteUpdateYaml to be called with %q, but it was called with %q instead", expectedCalledWith, s.WriteDqliteUpdateYamlCalledWith)
	}
}

func TestWaitForDqliteCluster(t *testing.T) {
	s := &mock.Snap{
		DqliteClusterYaml: `
- Address: 127.0.0.1:19001
  ID: 1236189235178654365
  Role: 0`,
	}

	t.Run("Cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := snaputil.WaitForDqliteCluster(ctx, s, func(snaputil.DqliteCluster) (bool, error) { return false, nil })
		if err == nil {
			t.Fatalf("Expected an error but did not receive any")
		}
	})

	t.Run("OK", func(t *testing.T) {

		s := &mock.Snap{
			DqliteClusterYaml: `
- Address: 127.0.0.1:19001
  ID: 1236189235178654365
  Role: 0`,
			DqliteInfoYaml: `
Address: 127.0.0.1:19001
ID: 1236189235178654365
Role: 0`,
		}

		// update cluster yaml asynchronously
		go func() {
			<-time.After(500 * time.Millisecond)
			s.DqliteClusterYaml = `
- Address: 10.10.10.10:19001
  ID: 1236189235178654365
  Role: 0
`
		}()

		cluster, err := snaputil.WaitForDqliteCluster(context.Background(), s, func(cluster snaputil.DqliteCluster) (bool, error) {
			return len(cluster) == 1 && cluster[0].Address == "10.10.10.10:19001", nil
		})
		if err != nil {
			t.Fatalf("Expected no errors but received: %q", err)
		}
		if cluster[0].Address != "10.10.10.10:19001" {
			t.Fatalf("Expected cluster to contain update node information, but received %#v", cluster)
		}
	})

}

func TestUpdateDqliteBindAddress(t *testing.T) {
	findMatchingBindAddressMock := func(addr string) (string, error) {
		return "10.10.10.10", nil
	}

	t.Run("MustFailIfNodeAlreadyKnown", func(t *testing.T) {

		s := &mock.Snap{
			DqliteClusterYaml: `
- Address: 127.0.0.1:19001
  ID: 1236189235178654365
  Role: 0`,
			DqliteInfoYaml: `
Address: 127.0.0.1:19001
ID: 1236189235178654365
Role: 0`,
		}

		g := NewWithT(t)
		err := snaputil.MaybeUpdateDqliteBindAddress(context.Background(), s, "127.0.0.1", "127.0.0.1", findMatchingBindAddressMock)
		g.Expect(err).To(MatchError("the joining node (127.0.0.1) is already known to dqlite"))

	})

	t.Run("MustNotUpdateIfMultipleNodes", func(t *testing.T) {
		s := &mock.Snap{
			DqliteClusterYaml: `
- Address: 127.0.0.1:19001
  ID: 1236189235178654365
  Role: 0
- Address: 10.10.10.10:19001
  ID: 12345678987654321
  Role: 0`,
			DqliteInfoYaml: `
Address: 127.0.0.1:19001
ID: 1236189235178654365
Role: 0`,
		}

		g := NewWithT(t)
		err := snaputil.MaybeUpdateDqliteBindAddress(context.Background(), s, "127.0.0.1:19001", "8.8.8.8", findMatchingBindAddressMock)
		g.Expect(err).To(BeNil())
		g.Expect(s.WriteDqliteUpdateYamlCalledWith).To(BeEmpty())
	})

	t.Run("MustUpdateIfOneNodeAndBindsLocal", func(t *testing.T) {
		s := &mock.Snap{
			DqliteClusterYaml: `
- Address: 127.0.0.1:19001
  ID: 1236189235178654365
  Role: 0`,
			DqliteInfoYaml: `
Address: 127.0.0.1:19001
ID: 1236189235178654365
Role: 0`,
		}
		// Update cluster yaml asynchronously
		// The mock snap implementation doesn't actually write the yaml but instead
		// sets `WriteDqliteUpdateYamlCalledWith`.
		// Thus, we need to manually change the config as otherwise `WaitForDqliteCluster`
		// would wait forever.
		go func() {
			<-time.After(500 * time.Millisecond)
			s.DqliteClusterYaml = `
- Address: 10.10.10.10:19001
  ID: 1236189235178654365
  Role: 0`
		}()

		g := NewWithT(t)
		err := snaputil.MaybeUpdateDqliteBindAddress(context.Background(), s, "127.0.0.1:19001", "10.10.10.10", findMatchingBindAddressMock)
		g.Expect(err).To(BeNil())
		g.Expect(s.WriteDqliteUpdateYamlCalledWith).To(ConsistOf("Address: 10.10.10.10:19001\n"))
	})
}
