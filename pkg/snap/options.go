package snap

import (
	"context"
	"time"
)

// WithRetryApplyCNI configures how many times the ApplyCNI operation is retries before giving up.
func WithRetryApplyCNI(times int, backoff time.Duration) func(s *snap) {
	return func(s *snap) {
		s.applyCNIRetries = times
		if s.applyCNIRetries <= 0 {
			s.applyCNIRetries = 1
		}
		s.applyCNIBackoff = backoff
	}
}

// WithCommandRunner configures how shell commands are executed.
func WithCommandRunner(f func(context.Context, ...string) error) func(s *snap) {
	return func(s *snap) {
		s.runCommand = f
	}
}

// WithCAPIPath configures the path to the CAPI directory.
func WithCAPIPath(path string) func(s *snap) {
	return func(s *snap) {
		s.capiPath = path
	}
}
