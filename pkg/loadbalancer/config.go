package loadbalancer

import (
	"io"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// LoadBalancerConfig is the configuration for the load balancer.
type LoadBalancerConfig struct {
	ListenAddress string              `yaml:"listenAddress"`
	TLSParams     TLSConfigParams     `yaml:"tlsParams"`
	LogLevel      string              `yaml:"logLevel"`
	TargetGroups  []TargetGroupConfig `yaml:"targetGroups"`
	Clients       []ClientConfig      `yaml:"clients"`
	Logger        *logrus.Logger
}

// TLSConfigParams is the configuration for the TLS.
type TLSConfigParams struct {
	// Load balancer certificate.
	Certificate string `yaml:"certificate"`
	// Load balancer private key.
	PrivateKey string `yaml:"privateKey"`
	// Root CA's certificate. Needed for self-signed certificate support.
	CACertificate string `yaml:"caCert"`
}

// TargetGroupConfig is the configuration for the target group.
// A TargetGroup is a collection of upstream servers serving a particular
// application e.g. Frontend, DB etc.
type TargetGroupConfig struct {
	Name            string   `yaml:"name"`
	UpstreamServers []string `yaml:"upstreamServers"`
}

// ClientConfig is the configuration for the client access.
type ClientConfig struct {
	// ClientId is the unique identifier for the client. This should match the CN in the client's
	// certificate.
	ClientId string `yaml:"clientId"`
	// AllowedTargetGroup is the target group that the client is allowed to access.
	AllowedTargetGroup string `yaml:"allowedTargetGroup"`
	RequestsPerSecond  int    `yaml:"requestsPerSecond"`
}

// ParseConfig parses the load balancer configuration from the given file.
func ParseConfig(data io.Reader) (*LoadBalancerConfig, error) {
	configData, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}

	config := &LoadBalancerConfig{}
	if err = yaml.Unmarshal(configData, config); err != nil {
		return nil, err
	}

	return config, nil
}
