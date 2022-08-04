package internal

import (
	"os"
	"testing"
)

func TestGetDefaultProviderFile(t *testing.T) {
	for _, tc := range []struct {
		name         string
		snapData     string
		expectedPath string
	}{
		{
			name:         "NoSnapData",
			snapData:     "",
			expectedPath: "/var/snap/microk8s/current/args/traefik/provider.yaml",
		},
		{
			name:         "SnapDataCurrent",
			snapData:     "/var/snap/microk8s/current",
			expectedPath: "/var/snap/microk8s/current/args/traefik/provider.yaml",
		},
		{
			name:         "SnapDataRevisionNumber",
			snapData:     "/var/snap/microk8s2/1010",
			expectedPath: "/var/snap/microk8s2/current/args/traefik/provider.yaml",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("SNAP_DATA", tc.snapData)
			if v := getDefaultProviderFile(); v != tc.expectedPath {
				t.Fatalf("Expected provider file to be %q but it was %q instead", tc.expectedPath, v)
			}
		})
	}
}
