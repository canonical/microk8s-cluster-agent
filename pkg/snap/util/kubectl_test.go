package snaputil

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestParseNodeInternalIPs(t *testing.T) {
	nodeList := []v1.Node{
		{
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{Type: v1.NodeHostName, Address: "host1"},
					{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				},
			},
		},
		{
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{Type: v1.NodeHostName, Address: "host2"},
					{Type: v1.NodeInternalIP, Address: "10.0.0.2"},
				},
			},
		},
	}
	expectedIPs := []string{"10.0.0.1", "10.0.0.2"}

	ips := parseNodeInternalIPs(nodeList)
	if !reflect.DeepEqual(ips, expectedIPs) {
		t.Fatalf("expected list of nodes to be %v but it was %v instead", expectedIPs, ips)
	}
}
