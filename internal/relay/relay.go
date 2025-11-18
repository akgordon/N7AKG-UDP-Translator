package relay

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/akgordon/UDP-Logger-Relay/internal/config"
	"github.com/akgordon/UDP-Logger-Relay/internal/formatter"
)

// Relay manages the UDP listener and broadcaster
type Relay struct {
	config    *config.Config
	formatter *formatter.Formatter
	listener  *net.UDPConn
	sender    *net.UDPConn
	running   bool
	stopChan  chan bool
	wg        sync.WaitGroup
	mu        sync.RWMutex
}

// New creates a new relay instance
func New(cfg *config.Config) (*Relay, error) {
	f := formatter.New(
		cfg.Formatting.N1MM.Station,
		cfg.Formatting.N1MM.Operator,
		cfg.Formatting.N1MM.Contest,
	)

	return &Relay{
		config:    cfg,
		formatter: f,
		stopChan:  make(chan bool, 1),
	}, nil
}

// Start begins listening for UDP messages and relaying them
func (r *Relay) Start() error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("relay is already running")
	}
	r.running = true
	r.mu.Unlock()

	// Setup UDP listener
	listenAddr := net.JoinHostPort(r.config.Listen.Address, strconv.Itoa(r.config.Listen.Port))
	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve listen address: %w", err)
	}

	r.listener, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to start UDP listener: %w", err)
	}

	// Setup UDP sender
	targetAddr := net.JoinHostPort(r.config.Target.Address, strconv.Itoa(r.config.Target.Port))
	targetUDPAddr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		r.listener.Close()
		return fmt.Errorf("failed to resolve target address: %w", err)
	}

	r.sender, err = net.DialUDP("udp", nil, targetUDPAddr)
	if err != nil {
		r.listener.Close()
		return fmt.Errorf("failed to create UDP sender: %w", err)
	}

	if r.config.Verbose {
		log.Printf("UDP Relay started - listening on %s, forwarding to %s", listenAddr, targetAddr)
	}

	// Start listening for messages
	r.wg.Add(1)
	go r.listen()

	// Wait for stop signal
	<-r.stopChan

	return nil
}

// Stop gracefully stops the relay
func (r *Relay) Stop() {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return
	}
	r.running = false
	r.mu.Unlock()

	if r.config.Verbose {
		log.Println("Stopping UDP relay...")
	}

	// Close connections
	if r.listener != nil {
		r.listener.Close()
	}
	if r.sender != nil {
		r.sender.Close()
	}

	// Signal stop and wait for goroutines
	select {
	case r.stopChan <- true:
	default:
	}

	r.wg.Wait()

	if r.config.Verbose {
		log.Println("UDP relay stopped")
	}
}

// listen continuously listens for incoming UDP messages
func (r *Relay) listen() {
	defer r.wg.Done()

	buffer := make([]byte, 4096)

	for {
		r.mu.RLock()
		running := r.running
		r.mu.RUnlock()

		if !running {
			break
		}

		// Set a read timeout to allow periodic checking of running status
		err := r.listener.SetReadDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			if r.config.Verbose {
				log.Printf("Error setting read deadline: %v", err)
			}
			continue
		}

		n, clientAddr, err := r.listener.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Timeout is expected, continue
				continue
			}
			if r.config.Verbose {
				log.Printf("Error reading UDP message: %v", err)
			}
			continue
		}

		message := string(buffer[:n])

		if r.config.Verbose {
			log.Printf("UDP packet received from %s (%d bytes)", clientAddr, n)
		}

		// Process the message
		go r.processMessage(message, clientAddr, n)
	}
}

// processMessage handles the conversion and forwarding of a single message
func (r *Relay) processMessage(message string, sourceAddr *net.UDPAddr, packetSize int) {
	// Filter messages based on source port - only process messages from expected application ports
	// Common ham radio application UDP ports:
	// 2333 - WSJT-X logging port (what we're listening on)
	// 2237 - Fldigi
	// 2442 - JS8Call
	// 12060 - N1MM Logger Plus
	// Random high ports (like 60463) are typically binary protocol messages - ignore them
	sourcePort := sourceAddr.Port

	// Allow messages from well-known ham radio application ports or the same port we're listening on
	expectedPorts := []int{2333, 2237, 2442, 12060, r.config.Listen.Port}
	isExpectedPort := false
	for _, port := range expectedPorts {
		if sourcePort == port {
			isExpectedPort = true
			break
		}
	}

	// Also allow messages from localhost on any port below 10000 (likely configured applications)
	if sourceAddr.IP.IsLoopback() && sourcePort < 10000 {
		isExpectedPort = true
	}

	if !isExpectedPort {
		// Silently ignore messages from unexpected ports (likely binary protocol)
		return
	}

	// Detect message type if auto-detection is enabled
	var msgType formatter.MessageType
	if r.config.Formatting.AutoDetect {
		msgType = r.formatter.DetectMessageType(message)
	} else {
		msgType = formatter.MessageType(r.config.Formatting.SourceType)
	}

	// Parse the message
	qso, err := r.formatter.ParseMessage(message, msgType)
	if err != nil {
		if r.config.Verbose {
			log.Printf("Skipping message from %s: %v", sourceAddr, err)
		}
		return
	}

	if r.config.Verbose {
		log.Printf("Parsed message type: %s, Callsign: %s, Band: %s, Mode: %s",
			msgType, qso.Callsign, qso.Band, qso.Mode)
	}

	// Convert to N1MM format
	n1mmMessage, err := r.formatter.FormatForN1MM(qso)
	if err != nil {
		if r.config.Verbose {
			log.Printf("Failed to format message for N1MM: %v", err)
		}
		return
	}

	// Send to target
	err = r.sendMessage(n1mmMessage)
	if err != nil {
		log.Printf("Failed to relay packet: %v", err)
		if r.config.Verbose {
			log.Printf("Failed to send message: %v", err)
		}
		return
	}

	// Only log when packet is successfully received and relayed
	log.Printf("UDP packet received (%d bytes) from %s and relayed to %s:%d (QSO: %s on %s %s)",
		packetSize, sourceAddr, r.config.Target.Address, r.config.Target.Port,
		qso.Callsign, qso.Band, qso.Mode)

	if r.config.Verbose {
		log.Printf("N1MM message sent: %s", n1mmMessage)
	}
}

// sendMessage sends a message to the target UDP address
func (r *Relay) sendMessage(message string) error {
	_, err := r.sender.Write([]byte(message))
	return err
}

// GetStats returns statistics about the relay operation
func (r *Relay) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"running":     r.running,
		"listen_addr": fmt.Sprintf("%s:%d", r.config.Listen.Address, r.config.Listen.Port),
		"target_addr": fmt.Sprintf("%s:%d", r.config.Target.Address, r.config.Target.Port),
	}
}
