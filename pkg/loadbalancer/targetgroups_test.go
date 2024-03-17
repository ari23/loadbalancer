package loadbalancer_test

import (
	"testing"

	"github.com/ari23/loadbalancer/pkg/loadbalancer"
	"github.com/ari23/loadbalancer/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewTargetGroupsStore(t *testing.T) {
	store := loadbalancer.NewTargetGroupsStore(&mocks.MockNetDialerInterface{})
	assert.NotNil(t, store, "NewTargetGroupsStore() should not return nil")
	assert.Empty(t, store.GetTargetGroups(), "targetGroups map should be empty")
}

func TestAddTargetGroups(t *testing.T) {
	store := loadbalancer.NewTargetGroupsStore(&mocks.MockNetDialerInterface{})
	configs := []loadbalancer.TargetGroupConfig{
		{
			Name:            "group1",
			UpstreamServers: []string{"192.168.1.1:8081", "192.168.1.1:8082"},
		},
	}

	store.AddTargetGroups(configs)

	assert.Len(t, store.GetTargetGroups(), 1, "There should be 1 target group")
	assert.Len(t, store.GetTargetGroups()["group1"], 2, "There should be 2 upstream servers in 'group1'")
}

func TestGetNextUpstreamServer(t *testing.T) {
	store := loadbalancer.NewTargetGroupsStore(&mocks.MockNetDialerInterface{})
	configs := []loadbalancer.TargetGroupConfig{
		{
			Name:            "group1",
			UpstreamServers: []string{"192.168.1.1:8081"},
		},
	}

	store.AddTargetGroups(configs)

	server, err := store.GetNextUpstreamServer("group1")
	assert.NoError(t, err, "Should not error when getting next upstream server for an existing group")
	assert.NotNil(t, server, "Next upstream server should not be nil")

	_, err = store.GetNextUpstreamServer("nonexistent")
	assert.Error(t, err, "Expected error when getting next upstream server for nonexistent group")
}

func TestGetTargetGroups(t *testing.T) {
	store := loadbalancer.NewTargetGroupsStore(&mocks.MockNetDialerInterface{})
	configs := []loadbalancer.TargetGroupConfig{
		{
			Name:            "group1",
			UpstreamServers: []string{"192.168.1.1:8081", "192.168.1.1:8082"},
		},
	}

	store.AddTargetGroups(configs)

	targetGroups := store.GetTargetGroups()
	assert.Len(t, targetGroups, 1, "There should be 1 target group")
	assert.Len(t, targetGroups["group1"], 2, "There should be 2 upstream servers in 'group1'")
}
