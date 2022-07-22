package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/canonical/microk8s-cluster-agent/pkg/proxy"
	"github.com/spf13/cobra"
)

var (
	apiServerProxyConfig          string
	apiServerProxyKubeconfig      string
	apiServerProxyRefreshInterval time.Duration

	apiServerProxyCmd = &cobra.Command{
		Use: "apiserver-proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := &proxy.APIServerProxy{
				ConfigFile:     apiServerProxyConfig,
				KubeconfigFile: apiServerProxyKubeconfig,
				RefreshCh:      time.NewTicker(apiServerProxyRefreshInterval).C,
			}

			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			if err := p.Run(ctx); err != nil {
				return fmt.Errorf("proxy failed: %w", err)
			}
			return nil
		},
	}
)

func init() {
	apiServerProxyCmd.Flags().StringVar(&apiServerProxyConfig, "config", filepath.Join(os.Getenv("SNAP_DATA"), "args", "apiserver-proxy-config"), "path to apiserver proxy config file")
	apiServerProxyCmd.Flags().StringVar(&apiServerProxyKubeconfig, "kubeconfig", filepath.Join(os.Getenv("SNAP_DATA"), "credentials", "kubelet.config"), "path to kubeconfig file to use for updating control plane nodes")
	apiServerProxyCmd.Flags().DurationVar(&apiServerProxyRefreshInterval, "refresh-interval", 30*time.Second, "refresh interval")

	rootCmd.AddCommand(apiServerProxyCmd)
}
