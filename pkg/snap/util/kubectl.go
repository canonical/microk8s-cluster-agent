package snaputil

import (
	"context"
	"fmt"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func parseNodeInternalIPs(nodeList []v1.Node) []string {
	addresses := make([]string, 0, len(nodeList))
	for _, node := range nodeList {
		for _, address := range node.Status.Addresses {
			if address.Type == v1.NodeInternalIP {
				addresses = append(addresses, address.Address)
			}
		}
	}

	return addresses
}

// ListControlPlaneNodeIPs returns the internal IPs of the control plane nodes of the MicroK8s cluster.
func ListControlPlaneNodeIPs(ctx context.Context, s snap.Snap) ([]string, error) {
	config, err := clientcmd.BuildConfigFromFlags("", s.GetKubeconfigFile())
	if err != nil {
		return nil, fmt.Errorf("failed to read load kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kubernetes client: %w", err)
	}

	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: "node.kubernetes.io/microk8s-controlplane=microk8s-controlplane",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return parseNodeInternalIPs(nodes.Items), nil
}
