@echo off
SETLOCAL

set _TARGETS=build

set VERSION=1.1.0

REM Set build number from git commit hash
for /f %%i in ('git rev-parse HEAD') do set BUILD=%%i

if [%1]==[] goto usage

set LDFLAGS=-ldflags "-X main.version=%VERSION% -X main.build=%BUILD%"

goto %1

:build
    set GOARCH=amd64

    go build %LDFLAGS%

    goto :clean

:usage
	echo usage: make [target]
	echo.
	echo target is one of {%_TARGETS%}.
	exit /b 2
	goto :eof

:clean
    set _TARGETS=
    set VERSION=
    set BUILD=
    set LDFLAGS=
    set GOARCH=

    goto :EOF
