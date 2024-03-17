package ratelimit

import "time"

func RequestAllowed(clientInfo ClientInfoInterface) bool {
	if now := time.Now(); now.Sub(clientInfo.GetLastWindow()) > time.Second {
		clientInfo.SetLastWindow(now)
	}

	if clientInfo.GetRequestCount() >= clientInfo.GetMaxRequestsPerWindow() {
		return false
	}

	clientInfo.IncrementRequestCount()

	return true
}
