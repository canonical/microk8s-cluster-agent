package snap_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

func TestCAPIAuthToken(t *testing.T) {
	capiTestPath := "./capi-test"
	os.RemoveAll(capiTestPath)
	s := snap.NewSnap("", "", "", capiTestPath)
	token := "token123"

	g := NewWithT(t)

	isValid, err := s.IsCAPIAuthTokenValid(token)
	g.Expect(err).To(MatchError(os.ErrNotExist))
	g.Expect(isValid).To(BeFalse())

	capiEtc := filepath.Join(capiTestPath, "etc")
	defer os.RemoveAll(capiTestPath)
	g.Expect(os.MkdirAll(capiEtc, 0755)).To(Succeed())
	g.Expect(os.WriteFile("./capi-test/etc/token", []byte(token), 0600)).To(Succeed())

	isValid, err = s.IsCAPIAuthTokenValid("random-token")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(isValid).To(BeFalse())

	isValid, err = s.IsCAPIAuthTokenValid(token)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(isValid).To(BeTrue())
}
