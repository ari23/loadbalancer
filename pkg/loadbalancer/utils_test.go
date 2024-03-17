package loadbalancer_test

import (
	"net"
	"testing"

	"github.com/ari23/loadbalancer/pkg/loadbalancer"
	"github.com/ari23/loadbalancer/tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetClientConfigFromConnNotTLS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks.NewMockConn(ctrl)

	mockConn.EXPECT().RemoteAddr().AnyTimes().Return(&net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 8080})

	authorizedClientsStore := &loadbalancer.ClientsStore{}

	_, err := loadbalancer.GetClientConfigFromConn(mockConn, authorizedClientsStore)

	assert.Error(t, err, "Expected error when getting client config from non-TLS connection")
}
