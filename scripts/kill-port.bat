@echo off
setlocal enabledelayedexpansion

if "%~1"=="" (
    echo Usage: kill-port ^<port^>
    echo Example: kill-port 18080
    exit /b 1
)

set "PORT=%~1"

echo Checking port %PORT%...

for /f "tokens=2,5" %%a in ('netstat -ano ^| findstr ":%PORT% " ^| findstr "LISTENING"') do (
    set "PID=%%b"
    if defined PID (
        echo Found process !PID! on port %PORT%, killing...
        taskkill /PID !PID! /F >nul 2>&1
        if !errorlevel! equ 0 (
            echo Killed process !PID!.
        ) else (
            echo Failed to kill process !PID!.
        )
    )
)

echo Done.
