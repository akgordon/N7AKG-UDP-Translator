package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

// WSJT-X UDP message simulator
// This tool simulates WSJT-X UDP broadcasts for testing the UDP Logger Relay

// Config represents the simulator configuration
type Config struct {
	Target struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port"`
	} `yaml:"target"`
	Timing struct {
		Interval int `yaml:"interval"`
	} `yaml:"timing"`
	Radio struct {
		Mode      string `yaml:"mode"`
		Callsign  string `yaml:"callsign"`
		Frequency string `yaml:"frequency"`
		Band      string `yaml:"band"`
		Grid      string `yaml:"grid"`
	} `yaml:"radio"`
	Messages struct {
		RemoteCallsign string `yaml:"remote_callsign"`
		RemoteGrid     string `yaml:"remote_grid"`
		SignalReport   string `yaml:"signal_report"`
	} `yaml:"messages"`
}

var (
	configFile string
	targetAddr string
	targetPort int
	interval   int
	mode       string
	callsign   string
	frequency  string
	band       string
	grid       string

	// Track which flags were explicitly set
	addrSet      bool
	portSet      bool
	intervalSet  bool
	modeSet      bool
	callsignSet  bool
	frequencySet bool
	bandSet      bool
	gridSet      bool
)

func init() {
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.StringVar(&targetAddr, "addr", "127.0.0.1", "Target address to send UDP messages")
	flag.IntVar(&targetPort, "port", 2333, "Target port to send UDP messages")
	flag.IntVar(&interval, "interval", 5, "Interval between messages in seconds")
	flag.StringVar(&mode, "mode", "FT8", "Mode to simulate (FT8, FT4, MSK144, etc.)")
	flag.StringVar(&callsign, "callsign", "W1TEST", "Callsign to use in messages")
	flag.StringVar(&frequency, "frequency", "14074000", "Frequency in Hz")
	flag.StringVar(&band, "band", "20m", "Band designation")
	flag.StringVar(&grid, "grid", "FN42", "Grid square")
}

// loadConfig loads configuration from a YAML file
func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func main() {
	// Store original values to detect changes
	origAddr := targetAddr
	origPort := targetPort
	origInterval := interval
	origMode := mode
	origCallsign := callsign
	origFreq := frequency
	origBand := band
	origGrid := grid

	flag.Parse()

	// Detect which flags were changed
	addrSet = targetAddr != origAddr
	portSet = targetPort != origPort
	intervalSet = interval != origInterval
	modeSet = mode != origMode
	callsignSet = callsign != origCallsign
	frequencySet = frequency != origFreq
	bandSet = band != origBand
	gridSet = grid != origGrid

	// Load config file if specified
	var cfg *Config
	if configFile != "" {
		var err error
		cfg, err = loadConfig(configFile)
		if err != nil {
			log.Fatalf("Error loading config file: %v", err)
		}

		// Use config file values for flags that weren't explicitly set
		if !addrSet {
			targetAddr = cfg.Target.Address
		}
		if !portSet {
			targetPort = cfg.Target.Port
		}
		if !intervalSet {
			interval = cfg.Timing.Interval
		}
		if !modeSet {
			mode = cfg.Radio.Mode
		}
		if !callsignSet {
			callsign = cfg.Radio.Callsign
		}
		if !frequencySet {
			frequency = cfg.Radio.Frequency
		}
		if !bandSet {
			band = cfg.Radio.Band
		}
		if !gridSet && cfg.Radio.Grid != "" {
			grid = cfg.Radio.Grid
		}
	}

	fmt.Println("WSJT-X UDP Simulator")
	fmt.Println("====================")
	if configFile != "" {
		fmt.Printf("Config File: %s\n", configFile)
	}
	fmt.Printf("Target: %s:%d\n", targetAddr, targetPort)
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("Callsign: %s\n", callsign)
	fmt.Printf("Grid: %s\n", grid)
	fmt.Printf("Frequency: %s Hz (%s)\n", frequency, band)
	fmt.Printf("Interval: %d seconds\n", interval)
	fmt.Println("====================")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Setup UDP connection
	addr := fmt.Sprintf("%s:%d", targetAddr, targetPort)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create ticker for periodic messages
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	messageCount := 0

	// Determine message parameters
	var remoteCall, remoteGrid, signalReport string
	if cfg != nil {
		remoteCall = cfg.Messages.RemoteCallsign
		remoteGrid = cfg.Messages.RemoteGrid
		signalReport = cfg.Messages.SignalReport
	}
	if remoteCall == "" {
		remoteCall = "K2ABC"
	}
	if remoteGrid == "" {
		remoteGrid = "FN31"
	}
	if signalReport == "" {
		signalReport = "-15"
	}

	// Send initial message immediately
	sendMessage(conn, messageCount, remoteCall, remoteGrid, signalReport)
	messageCount++

	// Send messages periodically
	for {
		select {
		case <-ticker.C:
			sendMessage(conn, messageCount, remoteCall, remoteGrid, signalReport)
			messageCount++
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
			fmt.Printf("Total messages sent: %d\n", messageCount)
			return
		}
	}
}

func sendMessage(conn *net.UDPConn, count int, remoteCall, remoteGrid, signalReport string) {
	// Simulate various WSJT-X message types
	// Format: timestamp frequency mode callsign grid snr exchange
	timestamp := time.Now().Format("150405")

	messages := []string{
		// CQ message
		fmt.Sprintf("%s %s %s CQ %s %s", timestamp, frequency, mode, callsign, grid),
		// Reply to CQ
		fmt.Sprintf("%s %s %s %s %s %s", timestamp, frequency, mode, callsign, remoteCall, remoteGrid),
		// Signal report
		fmt.Sprintf("%s %s %s %s %s %s", timestamp, frequency, mode, remoteCall, callsign, signalReport),
		// RRR confirmation
		fmt.Sprintf("%s %s %s %s %s RRR", timestamp, frequency, mode, callsign, remoteCall),
		// 73 final
		fmt.Sprintf("%s %s %s %s %s 73", timestamp, frequency, mode, remoteCall, callsign),
	}

	// Cycle through message types
	message := messages[count%len(messages)]

	// Send the message
	n, err := conn.Write([]byte(message))
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}

	fmt.Printf("[%d] Sent %d bytes: %s\n", count+1, n, message)
}
