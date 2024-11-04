package mdns

import (
	"net"
	"os"
	"s3MediaStreamer/app/internal/logs"
	"strconv"

	"github.com/hashicorp/mdns"
)

type Repository interface {
}

type Service struct {
	serviceName string
	port        int
	logger      *logs.Logger
	stopCh      chan struct{}
}

// NewMDNSService creates a new mDNS service instance.
func NewMDNSService(serviceName, port string, logger *logs.Logger) *Service {
	// Convert port from string to int
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil
	}

	return &Service{
		serviceName: serviceName,
		port:        portInt,
		logger:      logger,
		stopCh:      make(chan struct{}),
	}
}

// Start begins the mDNS service in a separate goroutine.
func (s *Service) Start() {
	go func() {
		// Get the hostname and IP addresses of the current machine
		host, _ := os.Hostname()
		ips, _ := net.LookupIP(host)
		info := []string{"Local Audio Streaming Service"}

		// Create and configure the mDNS service
		service, err := mdns.NewMDNSService(host, s.serviceName, "", "", s.port, ips, info)
		if err != nil {
			s.logger.Fatalf("Failed to start mDNS: %v", err)
		}

		// Start the mDNS server
		server, err := mdns.NewServer(&mdns.Config{Zone: service})
		if err != nil {
			s.logger.Fatalf("Failed to start mDNS server: %v", err)
		}
		defer func(server *mdns.Server) {
			err = server.Shutdown()
			if err != nil {
				s.logger.Fatalf("Failed to shutdown mDNS server: %v", err)
			}
		}(server)

		s.logger.Infof("mDNS service '%s' started on port %d", s.serviceName, s.port)

		// Wait until the stop signal is received
		<-s.stopCh
		s.logger.Info("mDNS service stopped")
	}()
}

// Stop gracefully stops the mDNS service.
func (s *Service) Stop() {
	close(s.stopCh) // Signal to stop the mDNS server
}
