package loadbalancer

import (
	"net"
	"time"

	"github.com/ari23/loadbalancer/lib/loadbalance"
)

// NetDialer uses the standard library's net.Dial function to dial network addresses.
type NetDialer struct {
	timeout    time.Duration
	retryLimit int
}

// NewNetDialer creates a new NetDialer.
func NewNetDialer(timeout time.Duration, retryLimit int) loadbalance.NetDialerInterface {
	return &NetDialer{
		timeout:    timeout,
		retryLimit: retryLimit,
	}
}

// Dial makes a network connection to the specified address.
func (d *NetDialer) Dial(network, address string) (net.Conn, error) {
	return net.Dial(network, address)
}

func (d *NetDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, address, timeout)
}

func (d *NetDialer) GetTimeout() time.Duration {
	return d.timeout
}

func (d *NetDialer) GetRetryLimit() int {
	return d.retryLimit
}
