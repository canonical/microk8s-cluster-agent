package snaputil

import (
	"fmt"
	"strings"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

// GetServiceArgument retrieves the value of a specific argument from the $SNAP_DATA/args/$service file.
// The argument name should include preceding dashes (e.g. "--secure-port").
// If any errors occur, or the argument is not present, an empty string is returned.
func GetServiceArgument(s snap.Snap, serviceName string, argument string) string {
	arguments, err := s.ReadServiceArguments(serviceName)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(arguments, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, argument) {
			continue
		}
		// parse "--argument value" and "--argument=value" variants
		line = line[strings.LastIndex(line, " ")+1:]
		line = line[strings.LastIndex(line, "=")+1:]
		return line
	}
	return ""
}

// UpdateServiceArguments updates the arguments file for a service.
// UpdateServiceArguments is a no-op if updateList and delete are empty.
// updateList is a map of key-value pairs. It will replace the argument with the new value (or just append).
// delete is a list of arguments to remove completely. The argument is removed if present.
func UpdateServiceArguments(s snap.Snap, serviceName string, updateList []map[string]string, delete []string) error {
	if updateList == nil {
		updateList = []map[string]string{}
	}
	if delete == nil {
		delete = []string{}
	}

	// If no updates are requested, exit early
	if len(updateList) == 0 && len(delete) == 0 {
		return nil
	}

	deleteMap := make(map[string]struct{}, len(delete))
	for _, k := range delete {
		deleteMap[k] = struct{}{}
	}

	updateMap := make(map[string]string, len(updateList))
	for _, update := range updateList {
		for key, value := range update {
			updateMap[key] = value
		}
	}

	arguments, err := s.ReadServiceArguments(serviceName)
	if err != nil {
		return fmt.Errorf("failed to read arguments of service %s: %w", serviceName, err)
	}

	newArguments := make([]string, 0, len(arguments))
	for _, line := range strings.Split(arguments, "\n") {
		line = strings.TrimSpace(line)
		// ignore empty lines
		if line == "" {
			continue
		}
		// handle "--argument value" and "--argument=value" variants
		key := strings.SplitN(line, " ", 2)[0]
		key = strings.SplitN(key, "=", 2)[0]
		if newValue, ok := updateMap[key]; ok {
			// update argument with new value
			newArguments = append(newArguments, fmt.Sprintf("%s=%s", key, newValue))
		} else if _, ok := deleteMap[key]; ok {
			// remove argument
			continue
		} else {
			// no change
			newArguments = append(newArguments, line)
		}
	}

	if err := s.WriteServiceArguments(serviceName, []byte(strings.Join(newArguments, "\n")+"\n")); err != nil {
		return fmt.Errorf("failed to update arguments for service %s: %q", serviceName, err)
	}
	return nil
}
