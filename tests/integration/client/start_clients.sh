#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <number_of_requests>"
    exit 1
fi

num_requests=$1

# Server details
server_address="127.0.0.1"
server_port="8080"
message="Hello from Client!"

ssl_cert="certs/clientA.crt"
ssl_key="certs/clientA.key"

# Loop to create the specified number of requests
for (( i=1; i<=num_requests; i++ ))
do
    echo -n "$message" | ncat --ssl --ssl-cert "$ssl_cert" --ssl-key "$ssl_key" "$server_address" "$server_port"
    echo "Request $i sent"
done

echo "Completed sending $num_requests requests."