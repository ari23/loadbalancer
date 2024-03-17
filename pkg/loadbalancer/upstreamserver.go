package loadbalancer

import (
	"sync"

	"github.com/ari23/loadbalancer/lib/loadbalance"
)

// UpstreamServer is a struct that represents an upstream server. It implements
// the loadbalance.UpstreamServerInterface.
type UpstreamServer struct {
	address string
	healthy bool
	numConn int
	mu      sync.Mutex
}

func NewUpstreamServer(address string) loadbalance.UpstreamServerInterface {
	return &UpstreamServer{
		address: address,
		healthy: false,
		numConn: 0,
	}
}

func (u *UpstreamServer) GetAddress() string {
	return u.address
}

func (u *UpstreamServer) IsHealthy() bool {
	return u.healthy
}

func (u *UpstreamServer) SetHealthy(healthy bool) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.healthy = healthy
}

func (u *UpstreamServer) GetConnectionCount() int {
	return u.numConn
}

func (u *UpstreamServer) IncrementConnectionCount() {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.numConn++
}

func (u *UpstreamServer) DecrementConnectionCount() {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.numConn--
}
