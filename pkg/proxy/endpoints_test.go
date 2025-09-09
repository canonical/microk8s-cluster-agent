package proxy

import (
	"reflect"
	"testing"

	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestParseEndpoints(t *testing.T) {
	httpsPortName := "https"
	for _, tc := range []struct {
		name           string
		endpointSlices *discoveryv1.EndpointSliceList
		addresses      []string
	}{
		{
			name: "one",
			endpointSlices: &discoveryv1.EndpointSliceList{
				Items: []discoveryv1.EndpointSlice{
					{
						ObjectMeta:  metav1.ObjectMeta{Name: "test-1"},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{"1.1.1.1"}},
						},
						Ports: []discoveryv1.EndpointPort{
							{Name: &httpsPortName, Port: ptr.To(int32(16643))},
						},
					},
				},
			},
			addresses: []string{"1.1.1.1:16643"},
		},
		{
			name: "two",
			endpointSlices: &discoveryv1.EndpointSliceList{
				Items: []discoveryv1.EndpointSlice{
					{
						ObjectMeta:  metav1.ObjectMeta{Name: "test-2"},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{"1.1.1.1"}},
							{Addresses: []string{"2.2.2.2"}},
						},
						Ports: []discoveryv1.EndpointPort{
							{Name: &httpsPortName, Port: ptr.To(int32(16643))},
						},
					},
				},
			},
			addresses: []string{"1.1.1.1:16643", "2.2.2.2:16643"},
		},
		{
			name: "IPv6",
			endpointSlices: &discoveryv1.EndpointSliceList{
				Items: []discoveryv1.EndpointSlice{
					{
						ObjectMeta:  metav1.ObjectMeta{Name: "test-ipv6"},
						AddressType: discoveryv1.AddressTypeIPv6,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{"fe80::e0b9:bfff:fe90:8d37"}},
						},
						Ports: []discoveryv1.EndpointPort{
							{Name: &httpsPortName, Port: ptr.To(int32(16643))},
						},
					},
				},
			},
			addresses: []string{"[fe80::e0b9:bfff:fe90:8d37]:16643"},
		},
		{
			name: "multiple-slices",
			endpointSlices: &discoveryv1.EndpointSliceList{
				Items: []discoveryv1.EndpointSlice{
					{
						ObjectMeta:  metav1.ObjectMeta{Name: "test-multi-1"},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{"1.1.1.1"}},
							{Addresses: []string{"2.2.2.2"}},
						},
						Ports: []discoveryv1.EndpointPort{
							{Name: &httpsPortName, Port: ptr.To(int32(16643))},
						},
					},
					{
						ObjectMeta:  metav1.ObjectMeta{Name: "test-multi-2"},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{"3.3.3.3"}},
						},
						Ports: []discoveryv1.EndpointPort{
							{Name: &httpsPortName, Port: ptr.To(int32(16643))},
						},
					},
				},
			},
			addresses: []string{"1.1.1.1:16643", "2.2.2.2:16643", "3.3.3.3:16643"},
		},
		{
			name: "override-port",
			endpointSlices: &discoveryv1.EndpointSliceList{
				Items: []discoveryv1.EndpointSlice{
					{
						ObjectMeta:  metav1.ObjectMeta{Name: "test-port-1"},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{"1.1.1.1"}},
							{Addresses: []string{"2.2.2.2"}},
						},
						Ports: []discoveryv1.EndpointPort{
							{Name: &httpsPortName, Port: ptr.To(int32(16643))},
						},
					},
					{
						ObjectMeta:  metav1.ObjectMeta{Name: "test-port-2"},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{"3.3.3.3"}},
						},
						Ports: []discoveryv1.EndpointPort{
							{Name: &httpsPortName, Port: ptr.To(int32(10000))},
						},
					},
				},
			},
			addresses: []string{"1.1.1.1:16643", "2.2.2.2:16643", "3.3.3.3:10000"},
		},
		{
			name: "sort",
			endpointSlices: &discoveryv1.EndpointSliceList{
				Items: []discoveryv1.EndpointSlice{
					{
						ObjectMeta:  metav1.ObjectMeta{Name: "test-sort-1"},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{"3.3.3.3"}},
							{Addresses: []string{"1.1.1.1"}},
						},
						Ports: []discoveryv1.EndpointPort{
							{Name: &httpsPortName, Port: ptr.To(int32(16643))},
						},
					},
					{
						ObjectMeta:  metav1.ObjectMeta{Name: "test-sort-2"},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{"2.2.2.2"}},
						},
						Ports: []discoveryv1.EndpointPort{
							{Name: &httpsPortName, Port: ptr.To(int32(10000))},
						},
					},
				},
			},
			addresses: []string{"1.1.1.1:16643", "2.2.2.2:10000", "3.3.3.3:16643"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if parsed := parseAddresses(tc.endpointSlices); !reflect.DeepEqual(parsed, tc.addresses) {
				t.Fatalf("expected addresses to be %v but they were %v instead", tc.addresses, parsed)
			}
		})
	}
}
