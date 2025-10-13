// VarAC Message Format Demo
//
// This example demonstrates how UDP Logger Relay detects and parses VarAC messages.
// VarAC is a popular HF digital mode application that can send UDP broadcasts
// when QSOs are completed.
//
// To run this demo:
//   go run examples/varac_demo.go

package main

import (
	"fmt"

	"github.com/akgordon/UDP-Logger-Relay/internal/formatter"
)

func main() {
	fmt.Println("UDP Logger Relay - VarAC Message Format Demo")
	fmt.Println("===========================================")
	fmt.Println("This demo shows how VarAC messages are detected and parsed.")
	fmt.Println()

	f := formatter.New("W1TEST", "OP", "DEMO")

	// Test VarAC message detection and parsing
	testMessages := []struct {
		description string
		message     string
	}{
		{
			"Full JSON format with all fields (typical VarAC UDP broadcast)",
			`{"app":"VarAC","call":"W1ABC","freq":"14.105","mode":"VARA HF","timestamp":"2023-10-12 14:30:00","rst_sent":"599","rst_rcvd":"599","band":"20m"}`,
		},
		{
			"Plain text format with frequency and mode",
			"QSO with VK2XYZ on 7.105 VARA",
		},
		{
			"VarAC completion message (simple format)",
			"VarAC QSO completed with EA1ABC",
		},
		{
			"Minimal JSON format (some VarAC configurations)",
			`{"call":"JA1DEF","frequency":"21.105","mode":"VARA"}`,
		},
		{
			"VARA FM format example",
			`{"app":"VarAC","call":"VK3ABC","freq":"145.500","mode":"VARA FM","band":"2m"}`,
		},
	}

	for i, test := range testMessages {
		fmt.Printf("\nTest %d: %s\n", i+1, test.description)
		fmt.Printf("Message: %s\n", test.message)

		// Detect message type
		msgType := f.DetectMessageType(test.message)
		fmt.Printf("Detected type: %s\n", msgType)

		// Parse the message
		qso, err := f.ParseMessage(test.message, msgType)
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			continue
		}

		fmt.Printf("Parsed QSO:\n")
		fmt.Printf("  Callsign: %s\n", qso.Callsign)
		fmt.Printf("  Frequency: %s\n", qso.Frequency)
		fmt.Printf("  Mode: %s\n", qso.Mode)
		fmt.Printf("  Band: %s\n", qso.Band)
		fmt.Printf("  RST Sent: %s\n", qso.RST_Sent)
		fmt.Printf("  RST Rcvd: %s\n", qso.RST_Rcvd)

		// Format for N1MM
		n1mmXML, err := f.FormatForN1MM(qso)
		if err != nil {
			fmt.Printf("N1MM format error: %v\n", err)
			continue
		}

		fmt.Printf("N1MM XML (first 200 chars): %s...\n", n1mmXML[:min(200, len(n1mmXML))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
