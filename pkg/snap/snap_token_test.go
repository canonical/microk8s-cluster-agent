package snap_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
)

func TestClusterTokens(t *testing.T) {
	os.RemoveAll("testdata/credentials")
	s := snap.NewSnap("testdata", "testdata", nil)
	t.Run("MissingTokensFile", func(t *testing.T) {
		if s.ConsumeClusterToken("token1") {
			t.Fatal("Expected token1 to not be valid, but it is")
		}
	})
	if err := os.MkdirAll("testdata/credentials", 0755); err != nil {
		t.Fatal("Failed to create test directory")
	}
	defer os.RemoveAll("testdata/credentials")
	now := time.Now().Unix()
	clusterTokens := fmt.Sprintf(`
one-time-token
token-invalid-timestamp|-10a
token-expired|%d
token-not-expired|%d
`, now-300, now+300)

	if err := os.WriteFile("testdata/credentials/cluster-tokens.txt", []byte(clusterTokens), 0600); err != nil {
		t.Fatalf("Failed to create test cluster-tokens.txt file: %s", err)
	}

	for _, tc := range []struct {
		token         string
		expectedValid bool
	}{
		// empty token is never valid
		{token: "", expectedValid: false},

		// missing token and invalid timestamp is not valid
		{token: "missing-token", expectedValid: false},
		{token: "token-invalid-timestamp", expectedValid: false},

		// tokens with expired TTL are not valid
		{token: "token-expired", expectedValid: false},

		// one-time tokens are only valid for a single use
		{token: "one-time-token", expectedValid: true},
		{token: "one-time-token", expectedValid: false},
		{token: "one-time-token", expectedValid: false},

		// tokens with a TTL may be reused
		{token: "token-not-expired", expectedValid: true},
		{token: "token-not-expired", expectedValid: true},
		{token: "token-not-expired", expectedValid: true},
	} {
		t.Run(tc.token, func(t *testing.T) {
			if s.ConsumeClusterToken(tc.token) != tc.expectedValid {
				if tc.expectedValid {
					t.Fatalf("Token %s should be valid, but it is not", tc.token)
				} else {
					t.Fatalf("Token %s should not be valid, but it is", tc.token)
				}
			}
		})
	}
}

func TestCertificateRequestTokens(t *testing.T) {
	if err := os.MkdirAll("testdata/credentials", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %s", err)
	}
	defer os.RemoveAll("testdata/credentials")
	s := snap.NewSnap("testdata", "testdata", nil)
	if err := s.AddCertificateRequestToken("my-token"); err != nil {
		t.Fatalf("Failed to add certificate request token: %s", err)
	}
	contents, err := util.ReadFile("testdata/credentials/certs-request-tokens.txt")
	if err != nil {
		t.Fatalf("Failed to retrieve tokens: %s", err)
	}
	if !strings.Contains(contents, "my-token\n") {
		t.Fatal("Expected my-token to exist in tokens file, but it did not")
	}
	if !s.ConsumeCertificateRequestToken("my-token") {
		t.Fatal("Expected my-token to be a valid certificate request token, but it is not")
	}
}

func TestCallbackTokens(t *testing.T) {
	if err := os.MkdirAll("testdata/credentials", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %s", err)
	}
	defer os.RemoveAll("testdata/credentials")
	s := snap.NewSnap("testdata", "testdata", nil)
	if err := s.AddCallbackToken("ip:port", "my-token"); err != nil {
		t.Fatalf("Failed to add certificate request token: %s", err)
	}
	contents, err := util.ReadFile("testdata/credentials/callback-tokens.txt")
	if err != nil {
		t.Fatalf("Failed to retrieve tokens: %s", err)
	}
	if !strings.Contains(contents, "ip:port my-token\n") {
		t.Fatal("Expected my-token to exist in tokens file, but it did not")
	}
}

