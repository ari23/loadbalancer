package loadbalance

import (
	"context"
	"time"
)

// HealthCheck performs a health check on an upstream server.
func HealthCheck(
	ctx context.Context,
	server UpstreamServerInterface,
	dialer NetDialerInterface,
) (bool, error) {
	if dialer == nil {
		return false, ErrDialerIsNil
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	failureCount := 0
	timeout := dialer.GetTimeout()
	retry := dialer.GetRetryLimit()
	serverAddress := server.GetAddress()

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()

			return false, nil
		case <-ticker.C:
			_, err := dialer.DialTimeout(
				"tcp", serverAddress, timeout)
			// defer conn.Close()

			if err != nil {
				failureCount++

				if failureCount >= retry {
					server.SetHealthy(false)

					return false, ErrHealthCheckFailedAfterRetry
				}
			} else {
				server.SetHealthy(true)

				failureCount = 0
			}
		}
	}
}
