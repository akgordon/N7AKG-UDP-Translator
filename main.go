package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/akgordon/N7AKG-UDP-Translator/internal/config"
	"github.com/akgordon/N7AKG-UDP-Translator/internal/relay"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "N7AKG-UDP-Translator",
	Short: "UDP Logger Relay - Listen for UDP broadcasts from HF apps and reformat for N1MM",
	Long: `UDP Logger Relay listens for UDP broadcast messages from HF (Ham Radio) applications,
reformats them according to N1MM logger format, and re-broadcasts them via UDP.

This allows integration between various HF logging applications and N1MM Logger Plus.

Supported Source Applications:
  • WSJT-X (FT8, FT4, MSK144, etc.)
  • JS8Call
  • Fldigi (PSK31, RTTY, etc.)
  • VaraC
  • N1MM Logger Plus (pass-through)

Examples:
  # Start with default settings (listen on 2333, forward to N1MM on 12060)
  N7AKG-UDP-Translator

  # Use custom ports and addresses
  N7AKG-UDP-Translator --listen-port 2334 --target-addr 192.168.1.100 --target-port 12061

  # Enable verbose logging
  N7AKG-UDP-Translator --verbose

  # Use a specific configuration file
  N7AKG-UDP-Translator --config /path/to/config.yaml

  # Force specific source type (disable auto-detection)
  N7AKG-UDP-Translator --source-type wsjt-x`,
	Run: runRelay,
}

var (
	configFile string
	listenAddr string
	listenPort int
	targetAddr string
	targetPort int
	sourceType string
	verbose    bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default is $HOME/.N7AKG-UDP-Translator.yaml)")
	rootCmd.PersistentFlags().StringVar(&listenAddr, "listen-addr", "0.0.0.0", "address to listen for incoming UDP messages")
	rootCmd.PersistentFlags().IntVar(&listenPort, "listen-port", 2333, "port to listen for incoming UDP messages")
	rootCmd.PersistentFlags().StringVar(&targetAddr, "target-addr", "127.0.0.1", "address to send reformatted UDP messages")
	rootCmd.PersistentFlags().IntVar(&targetPort, "target-port", 12060, "port to send reformatted UDP messages (N1MM default)")
	rootCmd.PersistentFlags().StringVar(&sourceType, "source-type", "auto", "expected source message type (auto, wsjt-x, fldigi, js8call, varac, n1mm)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("UDP Logger Relay %s (commit: %s, built: %s)\n", version, commit, date)
		},
	})

	// Add help command with extended information
	rootCmd.AddCommand(&cobra.Command{
		Use:   "help-extended",
		Short: "Show extended help and configuration information",
		Long:  "Display detailed help including configuration file format, supported source types, and troubleshooting information.",
		Run: func(cmd *cobra.Command, args []string) {
			showExtendedHelp()
		},
	})
}

func showExtendedHelp() {
	fmt.Println("UDP Logger Relay - Extended Help")
	fmt.Println("================================")
	fmt.Println()

	fmt.Println("COMMAND LINE FLAGS:")
	fmt.Println("  -c, --config <file>        Configuration file path")
	fmt.Println("      --listen-addr <addr>   Listen address (default: 0.0.0.0)")
	fmt.Println("      --listen-port <port>   Listen port (default: 2333)")
	fmt.Println("      --target-addr <addr>   Target address (default: 127.0.0.1)")
	fmt.Println("      --target-port <port>   Target port (default: 12060)")
	fmt.Println("      --source-type <type>   Source type: auto, wsjt-x, fldigi, js8call, varac, n1mm")
	fmt.Println("  -v, --verbose              Enable verbose logging")
	fmt.Println("  -h, --help                 Show basic help")
	fmt.Println()

	fmt.Println("SUPPORTED SOURCE TYPES:")
	fmt.Println("  auto     - Auto-detect message type (recommended)")
	fmt.Println("  wsjt-x   - WSJT-X applications (FT8, FT4, MSK144, etc.)")
	fmt.Println("  js8call  - JS8Call digital mode")
	fmt.Println("  fldigi   - Fldigi (PSK31, RTTY, CW, etc.)")
	fmt.Println("  varac    - VaraC HF digital mode")
	fmt.Println("  n1mm     - N1MM Logger Plus (pass-through)")
	fmt.Println()

	fmt.Println("CONFIGURATION FILE:")
	fmt.Println("  Create a YAML file with the following structure:")
	fmt.Println("  ```")
	fmt.Println("  listen:")
	fmt.Println("    address: \"0.0.0.0\"")
	fmt.Println("    port: 2333")
	fmt.Println("  target:")
	fmt.Println("    address: \"127.0.0.1\"")
	fmt.Println("    port: 12060")
	fmt.Println("  formatting:")
	fmt.Println("    source_type: \"auto\"")
	fmt.Println("    auto_detect: true")
	fmt.Println("    n1mm:")
	fmt.Println("      station: \"YOUR_CALL\"")
	fmt.Println("      operator: \"YOUR_CALL\"")
	fmt.Println("      contest: \"GENERAL\"")
	fmt.Println("  verbose: false")
	fmt.Println("  ```")
	fmt.Println()

	fmt.Println("DEFAULT PORTS:")
	fmt.Println("  2333   - Common WSJT-X UDP broadcast port")
	fmt.Println("  12060  - N1MM Logger Plus default UDP port")
	fmt.Println("  2237   - Fldigi default UDP port")
	fmt.Println("  2442   - JS8Call default UDP port")
	fmt.Println()

	fmt.Println("TROUBLESHOOTING:")
	fmt.Println("  • Use --verbose flag to see detailed message flow")
	fmt.Println("  • Check firewall settings for UDP ports")
	fmt.Println("  • Verify source application is broadcasting UDP messages")
	fmt.Println("  • Ensure N1MM Logger is listening on target port")
	fmt.Println("  • Use 'netstat -an | findstr UDP' to check port usage")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runRelay(cmd *cobra.Command, args []string) {
	// Display startup message
	fmt.Printf("UDP Logger Relay %s starting up...\n", version)
	fmt.Printf("Built: %s (commit: %s)\n", date, commit)
	fmt.Println("=========================================")

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
	if cmd.Flag("source-type").Changed {
		cfg.Formatting.SourceType = sourceType
	}
	if cmd.Flag("verbose").Changed {
		cfg.Verbose = verbose
	}

	// Display configuration information
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Listen Address: %s:%d\n", cfg.Listen.Address, cfg.Listen.Port)
	fmt.Printf("  Target Address: %s:%d\n", cfg.Target.Address, cfg.Target.Port)
	fmt.Printf("  Source Type:    %s\n", cfg.Formatting.SourceType)
	fmt.Printf("  Verbose Mode:   %t\n", cfg.Verbose)
	fmt.Println("=========================================")

	fmt.Printf("Start with option \"help\" to see all command line options.\n\n")

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
