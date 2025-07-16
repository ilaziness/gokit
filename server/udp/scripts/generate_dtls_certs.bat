@echo off
REM 生成DTLS测试证书脚本 (Windows)

echo Generating DTLS test certificates...

REM 检查是否安装了OpenSSL
where openssl >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo OpenSSL not found in PATH. Please install OpenSSL first.
    echo You can download it from: https://slproweb.com/products/Win32OpenSSL.html
    pause
    exit /b 1
)

REM 生成私钥
openssl genrsa -out server.key 2048

REM 生成自签名证书
openssl req -new -x509 -key server.key -out server.crt -days 365 -subj "/C=US/ST=CA/L=San Francisco/O=Test/OU=Test/CN=localhost"

echo Generated server.key and server.crt
echo These are self-signed certificates for testing only!
echo Do not use in production!

echo Certificate generation complete.
pause