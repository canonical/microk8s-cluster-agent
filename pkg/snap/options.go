package snap

import (
	"context"
	"time"
)


// WithCommandRunner configures how shell commands are executed.
func WithCommandRunner(f func(context.Context, ...string) error) func(s *snap) {
	return func(s *snap) {
		s.runCommand = f
	}
}
