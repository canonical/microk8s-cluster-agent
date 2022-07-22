package cmd

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	v1 "github.com/canonical/microk8s-cluster-agent/pkg/api/v1"
	v2 "github.com/canonical/microk8s-cluster-agent/pkg/api/v2"
	"github.com/canonical/microk8s-cluster-agent/pkg/server"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
	"github.com/spf13/cobra"
)

var (
	bind          string
	keyfile       string
	certfile      string
	timeout       int
	enableMetrics bool
)

// clusterAgentCmd represents the base command when called without any subcommands
var clusterAgentCmd = &cobra.Command{
	Use:   "cluster-agent",
	Short: "MicroK8s cluster agent",
	Long: `The MicroK8s cluster agent is an API server that orchestrates the
lifecycle of a MicroK8s cluster.`,
	Run: func(cmd *cobra.Command, args []string) {

		s := snap.NewSnap(
			os.Getenv("SNAP"),
			os.Getenv("SNAP_DATA"),
			snap.WithRetryApplyCNI(20, 3*time.Second),
		)
		apiv1 := &v1.API{
			Snap:     s,
			LookupIP: net.LookupIP,
		}
		apiv2 := &v2.API{
			Snap:                    s,
			LookupIP:                net.LookupIP,
			ListControlPlaneNodeIPs: snaputil.ListControlPlaneNodeIPs,
		}
		srv := server.NewServer(time.Duration(timeout)*time.Second, enableMetrics, apiv1, apiv2)
		log.Printf("Starting cluster agent on https://%s\n", bind)
		if err := http.ListenAndServeTLS(bind, certfile, keyfile, srv); err != nil {
			log.Fatalf("Failed to listen: %s", err)
		}
	},
}

func init() {
	clusterAgentCmd.Flags().StringVar(&bind, "bind", "0.0.0.0:25000", "Listen address for server")
	clusterAgentCmd.Flags().StringVar(&keyfile, "keyfile", "", "Private key for serving TLS")
	clusterAgentCmd.Flags().StringVar(&certfile, "certfile", "", "Certificate for serving TLS")
	clusterAgentCmd.Flags().IntVar(&timeout, "timeout", 240, "Default request timeout (in seconds)")
	clusterAgentCmd.Flags().BoolVar(&enableMetrics, "enable-metrics", false, "Enable metrics endpoint")

	rootCmd.AddCommand(clusterAgentCmd)
}
