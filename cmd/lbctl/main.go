package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ari23/loadbalancer/pkg/loadbalancer"
	"github.com/ari23/loadbalancer/utils"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lbctl",
	Short: "CLI tool for managing the TCP load balancer",
}

var configPath string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the TCP load balancer",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Set up signal handling to cancel the context on SIGINT or SIGTERM
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigChan // Wait for a signal
			fmt.Fprintf(os.Stderr, "Received signal: %v, shutting down...\n", sig)
			cancel() // Cancel the context
		}()

		// Parse the configuration file.
		dataFile, err := os.Open(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open config file: %v", err)
			return err
		}

		defer dataFile.Close()

		config, err := loadbalancer.ParseConfig(dataFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse configuration: %v\n", err)
			return err
		}

		// Init logger.
		config.Logger = utils.NewLogger(config.LogLevel)

		lb, err := loadbalancer.NewLoadBalancer(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create Load Balancer instance due to: %v\n", err)
			return err
		}

		if err := lb.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start Load Balancer due to: %v\n", err)
			return err
		}

		return nil
	},
}

func init() {
	startCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to the load balancer configuration file")
	rootCmd.AddCommand(startCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
