package proxy

import (
	"context"
	"fmt"
	"sort"

	discoveryv1 "k8s.io/api/discovery/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func parseAddresses(endpointSlices *discoveryv1.EndpointSliceList) []string {
	if endpointSlices == nil {
		return nil
	}

	addresses := make([]string, 0, len(endpointSlices.Items))

	for _, endpointSlice := range endpointSlices.Items {
		portNumber := 16443
		for _, port := range endpointSlice.Ports {
			if port.Name != nil && *port.Name == "https" {
				if port.Port != nil {
					portNumber = int(*port.Port)
					break
				}
			}
		}

		for _, endpoint := range endpointSlice.Endpoints {
			for _, addr := range endpoint.Addresses {
				if addr != "" {
					var address string
					switch endpointSlice.AddressType {
					case discoveryv1.AddressTypeIPv4:
						address = addr
					case discoveryv1.AddressTypeIPv6:
						address = fmt.Sprintf("[%s]", addr)
					case discoveryv1.AddressTypeFQDN:
						// Not supported, skip.
						continue
					default:
						// Unknown address type, skip.
						continue
					}
					addresses = append(addresses, fmt.Sprintf("%s:%d", address, portNumber))
				}
			}
		}

	}

	sort.Strings(addresses)
	return addresses
}

func getKubernetesEndpoints(ctx context.Context, kubeconfigFile string) ([]string, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read load kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kubernetes client: %w", err)
	}

	endpointSlices, err := clientset.DiscoveryV1().EndpointSlices("default").List(ctx, metav1.ListOptions{
		LabelSelector: "kubernetes.io/service-name=kubernetes",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve endpoints for kubernetes service: %w", err)
	}
	return parseAddresses(endpointSlices), nil
}
