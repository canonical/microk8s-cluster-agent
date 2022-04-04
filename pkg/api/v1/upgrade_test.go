package v1_test

import (
	"context"
	"reflect"
	"testing"

	v1 "github.com/canonical/microk8s-cluster-agent/pkg/api/v1"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
)

func TestUpgrade(t *testing.T) {
	s := &mock.Snap{
		SelfCallbackTokens: []string{"valid-token"},
	}
	apiv1 := &v1.API{Snap: s}

	for _, tc := range []struct {
		token       string
		expectErr   bool
		expectCalls []string
	}{
		{token: "valid-token", expectErr: false, expectCalls: []string{"upgrade phase"}},
		{token: "invalid-token", expectErr: true},
	} {
		t.Run(tc.token, func(t *testing.T) {
			s.RunUpgradeCalledWith = nil
			req := v1.UpgradeRequest{
				CallbackToken: tc.token,
				UpgradeName:   "upgrade",
				UpgradePhase:  "phase",
			}

			err := apiv1.Upgrade(context.Background(), req)
			switch {
			case tc.expectErr && err == nil:
				t.Fatalf("expected an error but did not receive any")
			case !tc.expectErr && err != nil:
				t.Fatalf("expected no errors but received %q", err)
			}

			if !reflect.DeepEqual(tc.expectCalls, s.RunUpgradeCalledWith) {
				t.Fatalf("expected calls %#v but received %#v", tc.expectCalls, s.RunUpgradeCalledWith)
			}
		})
	}
}
