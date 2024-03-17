package loadbalancer

import (
	"errors"
	"fmt"
)

var (
	ErrMTLSParamsMissing = errors.New("certificate, private key and CA certificate are required for mTLS")

	ErrNotTLSConnection = errors.New("not a TLS connection")

	ErrClientNameNotFound = errors.New("client name not found")

	ErrNoAuthorizedClients = errors.New("no authorized clients")

	ErrNilConnection = errors.New("nil connection")

	ErrNilClientConfig = errors.New("nil client config")

	ErrNilTargetGroupsStore = errors.New("nil target groups store")
)

func ErrLoadingKeyPair(err error) error {
	return fmt.Errorf("failed to load key pair: %v", err)
}

func ErrLoadingCACert(err error) error {
	return fmt.Errorf("failed to load CA Cert: %v", err)
}

func ErrTLSHandshakeFailed(err error) error {
	return fmt.Errorf("TLS handshake failed: %v", err)
}

func ErrClientNotAuthorized(clientName string) error {
	return fmt.Errorf("client %s is not authorized", clientName)
}

func ErrTargetGroupNotFound(targetGroupName string) error {
	return fmt.Errorf("target group %s not found", targetGroupName)
}
