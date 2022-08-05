package internal

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"
)

const (
	testTraefikYaml = `
entryPoints:
  apiserver:
    address: ":16443"
providers:
  file:
    filename: testdata/provider.yaml
    watch: true`

	testProviderYaml = `
tcp:
  routers:
    Router-1:
      rule: "HostSNI(` + "`*`" + `)"
      service: "kube-apiserver"
      tls:
        passthrough: true
  services:
    kube-apiserver:
      loadBalancer:
        servers:
          - address: 10.0.0.1:16443
          - address: 10.0.0.2:16443
          - address: 10.0.0.3:16443`
)

func TestConfiguration(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := os.MkdirAll("testdata", 0755); err != nil {
		t.Fatal("Failed to setup testdata directory")
	}
	defer os.RemoveAll("testdata")
	for fname, contents := range map[string]string{
		"testdata/traefik.yaml":  testTraefikYaml,
		"testdata/provider.yaml": testProviderYaml,
	} {
		if err := os.WriteFile(fname, []byte(contents), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %q", fname, err)
		}
	}

	t.Run("Load", func(t *testing.T) {
		cfg, err := LoadConfiguration(ctx, "testdata/traefik.yaml")
		if err != nil {
			t.Fatalf("Expected no errors but received %q", err)
		}
		expectedAPIServers := []string{"10.0.0.1:16443", "10.0.0.2:16443", "10.0.0.3:16443"}
		if !reflect.DeepEqual(expectedAPIServers, cfg.Endpoints) {
			t.Fatalf("Expected endpoints to be %v but they were %v instead", expectedAPIServers, cfg.Endpoints)
		}
	})

	t.Run("Update", func(t *testing.T) {
		endpoints := []string{"10.0.0.4:16443", "10.0.0.5:16443"}

		if err := UpdateConfiguration(endpoints, "testdata/traefik.yaml"); err != nil {
			t.Fatalf("Expected no errors updating the configuration file but received %q instead", err)
		}

		newCfg, err := LoadConfiguration(ctx, "testdata/traefik.yaml")
		if err != nil {
			t.Fatalf("Expected no errors but received %q", err)
		}
		expectedAPIServers := []string{"10.0.0.4:16443", "10.0.0.5:16443"}
		if !reflect.DeepEqual(expectedAPIServers, newCfg.Endpoints) {
			t.Fatalf("Expected endpoints to be %v but they were %v instead", expectedAPIServers, newCfg.Endpoints)
		}
	})

	t.Run("ConfigChanged", func(t *testing.T) {
		cfg, err := LoadConfiguration(ctx, "testdata/traefik.yaml")
		if err != nil {
			t.Fatalf("Expected no errors but received %q", err)
		}
		if err := os.Remove("testdata/provider.yaml"); err != nil {
			t.Fatalf("Expected no errors when removing file but received %v instead", err)
		}
		select {
		case <-cfg.ChangedCh:
		case <-time.After(time.Second):
			t.Fatal("Timed out waiting for config changed event")
		}
	})

}
