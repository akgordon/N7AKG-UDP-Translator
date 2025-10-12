package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/akgordon/UDP-Logger-Relay/internal/config"
	"github.com/akgordon/UDP-Logger-Relay/internal/relay"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "udp-logger-relay",
	Short: "UDP Logger Relay - Listen for UDP broadcasts from HF apps and reformat for N1MM",
	Long: `UDP Logger Relay listens for UDP broadcast messages from HF (Ham Radio) applications,
reformats them according to N1MM logger format, and re-broadcasts them via UDP.

This allows integration between various HF logging applications and N1MM Logger Plus.`,
	Run: runRelay,
}

var (
	configFile string
	listenAddr string
	listenPort int
	targetAddr string
	targetPort int
	verbose    bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default is $HOME/.udp-logger-relay.yaml)")
	rootCmd.PersistentFlags().StringVar(&listenAddr, "listen-addr", "0.0.0.0", "address to listen for incoming UDP messages")
	rootCmd.PersistentFlags().IntVar(&listenPort, "listen-port", 2333, "port to listen for incoming UDP messages")
	rootCmd.PersistentFlags().StringVar(&targetAddr, "target-addr", "127.0.0.1", "address to send reformatted UDP messages")
	rootCmd.PersistentFlags().IntVar(&targetPort, "target-port", 12060, "port to send reformatted UDP messages (N1MM default)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("UDP Logger Relay %s (commit: %s, built: %s)\n", version, commit, date)
		},
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runRelay(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config with command line flags if provided
	if cmd.Flag("listen-addr").Changed {
		cfg.Listen.Address = listenAddr
	}
	if cmd.Flag("listen-port").Changed {
		cfg.Listen.Port = listenPort
	}
	if cmd.Flag("target-addr").Changed {
		cfg.Target.Address = targetAddr
	}
	if cmd.Flag("target-port").Changed {
		cfg.Target.Port = targetPort
	}
	if cmd.Flag("verbose").Changed {
		cfg.Verbose = verbose
	}

	if cfg.Verbose {
		log.Printf("Starting UDP Logger Relay...")
		log.Printf("Listening on %s:%d", cfg.Listen.Address, cfg.Listen.Port)
		log.Printf("Forwarding to %s:%d", cfg.Target.Address, cfg.Target.Port)
	}

	// Create and start the relay
	r, err := relay.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create relay: %v", err)
	}

	// Start the relay in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- r.Start()
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Fatalf("Relay error: %v", err)
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
		r.Stop()
	}

	log.Println("UDP Logger Relay stopped")
}
