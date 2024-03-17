package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ari23/loadbalancer/utils"
	"github.com/sirupsen/logrus"
)

const (
	shortLivedDelayInSeconds time.Duration = 0
	longLivedDelayInSeconds  time.Duration = 10 * time.Second
	defaultLogLevel          string        = "info"
	bytesToWrite             int           = 1024
)

type Client struct {
	numRequests int
	serverAddr  string
	longLived   bool
	tlsConfig   *tls.Config
	logger      *logrus.Logger
}

func main() {
	var (
		serverAddr  string
		numRequests int
		longLived   bool
		clientCert  string
		clientKey   string
		rootCAcert  string
	)

	flag.StringVar(&serverAddr, "server", "127.0.0.1:8080", "Server address in the format host:port")
	flag.IntVar(&numRequests, "requests", 1, "Number of requests to send")
	flag.BoolVar(&longLived, "long", false, "Use long-lived simultaneous connections")
	flag.StringVar(&clientCert, "client-cert", "certs/clientA.crt", "Client certificate")
	flag.StringVar(&clientKey, "client-key", "certs/clientA.key", "Client private key")
	flag.StringVar(&rootCAcert, "root-ca", "certs/rootCA.pem", "Root CA certificate")
	flag.Parse()

	tlsConfig, err := LoadTLSConfig(clientCert, clientKey, rootCAcert)
	if err != nil {
		log.Fatalf("client: load tls config: %s", err)
	}

	logger := utils.NewLogger(defaultLogLevel)

	client := &Client{
		numRequests: numRequests,
		serverAddr:  serverAddr,
		longLived:   longLived,
		tlsConfig:   tlsConfig,
		logger:      logger,
	}

	if err := client.GenerateTraffic(); err != nil {
		log.Fatalf("client: generate traffic: %s", err)
	}
}

func LoadTLSConfig(clientCert, clientKey, rootCAcert string) (*tls.Config, error) {
	// Load client's certificate and private key
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		log.Fatalf("client: loadkeys: %s", err)
	}

	// Load root CA
	caCert, err := os.ReadFile(rootCAcert)
	if err != nil {
		log.Fatalf("client: read ca cert: %s", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,   //nolint:nosnakecase
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, //nolint:nosnakecase
		},
	}

	return config, nil
}

func (c *Client) GenerateTraffic() error {
	if c.longLived {
		if err := c.generateLongLivedConnections(); err != nil {
			return err
		}
	} else {
		if err := c.generateShortLivedConnections(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) generateShortLivedConnections() error {
	for i := 0; i < c.numRequests; i++ {
		if err := c.sendRequest(i, shortLivedDelayInSeconds); err != nil {
			return err
		}

		// Evenly distribute requests within 1 sec.
		//if c.numRequests > 1 {
		//	time.Sleep(1 * time.Second / time.Duration(c.numRequests-1))
		//}

	}

	return nil
}

func (c *Client) generateLongLivedConnections() error {
	var wg sync.WaitGroup

	wg.Add(c.numRequests)

	for i := 0; i < c.numRequests; i++ {
		go func(connID int) {
			defer wg.Done()

			c.sendRequest(connID, longLivedDelayInSeconds)
		}(i)
	}

	wg.Wait() // Wait for all goroutines to finish

	return nil
}

func (c *Client) sendRequest(connID int, delay time.Duration) error {
	conn, err := tls.Dial("tcp", c.serverAddr, c.tlsConfig)
	if err != nil {
		c.logger.Errorf("client: dial: %s", err)

		return err
	}

	defer conn.Close()

	c.logger.Infof("client: connected to: %s", conn.RemoteAddr())

	clientMsg := "Hello from Client!"

	_, err = conn.Write([]byte(clientMsg))
	if err != nil {
		c.logger.Errorf("ConnID: %d, Write to server failed: %s", connID, err.Error())

		return err
	}

	c.logger.Info("[write to server] = ", clientMsg)

	if delay > 0 {
		time.Sleep(delay)
	}

	reply := make([]byte, bytesToWrite)

	_, err = conn.Read(reply)
	if err != nil {
		c.logger.Errorf("ConnID: %d, Read from server failed: %s", connID, err.Error())

		return err
	}

	c.logger.Info("[reply from server] = ", string(reply))

	return nil
}
