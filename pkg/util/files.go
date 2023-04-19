package util

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
)

// FileExists returns true if the specified path exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// SetupPermissions attempts to set file permissions to 0660 and group to `microk8s` for a given file.
// SetupPermissions will knowingly ignore any errors, as failing to update permissions will only occur
// in extraordinary situations, and will never break the MicroK8s cluster.
func SetupPermissions(path string, chownGroup string) {
	os.Chmod(path, 0660)
	if group, err := user.LookupGroup(chownGroup); err == nil {
		if gid, err := strconv.ParseInt(group.Gid, 10, 32); err == nil {
			os.Chown(path, -1, int(gid))
		}
	}
}

// ReadFile returns the file contents as a string.
func ReadFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", path, err)
	}
	return string(b), nil
}

// ParseArgumentLine parses a command-line argument from a single line.
// The returned key includes any dash prefixes.
func ParseArgumentLine(line string) (key string, value string) {
	line = strings.TrimSpace(line)

	// parse "--argument value" and "--argument=value" variants
	if parts := strings.Split(line, "="); len(parts) >= 2 {
		key = parts[0]
		value = parts[1]
	} else if parts := strings.Split(line, " "); len(parts) >= 2 {
		key = parts[0]
		value = strings.Join(parts[1:], " ")
	} else {
		key = line
	}

	return
}
