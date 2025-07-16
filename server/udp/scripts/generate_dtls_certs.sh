#!/bin/bash

# 生成DTLS测试证书脚本

echo "Generating DTLS test certificates..."

# 生成私钥
openssl genrsa -out server.key 2048

# 生成自签名证书
openssl req -new -x509 -key server.key -out server.crt -days 365 -subj "/C=US/ST=CA/L=San Francisco/O=Test/OU=Test/CN=localhost"

echo "Generated server.key and server.crt"
echo "These are self-signed certificates for testing only!"
echo "Do not use in production!"

# 设置权限
chmod 600 server.key
chmod 644 server.crt

echo "Certificate generation complete."