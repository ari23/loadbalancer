package loadbalancer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ari23/loadbalancer/pkg/loadbalancer"
)

// TestNewLoadBalancer tests the creation of a new load balancer instance.
func TestNewLoadBalancer(t *testing.T) {
	clientA := loadbalancer.ClientConfig{
		ClientId:           "clientA.bardomain.com",
		AllowedTargetGroup: "group1",
		RequestsPerSecond:  5,
	}

	clientB := loadbalancer.ClientConfig{
		ClientId:           "clientB.bardomain.com",
		AllowedTargetGroup: "group2",
		RequestsPerSecond:  10,
	}

	targetGroup1 := loadbalancer.TargetGroupConfig{
		Name:            "group1",
		UpstreamServers: []string{"192.168.1.1:8081", "192.168.1.1:8082"},
	}

	targetGroup2 := loadbalancer.TargetGroupConfig{
		Name:            "group2",
		UpstreamServers: []string{"192.168.2.1:8083", "192.168.2.1:8084"},
	}

	config := &loadbalancer.LoadBalancerConfig{
		ListenAddress: "10.0.0.1:8080",
		TLSParams: loadbalancer.TLSConfigParams{
			Certificate:   "../../certs/loadbalancer.crt",
			PrivateKey:    "../../certs/loadbalancer.key",
			CACertificate: "../../certs/rootCA.pem",
		},
		Clients:      []loadbalancer.ClientConfig{clientA, clientB},
		TargetGroups: []loadbalancer.TargetGroupConfig{targetGroup1, targetGroup2},
	}

	lb, err := loadbalancer.NewLoadBalancer(config)
	if err != nil {
		t.Errorf("Failed to create load balancer: %v", err)
	}

	// Assert that the load balancer's config matches the input config
	assert.Equal(t, lb.GetConfig(), config, "Load balancer config does not match the input config")

	// Assert that TLS configuration is correctly initialized
	assert.NotNil(t, lb.GetTLSConfig())

	// Assert that the authorized clients store is initialized and contains the expected clients
	assert.Equal(t, len(lb.GetAuthorizedClientsStore().GetClients()), len(config.Clients))

	// Assert that the target groups store is initialized and contains the expected target groups
	assert.Equal(t, len(lb.GetTargetGroupsStore().GetTargetGroups()), len(config.TargetGroups))
}
