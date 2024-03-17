package loadbalancer

import (
	"crypto/tls"
	"net"

	"github.com/ari23/loadbalancer/lib/loadbalance"
	"github.com/ari23/loadbalancer/lib/ratelimit"
)

func GetClientConfigFromConn(
	conn net.Conn,
	authorizedClientsStore *ClientsStore,
) (ratelimit.ClientInfoInterface, error) {
	if conn == nil {
		return nil, ErrNilConnection
	}

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return nil, ErrNotTLSConnection
	}

	if err := tlsConn.Handshake(); err != nil {
		return nil, ErrTLSHandshakeFailed(err)
	}

	clientName, err := GetClientName(tlsConn)
	if err != nil {
		return nil, err
	}

	clientConfig, ok := authorizedClientsStore.GetClient(clientName)
	if !ok {
		return nil, ErrClientNotAuthorized(clientName)
	}

	return clientConfig, nil
}

func GetClientName(conn *tls.Conn) (string, error) {
	state := conn.ConnectionState()
	for _, cert := range state.PeerCertificates {
		return cert.Subject.CommonName, nil
	}

	return "", ErrClientNameNotFound
}

func GetNextUpstreamServer(
	clientInfo ratelimit.ClientInfoInterface,
	targetGroupsStore *TargetGroupsStore,
) (loadbalance.UpstreamServerInterface, error) {
	if clientInfo == nil {
		return nil, ErrNilClientConfig
	}

	if targetGroupsStore == nil {
		return nil, ErrNilTargetGroupsStore
	}

	return targetGroupsStore.GetNextUpstreamServer(clientInfo.GetAllowedTargetGroup())
}
