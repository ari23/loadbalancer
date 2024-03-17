# Network Load Balancer

This repo provides a very basic implementation of a Network Load Balancer in Go.

## Scope

* Allow clients to securely communicate with Load balancer.
* Allow clients to access only authorized upstream servers.
* Provide per-client rate limiting.
* Maintain a list of active upstream servers.
* Load balance based on least number of connections to upstream servers.
* Provide a simple bootstrap config file.
* Provide a simple CLI to collect stats.

## Design Considerations

### 1. Client Communication

The load balancer requires clients to communicate via [mTLS](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/). This way the load balancer ensures only authenticated clients are allowed to establish connections. Also this provides protection from Man-in-the-middle attacks thus strenghtning security posture for the system.

For the initial implementation, TLS certs will be loaded from local directory. But in future, certs should be rotated and pushed to the load balancer (either via k8 secrets or shared vol).

NOTE: For self-signed certs, both, the client and the load balancer needs to have the root CA cert in its trust store. This is because the client needs to trust the load balancer and the load balancer needs to trust the client.

Sample process in generating self-signed certs (helper script will be provided to generate these):

Generate Root CA's private key and certificate.

```bash
openssl genrsa -out rootCA.key 4096
openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 1024 -out rootCA.pem
```

Generate Load Balancer's Certificate, create CSR and sign with root CA's cert.

```bash
openssl genrsa -out loadbalancer.key 2048
openssl req -new -key loadbalancer.key -out loadbalancer.csr
openssl x509 -req -in loadbalancer.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out loadbalancer.crt -days 500 -sha256
```

Generate Client's Certificate, create CSR and sign with root CA's cert.

```bash
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr
openssl x509 -req -in client.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out client.crt -days 500 -sha256
```

### 2. Client Authorization

