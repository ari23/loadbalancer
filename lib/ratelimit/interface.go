package ratelimit

import "time"

type ClientInfoInterface interface {
	GetClientID() string
	GetAllowedTargetGroup() string
	GetMaxConnections() int
	GetConnections() int
	GetMaxRequestsPerWindow() int
	GetLastWindow() time.Time
	GetRequestCount() int
	SetLastWindow(now time.Time)
	IncrementRequestCount()
}
