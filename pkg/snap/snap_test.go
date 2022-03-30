package snap_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	utiltest "github.com/canonical/microk8s-cluster-agent/pkg/util/test"
)

func TestServiceRestart(t *testing.T) {
	mockRunner := &utiltest.MockRunner{}
	s := snap.NewSnap("testdata", "testdata", mockRunner.Run)

	t.Run("NoKubelite", func(t *testing.T) {
		for _, tc := range []struct {
			service         string
			expectedCommand string
		}{
			{service: "apiserver", expectedCommand: "snapctl restart microk8s.daemon-apiserver"},
			{service: "proxy", expectedCommand: "snapctl restart microk8s.daemon-proxy"},
			{service: "kubelet", expectedCommand: "snapctl restart microk8s.daemon-kubelet"},
			{service: "scheduler", expectedCommand: "snapctl restart microk8s.daemon-scheduler"},
			{service: "controller-manager", expectedCommand: "snapctl restart microk8s.daemon-controller-manager"},
			{service: "kube-apiserver", expectedCommand: "snapctl restart microk8s.daemon-apiserver"},
			{service: "kube-proxy", expectedCommand: "snapctl restart microk8s.daemon-proxy"},
			{service: "kube-scheduler", expectedCommand: "snapctl restart microk8s.daemon-scheduler"},
			{service: "kube-controller-manager", expectedCommand: "snapctl restart microk8s.daemon-controller-manager"},
			{service: "k8s-dqlite", expectedCommand: "snapctl restart microk8s.daemon-k8s-dqlite"},
			{service: "cluster-agent", expectedCommand: "snapctl restart microk8s.daemon-cluster-agent"},
			{service: "containerd", expectedCommand: "snapctl restart microk8s.daemon-containerd"},
			{service: "microk8s.daemon-containerd", expectedCommand: "snapctl restart microk8s.daemon-containerd"},
		} {
			t.Run(tc.service, func(t *testing.T) {
				s.RestartService(context.Background(), tc.service)
				if lastCmd := mockRunner.CalledWithCommand[len(mockRunner.CalledWithCommand)-1]; lastCmd != tc.expectedCommand {
					t.Fatalf("Expected command %q, but %q was called instead", tc.expectedCommand, lastCmd)
				}
			})
		}
	})

	t.Run("Kubelite", func(t *testing.T) {
		if err := os.MkdirAll("testdata/var/lock", 0755); err != nil {
			t.Fatalf("Failed to create test directory: %s", err)
		}
		defer os.RemoveAll("testdata/var")
		if _, err := os.Create("testdata/var/lock/lite.lock"); err != nil {
			t.Fatalf("Failed to create kubelite lock file: %s", err)
		}
		for _, tc := range []struct {
			service         string
			expectedCommand string
		}{
			{service: "apiserver", expectedCommand: "snapctl restart microk8s.daemon-kubelite"},
			{service: "proxy", expectedCommand: "snapctl restart microk8s.daemon-kubelite"},
			{service: "kubelet", expectedCommand: "snapctl restart microk8s.daemon-kubelite"},
			{service: "scheduler", expectedCommand: "snapctl restart microk8s.daemon-kubelite"},
			{service: "controller-manager", expectedCommand: "snapctl restart microk8s.daemon-kubelite"},
			{service: "k8s-dqlite", expectedCommand: "snapctl restart microk8s.daemon-k8s-dqlite"},
			{service: "cluster-agent", expectedCommand: "snapctl restart microk8s.daemon-cluster-agent"},
			{service: "containerd", expectedCommand: "snapctl restart microk8s.daemon-containerd"},
		} {
			t.Run(tc.service, func(t *testing.T) {
				s.RestartService(context.Background(), tc.service)
				if lastCmd := mockRunner.CalledWithCommand[len(mockRunner.CalledWithCommand)-1]; lastCmd != tc.expectedCommand {
					t.Fatalf("Expected command %q, but %q was called instead", tc.expectedCommand, lastCmd)
				}
			})
		}
	})
}

