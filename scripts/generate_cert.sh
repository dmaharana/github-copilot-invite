#!/bin/bash

# Directory for certificates
CERT_DIR="../certs"
CERT_FILE="$CERT_DIR/server.crt"
KEY_FILE="$CERT_DIR/server.key"

# Create certs directory if it doesn't exist
mkdir -p $CERT_DIR

# Generate self-signed certificate
openssl req -x509 \
    -newkey rsa:4096 \
    -keyout $KEY_FILE \
    -out $CERT_FILE \
    -days 365 \
    -nodes \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

echo "Generated certificates:"
echo "Certificate: $CERT_FILE"
echo "Private Key: $KEY_FILE"
