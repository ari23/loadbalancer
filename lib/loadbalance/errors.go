package loadbalance

import "errors"

var (
	// ErrHealthCheckFailedAfterRetry is returned when health check fails after retry limit
	ErrHealthCheckFailedAfterRetry = errors.New("health check failed after retry limit")
	// ErrDialerTimeoutExceeded is returned when dialer timeout is exceeded.
	ErrDialerTimeoutExceeded = errors.New("dialer timeout exceeded")
	// ErrDialerIsNil is returned when dialer is nil.
	ErrDialerIsNil = errors.New("dialer is nil")
)
