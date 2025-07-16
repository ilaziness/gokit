#!/bin/bash

# 生成用于测试的自签名证书
# 注意：这些证书仅用于开发和测试，不要在生产环境中使用

echo "Generating self-signed certificates for UDP server testing..."

# 创建证书目录
mkdir -p certs

# 生成私钥
openssl genrsa -out certs/server.key 2048

# 生成证书签名请求
openssl req -new -key certs/server.key -out certs/server.csr -subj "/C=CN/ST=Beijing/L=Beijing/O=Test/OU=Test/CN=localhost"

# 生成自签名证书
openssl x509 -req -days 365 -in certs/server.csr -signkey certs/server.key -out certs/server.crt

# 清理临时文件
rm certs/server.csr

echo "Certificates generated in certs/ directory:"
echo "  - certs/server.key (private key)"
echo "  - certs/server.crt (certificate)"
echo ""
echo "To use TLS in your UDP server, set:"
echo "  CertFile: \"certs/server.crt\""
echo "  KeyFile:  \"certs/server.key\""