To ensure clients are only allowed to access upstream services that they are authorized for, ACL/Policy eval system is used. In production system this might be achieved by running something like [OPA](https://www.openpolicyagent.org/docs/latest/).

But for the sake of simplicity, this load balancer maintains a simple map of Client name (extracted from client's TLS certificate - clientCert.Subject.CommonName) and list of upstream target groups.

Note: A target group is a collection of servers e.g Financial Services target group will include a list of Finance backend services.

### 3. Client Rate Limiting

Rate Limiting ensures fair share across clients by controlling the rate of resource consumption.
e.g a client can not send more than 20 requests per sec.

Some of the popular rate limiting algorithms are:

* Token Bucket
* Leaky Bucket
* Fixed Window Counter
* Sliding Window Counter
* Sliding Window Log

For the initial implementation, Fixed Window Counter is chosen. Its easy to understand and reason about.
In Fixed Window Counter, the timeline is divided into "fixed-sized windows" e.g 1 sec window. And a counter is assigned to each Window. Each request increments the counter for that window. If the max threshold is reached, then the request (and any subsequent requests) are dropped within that window and a TCP RST is sent back.

Although it does come with a major drawback: Spike in traffic at the edges of a window can result in more requests than the allowed quota for a given window. Sliding Window is usually the recommended approach to counter this drawback.

Also, as a side note, when threshold is reached, dropping is just one option (and the easiest one). The algorithm can be updated to take other actions such as throttle or shaping.

### 4. Maintain Active Upstream Services

To ensure client requests are not timing out, the load balancer needs to constantly monitor health of upstream servers. Health checks can be active or passive. In Active mode, the load balancer actively sends probe to check if the service is healthy. In Passive mode, it keeps track of upstream service by monitoring client's request.

For the initial implementation, Active mode is chosen. The load balancer will periodically (configurable e.g. 5s) send TCP probe (a simple TCP handshake) to check if the upstream server is healthy. If a response is not received for 5s, it will tag the server "unhealthy" and remove it from the pool of "active upstream" servers.
Also in order to avoid overwhelming upstream server during recovery, an exponential backoff is recommended. But for the sake of simplicity, exponential backoff is left out. The load balancer will wait for 3 active probes to tag the server "healthy" and bring it back in the "active upstream rotation".

### 5. Load Balance Algorithm

Load balancing algorithms define the logic to distribute traffic across upstream servers.
Some of the popular algorithms are:

* Static
    1. Round Robin
    2. Weighted Round Robin
    3. IP Hash
* Dynamic
    1. Least Connection
    2. Weighted Least Connection
    3. Resource based

For the initial implementation, Least Connection algorithm is chosen.

The Least Connection algorithm:

1. It maintains a count of current active (or open) connections for each server in the pool of available servers.
2. When a new request arrives, the load balancer scans its list of servers and forwards the request to the server with the fewest active connections at that moment. NOTE: A min heap can be used to optimize this, but for this challenge, a simple linear scan is used.
3. As connections are established or terminated, the load balancer updates its count of active connections for each server, ensuring that subsequent routing decisions are based on the latest data.

### 6. Bootstrap config

A simple config file that constains important information for bootstapping load balancer. E.g:

1. target group to upstream server map
2. certificate location, log level etc.
3. IP and Port for load balancer.

In the future, this can be enriched with more details such as:

1. Use Round robin algorithm for Financial services target group.
2. Use Token Bucket for Client A and Leaky Bucker for Client B.

e.g:

```yaml
# Load Balancer Configuration
listenAddress: "0.0.0.0:8080"
tlsParams:
  certificate: "certs/loadbalancer.crt"
  privateKey: "certs/loadbalancer.key"
  caCert: "certs/rootCA.pem"
logLevel: "info"

# Target Groups with Upstream Servers
targetGroups:
  - name: "FrontEndService" # HTTP service
    upstreamServers:
      - "127.0.0.1:8081"
      - "127.0.0.1:8082"
  - name: "DBService" # non HTTP service
    upstreamServers:
      - "127.0.0.1:8085"
      - "127.0.0.1:8086"

# Client to Target Group Mapping
clients:
  - clientId: "clientA.bardomain.com"
    allowedTargetGroup: "FrontEndService"
    requests_per_second: 10
  - clientId: "clientB.bardomain.com"
    allowedTargetGroup: "DBService"
    requests_per_second: 5

```

### 7. CLI

A CLI tool to interact with the load balancer to perform some of the following tasks:

1. Add a new upstream server to a target group.

    ```bash
    lbctl add-upstream-server --target-group "FrontEndService" --server "frontend-server-3.example.com:8080"
    ```

2. Remove a client.

    ```bash
    lbctl remove-client --client "ClientA"
    ```

3. Toggle log level for debugging.

    ```bash
    lbctl set-log-level --level "debug"
    ```

4. Get current number of active connections per upstream server in a target group.

    ```bash
    lbctl get-active-connections --target-group "FrontEndService"
    ```

5. Get number of dropped requests for a given client.

    ```bash
    lbctl get-dropped-requests --client "ClientA"
    ```

In the future, this can be enriched with more details such as:

1. Client audit (How many long lived connections per client, Auth failed attempts, Too many requests etc).

    ```bash
    lbctl get-dropped-requests --client "ClientA"
    ```

2. Server audit (How often a given upstream server lot connectivity, Latency per request, etc).

    ```bash
    lbctl get-failed-connections --server "frontend-server-1.example.com:8080"
    ```

## Load Balancer Implementation Details

### Assumptions/Pre-reqs

1. Load balancer and client certs are already generated and distributed accordingly.
2. Clients have knowledge on how to connect to the load balancer (either via DNS or static config).
3. Upstream servers are reachable from the load balancer (i.e. IP connectivity exists).
4. Target groups are well defined and mapped to upstream servers.
5. Client to upstream server authorization is pre-configured.

### High level traffic flow

1. The network load balancer is configured to listen on a specific port (e.g., 8080) on all interfaces (0.0.0.0) or restricted to the WAN IP of the node.
2. Client A wishes to access a service (e.g., a database, an SSH server, etc.) through the load balancer.
3. The request from Client A reaches the network load balancer, either through DNS resolution or static configuration.
4. For the initial connection setup, Client A and the network load balancer perform an mTLS handshake, where both parties exchange and authenticate their certificates. This step ensures that all subsequent traffic for this session is encrypted and authenticated, enhancing security for non-HTTP protocols.
5. Once the mTLS session is established, the network load balancer is ready to route Client A's traffic.
6. The network load balancer checks if Client A is authorized to access any target group, which is a collection of upstream servers providing the requested service. This check involves a lookup in an internal map of Client to Target Group. If no matching target group is found, the load balancer terminates the connection, effectively denying access.
7. If Client A is authorized, the load balancer then checks if the client is within its rate limit for new connections or requests. If the client exceeds this limit, the load balancer terminates the connection to enforce rate limiting.
8. If Client A is authorized and within the rate limit, the network load balancer selects an upstream server from the target group (determined by parsing port an d protocol e.g. if TCP:80, send to application target group) that has the least number of active connections. This decision is based on the "Least Connection" load balancing algorithm, aiming to distribute the load evenly across the available servers.
9. Client A's traffic is then forwarded to the selected upstream server. The server processes the request (e.g., a database query, an SSH command, etc.) and sends the response back through the load balancer.
10. The network load balancer forwards the response from the upstream server back to Client A, completing the request-response cycle.

### Security Considerations

1. The load balancer is designed to use mTLS to accept client connections. This ensures that only authenticated clients are allowed to establish connections. Also this provides protection from Man-in-the-middle attacks thus strenghtning security posture for the system. TLS1.2 or higher with cipher suites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 is chosen for this implementation. ECDHE ensures forward secrecy and AES_128_GCM provides strong encryption.

2. In production deployment, the certificates should be properly managed by using trusted Certificate Authorities (CAs) and rotated frequently before they expire and securely store private keys.

3. Access to the upstream servers should be hidden from clients directly.

4. The load balancer uses a simple Authz system to ensure clients are only allowed to access upstream services that they are authorized for. This is not recommended for production systems as a compromised client can still access the upstream services until their certificate is rotated. To remedy this, per flow based Authz is recommended, every new flow of a client should be re-evaluated for access.

5. Use more secure rate limiting algorithms such as Sliding Window Counter + Connection Pool to protect against DDoS attacks.

6. The load balancer uses a simple health check system to ensure upstream servers are healthy. This is not recommended for production systems as it can lead to false positives. To remedy this, a more robust health check systems such as [Envoy's Health Check](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/health_checking#arch-overview-health-checking) is recommended.

## Building

To build the load balancer, run the following command:

```bash
make build_loadbalancer
```

## Running

The load balancer depends on a bootstrap config to start. The bootstrap config is a minimal config that contains the load balancer's cert and key, the root CA's cert and key and the load balancer's listening address and port.
There is a sample bootstrap.yaml is provided in pkg/loadbalancer. Edit it accordingly and run the following command:

```bash
./bin/loadbalancer start --config pkg/loadbalancer/bootstrap.yaml
```

## Testing

There are some sample scripts provided to test the load balancer with different scenarios.

### Sample Test Scenarios

#### Test Scenario 1: Client A connects to FrontEndService and sends 5 requests in 1 second

* In Terminal 1: Run the load balancer

  ```bash
    # Build loadbalancer
    make build_loadbalancer

    # Run loadbalancer
    ./bin/loadbalancer start --config pkg/loadbalancer/bootstrap.yaml
  ```

* In Terminal 2: Run 2 upstream servers for FrontEndService

  ```bash
    # Build upstream server
    make build_server

    # Run server(s)
    ./tests/integration/server/server -num_servers 2 -ip 127.0.0.1 -start_port 8081
  ```

* In Terminal 3: Run client and issue 5 requests which is within client's rate limit of 10 max requests.

  ```bash
    # Build client
    make build_client

    # Run a single client (e.g. client A)
    ./tests/integration/client/client -requests 5 -server 127.0.0.1:8080
  ```
