package internal

import (
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
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
	t.Run("EmptyData", func(t *testing.T) {
		g := NewWithT(t)

		file, err := os.CreateTemp("", "test-*.yaml")
		defer os.Remove(file.Name())

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(writeYaml(file.Name(), nil)).ToNot(Succeed())
	})

	t.Run("ValidData", func(t *testing.T) {
		g := NewWithT(t)

		file, err := os.CreateTemp("", "test-*.yaml")
		defer os.Remove(file.Name())

		g.Expect(err).ToNot(HaveOccurred())

		const (
			numWriters    = 100
			numIterations = 100
		)

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
					g.Expect(writeYaml(file.Name(), testData)).To(Succeed())

					content, err := os.ReadFile(file.Name())
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(string(content)).To(Equal("key: value\n"))

					time.Sleep(10 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()
	})
}
