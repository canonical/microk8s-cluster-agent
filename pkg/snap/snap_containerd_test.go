package snap_test

import (
	"os"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	. "github.com/onsi/gomega"
)

func TestUpdateContainerdRegistryConfigs(t *testing.T) {
	if err := os.MkdirAll("testdata/args", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %s", err)
	}
	defer os.RemoveAll("testdata/args")

	s := snap.NewSnap("testdata", "testdata")

	t.Run("Mirror", func(t *testing.T) {
		g := NewWithT(t)
		err := s.UpdateContainerdRegistryConfigs(map[string][]byte{
			"docker.io": []byte(`server = "http://dockerhub.mirror:32000"`),
			"quay.io":   []byte(`server = "http://quay.mirror:32000"`),
		})
		g.Expect(err).To(BeNil())

		b, err := os.ReadFile("testdata/args/certs.d/docker.io/hosts.toml")
		g.Expect(err).To(BeNil())
		g.Expect(b).To(Equal([]byte(`server = "http://dockerhub.mirror:32000"`)))

		b, err = os.ReadFile("testdata/args/certs.d/quay.io/hosts.toml")
		g.Expect(err).To(BeNil())
		g.Expect(b).To(Equal([]byte(`server = "http://quay.mirror:32000"`)))
	})

	t.Run("BadPath", func(t *testing.T) {
		g := NewWithT(t)
		err := s.UpdateContainerdRegistryConfigs(map[string][]byte{
			"../path/traversal": []byte(`server = "http://dockerhub.mirror:32000"`),
		})
		g.Expect(err).NotTo(BeNil())
	})
}
