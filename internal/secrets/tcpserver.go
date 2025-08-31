package secrets

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
)

// RotationCommand represents a command sent via TCP
type RotationCommand struct {
	Action string `json:"action"`
}

// RotationResponse represents the response sent back via TCP
type RotationResponse struct {
	Success       bool   `json:"success"`
	EncryptedData string `json:"encrypted_data,omitempty"`
	Error         string `json:"error,omitempty"`
}

// RotationHandler is a function that handles rotation commands
type RotationHandler func() (string, error)

// TCPRotationServer handles TCP connections for secret rotation
type TCPRotationServer struct {
	address  string
	handler  RotationHandler
	listener net.Listener
	mu       sync.RWMutex
	running  bool
}

// NewTCPRotationServer creates a new TCP server for rotation commands
func NewTCPRotationServer(address string, handler RotationHandler) (*TCPRotationServer, error) {
	if handler == nil {
		return nil, errors.New("rotation handler cannot be nil")
	}

	return &TCPRotationServer{
		address: address,
		handler: handler,
	}, nil
}

// Start starts the TCP server
func (s *TCPRotationServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("server is already running")
	}
	s.running = true
	s.mu.Unlock()

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return fmt.Errorf("failed to listen on %s: %w", s.address, err)
	}

	s.listener = listener

	// Update address to actual listening address (important for :0 ports)
	s.mu.Lock()
	s.address = listener.Addr().String()
	s.mu.Unlock()

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		if err := s.Stop(); err != nil {
			// Log error in production
			fmt.Printf("Error stopping server: %v\n", err)
		}
	}()

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Check if we're shutting down
			s.mu.RLock()
			running := s.running
			s.mu.RUnlock()

			if !running {
				return nil // Clean shutdown
			}

			// Otherwise, this is an actual error
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		// Handle connection in goroutine
		go s.handleConnection(conn)
	}
}

// Stop stops the TCP server
func (s *TCPRotationServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false

	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

// IsRunning returns whether the server is currently running
func (s *TCPRotationServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetAddress returns the server's listening address
func (s *TCPRotationServer) GetAddress() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.address
}

// handleConnection handles a single TCP connection
func (s *TCPRotationServer) handleConnection(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	scanner := bufio.NewScanner(conn)
	encoder := json.NewEncoder(conn)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse command
		var cmd RotationCommand
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			// Try simple string command for backward compatibility
			if strings.EqualFold(line, "rotate") {
				cmd.Action = "rotate"
			} else {
				response := RotationResponse{
					Success: false,
					Error:   "Invalid command format",
				}
				if err := encoder.Encode(response); err != nil {
					// Log error in production
					fmt.Printf("Failed to send error response: %v\n", err)
					return
				}
				continue
			}
		}

		// Handle command
		response := s.handleCommand(cmd)

		// Send response
		if err := encoder.Encode(response); err != nil {
			// Log error in real application
			fmt.Printf("Failed to send response: %v\n", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		// Log error in real application
		fmt.Printf("Connection error: %v\n", err)
	}
}

// handleCommand processes a rotation command
func (s *TCPRotationServer) handleCommand(cmd RotationCommand) RotationResponse {
	switch strings.ToLower(cmd.Action) {
	case "rotate":
		encryptedData, err := s.handler()
		if err != nil {
			return RotationResponse{
				Success: false,
				Error:   err.Error(),
			}
		}

		return RotationResponse{
			Success:       true,
			EncryptedData: encryptedData,
		}

	default:
		return RotationResponse{
			Success: false,
			Error:   "Unknown command: " + cmd.Action,
		}
	}
}
