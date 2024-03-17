package loadbalance

func GetNextUpstreamServer(upstreamServers []UpstreamServerInterface) (UpstreamServerInterface, error) {
	// TODO: Implement the algorithm.
	// For now just return the first one.
	return upstreamServers[0], nil
}