func TestSelfCallbackToken(t *testing.T) {
	if err := os.MkdirAll("testdata/credentials", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %s", err)
	}
	defer os.RemoveAll("testdata/credentials")
	s := snap.NewSnap("testdata", "testdata", nil)
	token, err := s.GetOrCreateSelfCallbackToken()
	if err != nil {
		t.Fatalf("Failed to configure callback token: %q", err)
	}
	if token == "" {
		t.Fatalf("Expected token to not be empty, but it is")
	}
	if !s.ConsumeSelfCallbackToken(token) {
		t.Fatal("Expected my-token to be a valid callback token for this node, but it is not")
	}
	tokenAgain, err := s.GetOrCreateSelfCallbackToken()
	if err != nil {
		t.Fatalf("Failed to retrieve callback token: %q", err)
	}
	if tokenAgain != token {
		t.Fatalf("Expected tokens to match, but they do not (%q and %q)", token, tokenAgain)
	}
}

func TestKnownTokens(t *testing.T) {
	if err := os.MkdirAll("testdata/credentials", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %s", err)
	}
	defer os.RemoveAll("testdata/credentials")
	s := snap.NewSnap("testdata", "testdata", nil)
	if token, err := s.GetKnownToken("user"); token != "" || err == nil {
		t.Fatalf("Expected an empty token and an error, but found token %s and error %s", token, err)
	}

	contents := `
admin-token,admin,admin,"system:masters"
token1,system:kube-proxy,kube-proxy
token2,system:node:existing-host,kubelet-0123,"system:nodes"
`
	if err := os.WriteFile("testdata/credentials/known_tokens.csv", []byte(contents), 0600); err != nil {
		t.Fatalf("Failed to create file with known tokens: %s", err)
	}
	for _, tc := range []struct {
		user        string
		expectToken string
		expectError bool
	}{
		{user: "missing-user", expectError: true},
		{user: "system:kube-proxy", expectToken: "token1"},
		{user: "system:node:existing-host", expectToken: "token2"},
		{user: "admin", expectToken: "admin-token"},
	} {
		t.Run(tc.user, func(t *testing.T) {
			token, err := s.GetKnownToken(tc.user)
			switch {
			case tc.expectError && err == nil:
				t.Fatal("Expected an error but did not get one")
			case !tc.expectError && err != nil:
				t.Fatalf("Expected no errors, but received %q", err)
			case tc.expectToken != token:
				t.Fatalf("Expected token %q but received %q", tc.expectToken, token)
			}
		})
	}
	t.Run("Kubelet", func(t *testing.T) {
		t.Run("Existing", func(t *testing.T) {
			token, err := s.GetOrCreateKubeletToken("existing-host")
			if err != nil {
				t.Fatalf("Expected no errors, but received %q", err)
			}
			if token != "token2" {
				t.Fatalf("Expected token %q, but received %q", "token2", token)
			}
		})
		t.Run("Create", func(t *testing.T) {
			newToken, err := s.GetOrCreateKubeletToken("new-host")
			if err != nil {
				t.Fatalf("Expected no errors, but received %q", err)
			}
			if newToken == "" {
				t.Fatal("Expected token to be not-empty, but it was")
			}
			token, err := s.GetOrCreateKubeletToken("new-host")
			if err != nil {
				t.Fatalf("Expected no errors, but received %q", err)
			}
			if token != newToken {
				t.Fatalf("Expected tokens to match, but they do not (%q != %q)", token, newToken)
			}
		})
	})
}

func TestStrictGroup(t *testing.T) {
	if err := os.MkdirAll("testdata/meta", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %s", err)
	}
	defer os.RemoveAll("testdata/meta")
	for _, tc := range []struct {
		confinement string
		group       string
	}{
		{confinement: "strict", group: "snap_microk8s"},
		{confinement: "classic", group: "microk8s"},
	} {
		if err := os.WriteFile("testdata/meta/snapcraft.yaml", []byte(fmt.Sprintf("confinement: %s", tc.confinement)), 0660); err != nil {
			t.Fatalf("Failed to create test file: %s", err)
		}
		group := snap.NewSnap("testdata", "testdata", nil).GetGroupName()
		if tc.group != group {
			t.Fatalf("Expected group to be %q but it was %q instead", tc.group, group)
		}
	}
}
