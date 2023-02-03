package util_test

import (
	"bytes"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/util"
)

func TestGenerateCSRConf(t *testing.T) {
	b, err := util.GenerateCSRConf([]string{"10.10.10.10", "microk8s.example.com"})
	if err != nil {
		t.Fatalf("did not expect an error but got %q", err)
	}

	if found := bytes.Contains(b, []byte("\nIP.2 = 10.10.10.10\n")); !found {
		t.Fatalf("expected IP address 10.10.10.10 SAN but it was not there")
	}
	if found := bytes.Contains(b, []byte("\nDNS.5 = microk8s.example.com\n")); !found {
		t.Fatalf("expected DNS name microk8s.example.com SAN but it was not there")
	}
}
