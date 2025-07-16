@echo off
REM 生成用于测试的自签名证书 (Windows版本)
REM 注意：这些证书仅用于开发和测试，不要在生产环境中使用
REM 需要安装OpenSSL for Windows

echo Generating self-signed certificates for UDP server testing...

REM 创建证书目录
if not exist certs mkdir certs

REM 生成私钥
openssl genrsa -out certs/server.key 2048

REM 生成证书签名请求
openssl req -new -key certs/server.key -out certs/server.csr -subj "/C=CN/ST=Beijing/L=Beijing/O=Test/OU=Test/CN=localhost"

REM 生成自签名证书
openssl x509 -req -days 365 -in certs/server.csr -signkey certs/server.key -out certs/server.crt

REM 清理临时文件
del certs\server.csr

echo.
echo Certificates generated in certs\ directory:
echo   - certs\server.key (private key)
echo   - certs\server.crt (certificate)
echo.
echo To use TLS in your UDP server, set:
echo   CertFile: "certs\server.crt"
echo   KeyFile:  "certs\server.key"

pause