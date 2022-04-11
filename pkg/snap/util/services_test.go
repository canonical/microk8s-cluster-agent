package snaputil_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
)

func TestGetServiceArgument(t *testing.T) {
	serviceOneArguments := `
--key=value
--key-with-space value2
   --key-with-padding=value3
--multiple=keys --in-the-same-row=this-is-lost
`
	serviceTwoArguments := `
--key=value-of-service-two
`
	s := &mock.Snap{
		ServiceArguments: map[string]string{
			"service":  serviceOneArguments,
			"service2": serviceTwoArguments,
		},
	}
	if err := os.MkdirAll("testdata/args", 0755); err != nil {
		t.Fatal("Failed to setup test directory")
	}
	for _, tc := range []struct {
		service       string
		key           string
		expectedValue string
	}{
		{service: "service", key: "--key", expectedValue: "value"},
		{service: "service2", key: "--key", expectedValue: "value-of-service-two"},
		{service: "service", key: "--key-with-padding", expectedValue: "value3"},
		{service: "service", key: "--key-with-space", expectedValue: "value2"},
		{service: "service", key: "--missing", expectedValue: ""},
		{service: "service3", key: "--missing-service", expectedValue: ""},
		// NOTE: the final test case documents that arguments in the same row will not be parsed properly.
		// This is carried over from the original Python code, and probably needs fixing in the future.
		{service: "service", key: "--in-the-same-row", expectedValue: ""},
	} {
		t.Run(fmt.Sprintf("%s/%s", tc.service, tc.key), func(t *testing.T) {
			if v := snaputil.GetServiceArgument(s, tc.service, tc.key); v != tc.expectedValue {
				t.Fatalf("Expected argument value to be %q, but it was %q instead", tc.expectedValue, v)
			}
		})
	}
}

func TestUpdateServiceArguments(t *testing.T) {
	initialArguments := `
--key=value
--other=other-value
--with-space value2
`
	for _, tc := range []struct {
		name           string
		update         []map[string]string
		delete         []string
		expectedValues map[string]string
	}{
		{
			name:   "simple-update",
			update: []map[string]string{{"--key": "new-value"}},
			delete: []string{},
			expectedValues: map[string]string{
				"--key":   "new-value",
				"--other": "other-value",
			},
		},
		{
			name:   "update-many-delete-one",
			update: []map[string]string{{"--key": "new-value"}, {"--other": "other-new-value"}},
			delete: []string{"--with-space"},
			expectedValues: map[string]string{
				"--key":        "new-value",
				"--other":      "other-new-value",
				"--with-space": "",
			},
		},
		{
			name:   "update-many-single-list",
			update: []map[string]string{{"--key": "new-value", "--other": "other-new-value"}},
			expectedValues: map[string]string{
				"--key":   "new-value",
				"--other": "other-new-value",
			},
		},
		{
			name: "no-updates",
			expectedValues: map[string]string{
				"--key": "value",
			},
		},
		{
			name:   "new-opt",
			update: []map[string]string{{"--new-opt": "opt-value"}},
			expectedValues: map[string]string{
				"--new-opt": "opt-value",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &mock.Snap{
				ServiceArguments: map[string]string{
					"service": initialArguments,
				},
			}

			if err := snaputil.UpdateServiceArguments(s, "service", tc.update, tc.delete); err != nil {
				t.Fatalf("Expected no error updating arguments file but received %q", err)
			}
			for key, expectedValue := range tc.expectedValues {
				if value := snaputil.GetServiceArgument(s, "service", key); value != expectedValue {
					t.Fatalf("Expected value for argument %q does not match (%q and %q)", key, value, expectedValue)
				}
			}
		})
	}
}
