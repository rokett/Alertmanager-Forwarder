@echo off
SETLOCAL

set APP=Alertmanager-Forwarder
set VERSION=1.1.0
set BINARY-WINDOWS-X64=%APP%_%VERSION%_Windows_amd64.exe
set BINARY-LINUX=%APP%_%VERSION%_Linux_amd64

REM Set build number from git commit hash
for /f %%i in ('git rev-parse HEAD') do set BUILD=%%i

set LDFLAGS=-ldflags "-X main.version=%VERSION% -X main.build=%BUILD% -s -w"

goto build

:build
    echo "=== Building Windows x64 ==="
    set GOOS=windows
    set GOARCH=amd64

    go build -mod=vendor -o %BINARY-WINDOWS-X64% %LDFLAGS%

    echo "=== Building Linux x64 ==="
    set GOOS=linux
    set GOARCH=amd64

    go build -mod=vendor -o %BINARY-LINUX% %LDFLAGS%

    goto :finalise

:finalise
    set GOOS=
    set GOARCH=

    goto :EOF
