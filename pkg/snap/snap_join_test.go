package snap_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	utiltest "github.com/canonical/microk8s-cluster-agent/pkg/util/test"
	. "github.com/onsi/gomega"
)

func TestJoinCluster(t *testing.T) {
	t.Run("PropagateError", func(t *testing.T) {
		g := NewWithT(t)
		runner := &utiltest.MockRunner{}
		s := snap.NewSnap("testdata", "testdata", snap.WithCommandRunner(runner.Run))
		runner.Err = fmt.Errorf("some error")

		err := s.JoinCluster(context.Background(), "some-url", false)
		g.Expect(err).ToNot(BeNil())
		g.Expect(errors.Is(err, runner.Err)).To(BeTrue())
	})

	t.Run("ControlPlane", func(t *testing.T) {
		g := NewWithT(t)
		runner := &utiltest.MockRunner{}
		s := snap.NewSnap("testdata", "testdata", snap.WithCommandRunner(runner.Run))

		err := s.JoinCluster(context.Background(), "10.10.10.10:25000/token/hash", false)
		g.Expect(err).To(BeNil())
		g.Expect(runner.CalledWithCommand).To(ConsistOf("testdata/microk8s-join.wrapper 10.10.10.10:25000/token/hash"))
	})

	t.Run("Worker", func(t *testing.T) {
		g := NewWithT(t)
		runner := &utiltest.MockRunner{}
		s := snap.NewSnap("testdata", "testdata", snap.WithCommandRunner(runner.Run))

		err := s.JoinCluster(context.Background(), "10.10.10.10:25000/token/hash", true)
		g.Expect(err).To(BeNil())
		g.Expect(runner.CalledWithCommand).To(ConsistOf("testdata/microk8s-join.wrapper 10.10.10.10:25000/token/hash --worker"))
	})
}
