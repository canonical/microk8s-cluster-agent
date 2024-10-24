package internal

import (
	"os"
	"sync"
	"testing"
	"time"
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

func TestWriteFile(t *testing.T) {
	file, err := os.CreateTemp("", "test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	const numWriters = 100
	const numIterations = 100

	var wg sync.WaitGroup
	wg.Add(numWriters)

	// The data to write to the file
	testData := map[string]interface{}{
		"key": "value",
	}

	for i := 0; i < numWriters; i++ {
		go func(writerID int) {
			defer wg.Done()

			for j := 0; j < numIterations; j++ {
				if err := writeYaml(file.Name(), testData); err != nil {
					t.Errorf("Writer %d failed to write yaml (iteration %d): %v", writerID, j, err)
				}

				content, err := os.ReadFile(file.Name())
				if err != nil {
					t.Errorf("Writer %d failed to read file (iteration %d): %v", writerID, j, err)
				}

				if string(content) != "key: value\n" {
					t.Errorf("Writer %d: Invalid content (iteration %d): %s", writerID, j, string(content))
				}

				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
}
