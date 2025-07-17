#!/bin/bash

# 生成QUIC服务器证书脚本

set -e

CERT_DIR="$(dirname "$0")/../example"
mkdir -p "$CERT_DIR"

echo "Generating QUIC server certificates..."

# 生成私钥
openssl genrsa -out "$CERT_DIR/server.key" 2048

# 生成证书签名请求
openssl req -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" -subj "/C=US/ST=CA/L=San Francisco/O=Test/OU=Test/CN=localhost"

# 生成自签名证书
openssl x509 -req -days 365 -in "$CERT_DIR/server.csr" -signkey "$CERT_DIR/server.key" -out "$CERT_DIR/server.crt" -extensions v3_req -extfile <(
cat <<EOF
[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF
)

# 清理临时文件
rm "$CERT_DIR/server.csr"

echo "Certificates generated successfully:"
echo "  Private key: $CERT_DIR/server.key"
echo "  Certificate: $CERT_DIR/server.crt"
echo ""
echo "Note: These are self-signed certificates for testing only."
echo "For production use, obtain certificates from a trusted CA."