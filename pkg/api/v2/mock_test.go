package v2_test

import (
	"context"
	"io"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

func mockListControlPlaneNodes(nodes ...string) func(context.Context, snap.Snap) ([]string, error) {
	return func(context.Context, snap.Snap) ([]string, error) {
		return nodes, nil
	}
}

// recordingReader wraps an io.Reader and records Read function calls.
type recordingReader struct {
	io.Reader
	CountReadBytes []int
}

func (r *recordingReader) Read(p []byte) (int, error) {
	if r.CountReadBytes == nil {
		r.CountReadBytes = make([]int, 0, 1)
	}
	n, err := r.Reader.Read(p)
	r.CountReadBytes = append(r.CountReadBytes, n)
	return n, err
}

var _ io.Reader = &recordingReader{}
