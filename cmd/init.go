package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/canonical/microk8s-cluster-agent/pkg/k8sinit"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"github.com/spf13/cobra"
)

var (
	initInputFile = ""

	initCmd = &cobra.Command{
		Use:    "init",
		Short:  "Apply MicroK8s configurations",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := snap.NewSnap(
				os.Getenv("SNAP"),
				os.Getenv("SNAP_DATA"),
			)
			l := k8sinit.NewLauncher(s)

			var (
				b   []byte
				err error
			)
			switch initInputFile {
			case "":
				return fmt.Errorf("no config file specified")
			case "-":
				b, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read config from stdin: %w", err)
				}
			default:
				b, err = os.ReadFile(initInputFile)
				if err != nil {
					return fmt.Errorf("failed to read config file %q: %w", initInputFile, err)
				}
			}

			c, err := k8sinit.ParseMultiPartConfiguration(b)
			if err != nil {
				return fmt.Errorf("failed to parse config file: %w", err)
			}

			if err := l.Apply(cmd.Context(), c); err != nil {
				return fmt.Errorf("failed to apply configuration: %w", err)
			}
			return nil
		},
	}
)

func init() {
	initCmd.Flags().StringVarP(&initInputFile, "config-file", "c", initInputFile, "configuration file to read, or '-' to read from stdin")

	rootCmd.AddCommand(initCmd)
}
