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
)

// WSJT-X UDP message simulator
// This tool simulates WSJT-X UDP broadcasts for testing the UDP Logger Relay

var (
	targetAddr string
	targetPort int
	interval   int
	mode       string
	callsign   string
	frequency  string
	band       string
)

func init() {
	flag.StringVar(&targetAddr, "addr", "127.0.0.1", "Target address to send UDP messages")
	flag.IntVar(&targetPort, "port", 2333, "Target port to send UDP messages")
	flag.IntVar(&interval, "interval", 5, "Interval between messages in seconds")
	flag.StringVar(&mode, "mode", "FT8", "Mode to simulate (FT8, FT4, MSK144, etc.)")
	flag.StringVar(&callsign, "callsign", "W1TEST", "Callsign to use in messages")
	flag.StringVar(&frequency, "frequency", "14074000", "Frequency in Hz")
	flag.StringVar(&band, "band", "20m", "Band designation")
}

func main() {
	flag.Parse()

	fmt.Println("WSJT-X UDP Simulator")
	fmt.Println("====================")
	fmt.Printf("Target: %s:%d\n", targetAddr, targetPort)
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("Callsign: %s\n", callsign)
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

	// Send initial message immediately
	sendMessage(conn, messageCount)
	messageCount++

	// Send messages periodically
	for {
		select {
		case <-ticker.C:
			sendMessage(conn, messageCount)
			messageCount++
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
			fmt.Printf("Total messages sent: %d\n", messageCount)
			return
		}
	}
}

func sendMessage(conn *net.UDPConn, count int) {
	// Simulate various WSJT-X message types
	// Format: timestamp frequency mode callsign grid snr exchange
	timestamp := time.Now().Format("150405")
	
	messages := []string{
		// CQ message
		fmt.Sprintf("%s %s %s CQ %s FN42", timestamp, frequency, mode, callsign),
		// Reply to CQ
		fmt.Sprintf("%s %s %s %s K2ABC FN42", timestamp, frequency, mode, callsign),
		// Signal report
		fmt.Sprintf("%s %s %s K2ABC %s -15", timestamp, frequency, mode, callsign),
		// RRR confirmation
		fmt.Sprintf("%s %s %s %s K2ABC RRR", timestamp, frequency, mode, callsign),
		// 73 final
		fmt.Sprintf("%s %s %s K2ABC %s 73", timestamp, frequency, mode, callsign),
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
