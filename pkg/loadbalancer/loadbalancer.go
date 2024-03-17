package loadbalancer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/ari23/loadbalancer/lib/loadbalance"
	"github.com/ari23/loadbalancer/lib/ratelimit"
)

const (
	dialTimeout time.Duration = 5 * time.Second
	retryLimit  int           = 3
)

// Instance represents an instance of the load balancer.
type Instance struct {
	config                 *LoadBalancerConfig
	tlsConfig              *tls.Config
	authorizedClientsStore *ClientsStore
	targetGroupsStore      *TargetGroupsStore
	netDialer              loadbalance.NetDialerInterface
	wg                     sync.WaitGroup
}

// NewLoadBalancer creates and initializes a new load balancer instance based on the provided
// configuration file.
func NewLoadBalancer(config *LoadBalancerConfig) (*Instance, error) {
	// Init TLS config.
	tlsConfig, err := NewTLSConfig(&config.TLSParams)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	// Initialize the load balancer instance with the parsed configuration.
	lb := &Instance{
		config:    config,
		tlsConfig: tlsConfig,
	}

	// Initialize the authorized clients store.
	lb.authorizedClientsStore = NewClientStore()
	lb.authorizedClientsStore.AddClientsFromClientConfigList(config.Clients)

	lb.netDialer = NewNetDialer(dialTimeout, retryLimit)

	// Initialize target groups store.
	lb.targetGroupsStore = NewTargetGroupsStore(lb.netDialer)
	lb.targetGroupsStore.AddTargetGroups(config.TargetGroups)

	return lb, nil
}

func NewTLSConfig(tlsConfigParams *TLSConfigParams) (*tls.Config, error) {
	if tlsConfigParams.Certificate == "" || tlsConfigParams.PrivateKey == "" || tlsConfigParams.CACertificate == "" {
		return nil, ErrMTLSParamsMissing
	}

	cert, err := tls.LoadX509KeyPair(tlsConfigParams.Certificate, tlsConfigParams.PrivateKey)
	if err != nil {
		return nil, ErrLoadingKeyPair(err)
	}

	// Load CA certificate to verify client certificates
	caCert, err := os.ReadFile(tlsConfigParams.CACertificate)
	if err != nil {
		return nil, ErrLoadingCACert(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,   //nolint:nosnakecase
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, //nolint:nosnakecase
		},
	}

	return config, nil
}

func (i *Instance) GetConfig() *LoadBalancerConfig {
	return i.config
}

func (i *Instance) GetTLSConfig() *tls.Config {
	return i.tlsConfig
}

func (i *Instance) GetAuthorizedClientsStore() *ClientsStore {
	return i.authorizedClientsStore
}

func (i *Instance) GetTargetGroupsStore() *TargetGroupsStore {
	return i.targetGroupsStore
}

// Start begins listening on the configured port and handling incoming connections.
func (i *Instance) Start(ctx context.Context) error {
	listenAddr := i.config.ListenAddress

	listener, err := tls.Listen("tcp", listenAddr, i.tlsConfig)
	if err != nil {
		i.config.Logger.Errorf("Failed to listen on %s: %v", listenAddr, err)

		return err
	}

	defer listener.Close()

	i.config.Logger.Infof("Network Load balancer listening on %s", listenAddr)

	// Start health checks for all target groups.
	i.targetGroupsStore.StartHealthChecks(ctx, &i.wg)

	i.wg.Add(1)
	// Listen for context cancellation to gracefully shut down the listener.
	go func() {
		<-ctx.Done()
		listener.Close()
		i.wg.Done()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				i.config.Logger.Infof("Shutting down listener on %s", listenAddr)
				i.wg.Wait()

				return nil
			default:
				// if listener is closed (i.e. socket got closed for some reason) then bail.
				if errors.Is(err, net.ErrClosed) {
					i.config.Logger.Infof("Listener closed on %s", listenAddr)

					return err
				}

				i.config.Logger.Errorf("Failed to accept connection: %v", err)
			}

			continue
		}

		// Handle the connection in a new goroutine.
		i.wg.Add(1)

		go func(conn net.Conn) {
			defer i.wg.Done()
			i.handleConnection(ctx, conn)
		}(conn)
	}
}

func (i *Instance) handleConnection(ctx context.Context, clientConn net.Conn) {
	defer clientConn.Close()

	// Set a deadline for read/write operations on the connection.
	// TODO: Make this configurable.
	timeoutDuration := 30 * time.Second
	clientConn.SetDeadline(time.Now().Add(timeoutDuration))

	clientInfo, err := GetClientConfigFromConn(clientConn, i.authorizedClientsStore)
	if err != nil {
		i.config.Logger.Errorf("[handleConnection] Error: %s", err.Error())

		return
	}

	i.config.Logger.Infof("ClientInfo: %+v", clientInfo)

	if !ratelimit.RequestAllowed(clientInfo) {
		i.config.Logger.Errorf("[handleConnection] Rate limit exceeded for client: %s", clientInfo.GetClientID())

		return
	}

	upstreamServer, err := GetNextUpstreamServer(clientInfo, i.targetGroupsStore)
	if err != nil {
		i.config.Logger.Errorf("[handleConnection] Error: %s", err.Error())

		return
	}

	upstreamServer.IncrementConnectionCount()
	defer upstreamServer.DecrementConnectionCount()

	upstreamConn, err := i.netDialer.DialTimeout(
		"tcp", upstreamServer.GetAddress(), dialTimeout)
	if err != nil {
		i.config.Logger.Errorf("Failed to dial upstream server: %v", err)

		return
	}
	defer upstreamConn.Close()

	upstreamConn.SetDeadline(time.Now().Add(timeoutDuration))

	i.wg.Add(1)

	// TODO: with current client.go and server.go combo, receiving this error:
	// ERRO[0256] [client_to_upstream] Error: readfrom tcp 127.0.0.1:52155->127.0.0.1:8081: read tcp
	// 127.0.0.1:8080->127.0.0.1:52154: use of closed network connection.
	// This needs to be debugged. Although this error doesn't show up when using real client such as
	// netcat.
	go func() {
		defer i.wg.Done()

		if _, err := io.Copy(upstreamConn, clientConn); err != nil {
			i.config.Logger.Errorf("[client_to_upstream] Error: %s", err.Error())

			return
		}
	}()

	if _, err := io.Copy(clientConn, upstreamConn); err != nil {
		i.config.Logger.Errorf("[upstream_to_client] Error: %s", err.Error())
	}
}
