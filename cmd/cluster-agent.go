package cmd

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	v1 "github.com/canonical/microk8s-cluster-agent/pkg/api/v1"
	v2 "github.com/canonical/microk8s-cluster-agent/pkg/api/v2"
	"github.com/canonical/microk8s-cluster-agent/pkg/k8sinit"
	"github.com/canonical/microk8s-cluster-agent/pkg/server"
	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
	"github.com/spf13/cobra"
)

var (
	bind                         string
	keyfile                      string
	certfile                     string
	timeout                      int
	enableMetrics                bool
	launchConfigurationsEnable   bool
	launchConfigurationsInterval time.Duration
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

		// Setup launch configuration handler
		if launchConfigurationsEnable {
			go func() {
				ctx := cmd.Context()
				log.Printf("Starting watch for launch configurations")

				if launchConfigurationsInterval < 5*time.Second {
					log.Printf("Launch configurations interval %v is less than minimum of 5s. Using the minimum 5s instead.\n", launchConfigurationsInterval)
					launchConfigurationsInterval = 5 * time.Second
				}
				refreshCh := time.NewTicker(launchConfigurationsInterval).C

			nextTick:
				for {
					select {
					case <-ctx.Done():
						return
					case <-refreshCh:
					}

					files, err := filepath.Glob(filepath.Join(os.Getenv("SNAP_COMMON"), "etc", "launcher", "*.yaml"))
					if err != nil {
						log.Printf("Failed to search for launch configuration files: %v", err)
						continue nextTick
					}

				nextFile:
					for _, file := range files {
						log.Printf("Applying %s", file)
						b, err := os.ReadFile(file)
						if err != nil {
							log.Printf("Failed to read launch configuration file %s: %v", file, err)
							continue nextFile
						}
						cfg, err := k8sinit.ParseMultiPartConfiguration(b)
						if err != nil {
							log.Printf("Failed to parse configuration file %s: %v", file, err)
							continue nextFile
						}
						launcher := k8sinit.NewLauncher(s, false)
						if err := launcher.Apply(ctx, cfg); err != nil {
							log.Printf("Failed to apply configuration file %s: %v", file, err)
							continue nextFile
						}
						if err := os.Rename(file, file+".applied"); err != nil {
							log.Printf("Failed to rename applied configuration file %s: %v", file, err)
						}
						log.Printf("Successfully applied %s", file)
					}
				}
			}()
		}

		// Setup HTTP server
		apiv1 := &v1.API{
			Snap:     s,
			LookupIP: net.LookupIP,
		}
		apiv2 := &v2.API{
			Snap:                    s,
			LookupIP:                net.LookupIP,
			InterfaceAddrs:          net.InterfaceAddrs,
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
	clusterAgentCmd.Flags().BoolVar(&launchConfigurationsEnable, "launch-configurations-enable", true, "Enable launch configurations")
	clusterAgentCmd.Flags().DurationVar(&launchConfigurationsInterval, "launch-configurations-interval", 5*time.Second, "Interval between checks for launch configurations")

	rootCmd.AddCommand(clusterAgentCmd)
}
