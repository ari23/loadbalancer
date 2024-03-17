package loadbalance

import (
	"net"
	"time"
)

type UpstreamServerInterface interface {
	// GetAddress returns the address of the server.
	GetAddress() string
	// IsHealthy returns the health status of the server.
	IsHealthy() bool
	// SetHealthy sets the health status of the server.
	SetHealthy(healthy bool)
	// IncrementConnectionCount increments the number of connections to the server by 1.
	IncrementConnectionCount()
	// DecrementConnectionCount decrements the number of connections to the server by 1.
	DecrementConnectionCount()
	// GetConnectionCount returns the number of connections to the server.
	GetConnectionCount() int
}

type NetDialerInterface interface {
	Dial(network, address string) (net.Conn, error)
	DialTimeout(network, address string, timeout time.Duration) (net.Conn, error)
	GetTimeout() time.Duration
	GetRetryLimit() int
}
