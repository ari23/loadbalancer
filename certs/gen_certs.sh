#!/bin/bash

echo "Generating CA key and certificate"
# Generate CA key
openssl genrsa -out rootCA.key 2048

# Generate CA certificate
openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 1024 -out rootCA.pem -subj "/C=US/ST=WA/O=FooOrg, Inc./CN=foodomain.com"

echo "Generating Load Balancer key, creating CSR and generate certificate with CA cert."
# Generate Load Balancer key
openssl genrsa -out loadbalancer.key 2048
# Create CSR
openssl req -new -key loadbalancer.key -out loadbalancer.csr -config loadbalancer.cnf
# Generate certificate with CA cert
openssl x509 -req -in loadbalancer.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out loadbalancer.crt -days 500 -sha256 -extfile loadbalancer.cnf -extensions req_ext

echo "Generating ClientA's key and certificate"
# Generate ClientA's key
openssl genrsa -out clientA.key 2048
# Create CSR
openssl req -new -key clientA.key -out clientA.csr -subj "/C=US/ST=WA/O=BarOrg, Inc./CN=clientA.bardomain.com"
# Generate certificate with CA cert
openssl x509 -req -in clientA.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out clientA.crt -days 500 -sha256


