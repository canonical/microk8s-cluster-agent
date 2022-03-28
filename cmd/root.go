package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	v1 "github.com/canonical/microk8s-cluster-agent/pkg/api/v1"
	v2 "github.com/canonical/microk8s-cluster-agent/pkg/api/v2"
	"github.com/canonical/microk8s-cluster-agent/pkg/server"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
	utiltest "github.com/canonical/microk8s-cluster-agent/pkg/util/test"
	"github.com/spf13/cobra"
)

var (
	bind          string
	keyfile       string
	certfile      string
	timeout       int
	devMode       bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cluster-agent",
	Short: "MicroK8s cluster agent",
	Long: `The MicroK8s cluster agent is an API server that orchestrates the
lifecycle of a MicroK8s cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiv1 := &v1.API{}
		apiv2 := &v2.API{
			ListControlPlaneNodeIPs: util.ListControlPlaneNodeIPs,
		}

		// devMode is used for debugging
		if devMode {
			log.Println("Running in development mode")
			util.SnapData = "data"
			util.Snap = "data"
			util.CommandRunner = (&utiltest.MockRunner{Log: true}).Run
			apiv2.ListControlPlaneNodeIPs = func(context.Context) ([]string, error) {
				return []string{"10.0.0.1", "10.0.0.2"}, nil
			}
		}

		s := server.NewServer(time.Duration(timeout)*time.Second, apiv1, apiv2)
		log.Printf("Starting cluster agent on https://%s\n", bind)
		if err := http.ListenAndServeTLS(bind, certfile, keyfile, s); err != nil {
			log.Fatalf("Failed to listen: %s", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&bind, "bind", "0.0.0.0:25000", "Listen address for server")
	rootCmd.Flags().StringVar(&keyfile, "keyfile", "", "Private key for serving TLS")
	rootCmd.Flags().StringVar(&certfile, "certfile", "", "Certificate for serving TLS")
	rootCmd.Flags().IntVar(&timeout, "timeout", 240, "Default request timeout (in seconds)")

	// TODO: remove
	rootCmd.Flags().BoolVar(&devMode, "devmode", false, "Turn on development mode (local data, mock commands)")
}
