package v1_test

import (
	"context"
	"reflect"
	"testing"

	v1 "github.com/canonical/microk8s-cluster-agent/pkg/api/v1"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
)

func TestConfigure(t *testing.T) {
	s := &mock.Snap{
		SelfCallbackTokens: []string{"valid-token"},
		ServiceArguments: map[string]string{
			"kube-apiserver": "--key=value\n--old=to-remove",
			"kube-proxy":     "--key=value2",
		},
	}
	apiv1 := &v1.API{Snap: s}
	t.Run("InvalidToken", func(t *testing.T) {
		err := apiv1.Configure(context.Background(), v1.ConfigureRequest{
			CallbackToken: "invalid-token",
			ConfigureServices: []v1.ConfigureServiceRequest{
				{Name: "kube-apiserver", Restart: true},
			},
		})
		if err == nil {
			t.Fatal("Expected an error but did not receive any")
		}
		if s.WriteServiceArgumentsCalled {
			t.Fatalf("Expected WriteServiceArguments to not be called, but it was")
		}
	})

	for _, tc := range []struct {
		name              string
		req               v1.ConfigureRequest
		expectedRestart   []string
		expectedEnable    []string
		expectedDisable   []string
		expectedArguments map[string]map[string]string
	}{
		{
			name: "update-services-add-addons",
			req: v1.ConfigureRequest{
				CallbackToken: "valid-token",
				ConfigureServices: []v1.ConfigureServiceRequest{
					{Name: "kube-apiserver", UpdateArguments: []map[string]string{{"--key": "new-value"}}, RemoveArguments: []string{"--old"}},
					{Name: "kube-proxy", Restart: true},
				},
				ConfigureAddons: []v1.ConfigureAddonRequest{
					{Name: "dns", Enable: true},
					{Name: "ingress", Disable: true},
					{Name: "other"},
				},
			},
			expectedRestart: []string{"kube-proxy"},
			expectedEnable:  []string{"dns"},
			expectedDisable: []string{"ingress"},
			expectedArguments: map[string]map[string]string{
				"kube-apiserver": {
					"--key": "new-value",
					"--old": "",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := apiv1.Configure(context.Background(), tc.req); err != nil {
				t.Fatalf("Expected no errors but received %q", err)
			}
			for serviceName, expectedArguments := range tc.expectedArguments {
				for key, expectedValue := range expectedArguments {
					if value := snaputil.GetServiceArgument(s, serviceName, key); value != expectedValue {
						t.Fatalf("Expected argument %q of service %q to be %q, but it is %q", key, serviceName, expectedValue, value)
					}
				}
			}

			if !reflect.DeepEqual(tc.expectedDisable, s.DisableAddonCalledWith) {
				t.Fatalf("Expected disable addons %#v but received %#v", tc.expectedDisable, s.DisableAddonCalledWith)
			}
			if !reflect.DeepEqual(tc.expectedEnable, s.EnableAddonCalledWith) {
				t.Fatalf("Expected enable addons %#v but received %#v", tc.expectedEnable, s.EnableAddonCalledWith)
			}
			if !reflect.DeepEqual(tc.expectedRestart, s.RestartServiceCalledWith) {
				t.Fatalf("Expected restart services %#v but received %#v", tc.expectedRestart, s.RestartServiceCalledWith)
			}
		})
	}
}
