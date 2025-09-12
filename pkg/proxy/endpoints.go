package proxy

//lint:file-ignore SA1019 Ignore usage of deprecated k8s api

import (
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
)

func parseEndpoints(endpoint *v1.Endpoints) []string {
	if endpoint == nil {
		return nil
	}
	addresses := make([]string, 0, len(endpoint.Subsets))
	for _, subset := range endpoint.Subsets {
		portNumber := 16443
		for _, port := range subset.Ports {
			if port.Name == "https" {
				portNumber = int(port.Port)
				break
			}
		}

		for _, addr := range subset.Addresses {
			addresses = append(addresses, fmt.Sprintf("%s:%d", addr.IP, portNumber))
		}
	}

	sort.Strings(addresses)
	return addresses
}
