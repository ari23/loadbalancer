package loadbalancer

import (
	"context"
	"sync"

	"github.com/ari23/loadbalancer/lib/loadbalance"
)

// TargetGroupsStore is the store for the target groups.
type TargetGroupsStore struct {
	// targetGroups is a map of target group name to upstream servers.
	targetGroups map[string][]loadbalance.UpstreamServerInterface
	mu           sync.RWMutex
	netDialer    loadbalance.NetDialerInterface
}

func NewTargetGroupsStore(dialer loadbalance.NetDialerInterface) *TargetGroupsStore {
	return &TargetGroupsStore{
		targetGroups: make(map[string][]loadbalance.UpstreamServerInterface),
		netDialer:    dialer,
	}
}

func (t *TargetGroupsStore) AddTargetGroups(targetGroups []TargetGroupConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, tg := range targetGroups {
		upstreamServers := make([]loadbalance.UpstreamServerInterface, len(tg.UpstreamServers))
		for i, address := range tg.UpstreamServers {
			upstreamServers[i] = NewUpstreamServer(address)
		}

		t.targetGroups[tg.Name] = upstreamServers
	}
}

// StartHealthChecks starts the health checks for all the target groups.
func (t *TargetGroupsStore) StartHealthChecks(ctx context.Context, wg *sync.WaitGroup) {
	for _, upstreamServers := range t.targetGroups {
		for _, upstream := range upstreamServers {
			wg.Add(1)

			go func(server loadbalance.UpstreamServerInterface) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					return
				default:
					loadbalance.HealthCheck(ctx, server, t.netDialer)
				}
			}(upstream)
		}
	}
}

func (t *TargetGroupsStore) GetNextUpstreamServer(targetGroupName string) (loadbalance.UpstreamServerInterface, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	upstreamServers, ok := t.targetGroups[targetGroupName]
	if !ok {
		return nil, ErrTargetGroupNotFound(targetGroupName)
	}

	nextUpstreamServer, err := loadbalance.GetNextUpstreamServer(upstreamServers)
	if err != nil {
		return nil, err
	}

	return nextUpstreamServer, nil
}

func (t *TargetGroupsStore) GetTargetGroups() map[string][]loadbalance.UpstreamServerInterface {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.targetGroups
}
