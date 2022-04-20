package cmd

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	v1 "github.com/canonical/microk8s-cluster-agent/pkg/api/v1"
	v2 "github.com/canonical/microk8s-cluster-agent/pkg/api/v2"
	v3 "github.com/canonical/microk8s-cluster-agent/pkg/api/v3"
	"github.com/canonical/microk8s-cluster-agent/pkg/server"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
	"github.com/spf13/cobra"
)

var (
	bind          string
	keyfile       string
	certfile      string
	timeout       int
	enableMetrics bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cluster-agent",
	Short: "MicroK8s cluster agent",
	Long: `The MicroK8s cluster agent is an API server that orchestrates the
lifecycle of a MicroK8s cluster.`,
	Run: func(cmd *cobra.Command, args []string) {

		s := snap.NewSnap(os.Getenv("SNAP"), os.Getenv("SNAP_DATA"), util.RunCommand)
		apiv1 := &v1.API{
			Snap:     s,
			LookupIP: net.LookupIP,
		}
		apiv2 := &v2.API{
			Snap:                    s,
			LookupIP:                net.LookupIP,
			ListControlPlaneNodeIPs: snaputil.ListControlPlaneNodeIPs,
		}
		apiv3 := &v3.API{
			Snap:                    s,
			LookupIP:                net.LookupIP,
			ListControlPlaneNodeIPs: snaputil.ListControlPlaneNodeIPs,
		}
		srv := server.NewServer(time.Duration(timeout)*time.Second, enableMetrics, apiv1.RegisterServer, apiv2.RegisterServer, apiv3.RegisterServer)
		log.Printf("Starting cluster agent on https://%s\n", bind)
		if err := http.ListenAndServeTLS(bind, certfile, keyfile, srv); err != nil {
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
	rootCmd.Flags().BoolVar(&enableMetrics, "enable-metrics", false, "Enable metrics endpoint")
}
