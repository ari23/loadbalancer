package loadbalance_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ari23/loadbalancer/lib/loadbalance"
	"github.com/ari23/loadbalancer/tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckServerNotReachableRetryLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server := mocks.NewMockUpstreamServerInterface(ctrl)
	dialer := mocks.NewMockNetDialerInterface(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	// Setup mocks
	serverAddress := "192.168.1.1:8081"
	server.EXPECT().GetAddress().Return(serverAddress).AnyTimes()
	dialer.EXPECT().DialTimeout("tcp", serverAddress, 2*time.Second).Return(nil, errors.New("connection error")).AnyTimes()
	dialer.EXPECT().GetTimeout().Return(2 * time.Second).Times(1)
	dialer.EXPECT().GetRetryLimit().Return(1).Times(1)
	server.EXPECT().SetHealthy(false).Times(1)

	// Call HealthCheck
	_, err := loadbalance.HealthCheck(ctx, server, dialer)

	// Assert
	assert.Error(t, err)
}

func TestHealthCheckServerReachable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server := mocks.NewMockUpstreamServerInterface(ctrl)
	dialer := mocks.NewMockNetDialerInterface(ctrl)
	mockConn := mocks.NewMockConn(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	// Setup mocks
	serverAddress := "192.168.1.1:8081"
	server.EXPECT().GetAddress().Return(serverAddress).AnyTimes()
	dialer.EXPECT().DialTimeout("tcp", serverAddress, 2*time.Second).Return(mockConn, nil).AnyTimes()
	dialer.EXPECT().GetTimeout().Return(2 * time.Second).Times(1)
	dialer.EXPECT().GetRetryLimit().Return(1).Times(1)
	server.EXPECT().SetHealthy(true).Times(1)

	// Call HealthCheck
	_, err := loadbalance.HealthCheck(ctx, server, dialer)

	assert.NoError(t, err)
}
