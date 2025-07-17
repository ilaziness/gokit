@echo off
REM 生成QUIC服务器证书脚本 (Windows)

setlocal enabledelayedexpansion

set "CERT_DIR=%~dp0..\example"
if not exist "%CERT_DIR%" mkdir "%CERT_DIR%"

echo Generating QUIC server certificates...

REM 检查OpenSSL是否可用
where openssl >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: OpenSSL is not installed or not in PATH
    echo Please install OpenSSL and add it to your PATH
    echo You can download it from: https://slproweb.com/products/Win32OpenSSL.html
    pause
    exit /b 1
)

REM 生成私钥
openssl genrsa -out "%CERT_DIR%\server.key" 2048
if %errorlevel% neq 0 (
    echo Error generating private key
    pause
    exit /b 1
)

REM 创建临时配置文件
set "TEMP_CONFIG=%TEMP%\openssl_config.tmp"
(
echo [req]
echo distinguished_name = req_distinguished_name
echo req_extensions = v3_req
echo prompt = no
echo.
echo [req_distinguished_name]
echo C = US
echo ST = CA
echo L = San Francisco
echo O = Test
echo OU = Test
echo CN = localhost
echo.
echo [v3_req]
echo keyUsage = keyEncipherment, dataEncipherment
echo extendedKeyUsage = serverAuth
echo subjectAltName = @alt_names
echo.
echo [alt_names]
echo DNS.1 = localhost
echo DNS.2 = *.localhost
echo IP.1 = 127.0.0.1
echo IP.2 = ::1
) > "%TEMP_CONFIG%"

REM 生成证书签名请求
openssl req -new -key "%CERT_DIR%\server.key" -out "%CERT_DIR%\server.csr" -config "%TEMP_CONFIG%"
if %errorlevel% neq 0 (
    echo Error generating certificate signing request
    del "%TEMP_CONFIG%" 2>nul
    pause
    exit /b 1
)

REM 生成自签名证书
openssl x509 -req -days 365 -in "%CERT_DIR%\server.csr" -signkey "%CERT_DIR%\server.key" -out "%CERT_DIR%\server.crt" -extensions v3_req -extfile "%TEMP_CONFIG%"
if %errorlevel% neq 0 (
    echo Error generating certificate
    del "%TEMP_CONFIG%" 2>nul
    del "%CERT_DIR%\server.csr" 2>nul
    pause
    exit /b 1
)

REM 清理临时文件
del "%TEMP_CONFIG%" 2>nul
del "%CERT_DIR%\server.csr" 2>nul

echo.
echo Certificates generated successfully:
echo   Private key: %CERT_DIR%\server.key
echo   Certificate: %CERT_DIR%\server.crt
echo.
echo Note: These are self-signed certificates for testing only.
echo For production use, obtain certificates from a trusted CA.
echo.
pause