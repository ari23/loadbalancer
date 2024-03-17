package loadbalancer

import (
	"sync"
	"time"

	"github.com/ari23/loadbalancer/lib/ratelimit"
)

type ClientInfo struct {
	mu                   sync.Mutex
	clientID             string
	allowedTargetGroup   string
	maxConnections       int
	connections          int
	maxRequestsPerWindow int
	lastWindow           time.Time
	requestCount         int
}

func (c *ClientInfo) GetClientID() string {
	return c.clientID
}

func (c *ClientInfo) GetAllowedTargetGroup() string {
	return c.allowedTargetGroup
}

func (c *ClientInfo) GetMaxConnections() int {
	return c.maxConnections
}

func (c *ClientInfo) GetConnections() int {
	return c.connections
}

func (c *ClientInfo) GetMaxRequestsPerWindow() int {
	return c.maxRequestsPerWindow
}

func (c *ClientInfo) GetLastWindow() time.Time {
	return c.lastWindow
}

func (c *ClientInfo) GetRequestCount() int {
	return c.requestCount
}

func (c *ClientInfo) SetLastWindow(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastWindow = now
	c.requestCount = 0
}

func (c *ClientInfo) IncrementRequestCount() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requestCount++
}

// NewClientInfo initializes a ClientInfo with specified limits.
func NewClientInfo(clientID, allowedTargetGroup string, maxRequestsPerWindow, maxConnections int) *ClientInfo {
	return &ClientInfo{
		clientID:             clientID,
		maxRequestsPerWindow: maxRequestsPerWindow,
		maxConnections:       maxConnections,
		allowedTargetGroup:   allowedTargetGroup,
		lastWindow:           time.Now(),
	}
}

type ClientsStore struct {
	// authorizedClients is a map of client name to client config.
	authorizedClients map[string]ratelimit.ClientInfoInterface
	mu                sync.RWMutex
}

func NewClientStore() *ClientsStore {
	return &ClientsStore{
		authorizedClients: make(map[string]ratelimit.ClientInfoInterface),
	}
}

func (cs *ClientsStore) AddClientsFromClientConfigList(clients []ClientConfig) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for _, c := range clients {
		clientInfo := NewClientInfo(
			c.ClientId, c.AllowedTargetGroup, c.RequestsPerSecond, 5)

		cs.authorizedClients[c.ClientId] = clientInfo
	}
}

func (cs *ClientsStore) GetClient(clientId string) (ratelimit.ClientInfoInterface, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	clientInfo, ok := cs.authorizedClients[clientId]

	return clientInfo, ok
}

func (cs *ClientsStore) GetClients() map[string]ratelimit.ClientInfoInterface {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	return cs.authorizedClients
}
