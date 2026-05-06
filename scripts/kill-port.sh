#!/usr/bin/env bash
set -euo pipefail

if [ $# -eq 0 ]; then
    echo "Usage: kill-port <port>"
    echo "Example: kill-port 18080"
    exit 1
fi

PORT="$1"

if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    PIDS=$(netstat -ano | grep ":${PORT} " | grep "LISTENING" | awk '{print $5}' | sort -u)
    if [ -z "$PIDS" ]; then
        echo "No process found on port $PORT"
        exit 0
    fi
    for PID in $PIDS; do
        echo "Killing process $PID on port $PORT..."
        taskkill //PID "$PID" //F 2>/dev/null && echo "Killed $PID" || echo "Failed to kill $PID"
    done
elif [[ "$OSTYPE" == "darwin"* ]]; then
    PIDS=$(lsof -i ":${PORT}" -t -sTCP:LISTEN 2>/dev/null || true)
    if [ -z "$PIDS" ]; then
        echo "No process found on port $PORT"
        exit 0
    fi
    for PID in $PIDS; do
        echo "Killing process $PID on port $PORT..."
        kill -9 "$PID" 2>/dev/null && echo "Killed $PID" || echo "Failed to kill $PID"
    done
else
    PIDS=$(ss -tlnp "sport = :${PORT}" 2>/dev/null | grep -oP 'pid=\K[0-9]+' | sort -u || true)
    if [ -z "$PIDS" ]; then
        PIDS=$(lsof -i ":${PORT}" -t -sTCP:LISTEN 2>/dev/null || true)
    fi
    if [ -z "$PIDS" ]; then
        echo "No process found on port $PORT"
        exit 0
    fi
    for PID in $PIDS; do
        echo "Killing process $PID on port $PORT..."
        kill -9 "$PID" 2>/dev/null && echo "Killed $PID" || echo "Failed to kill $PID"
    done
fi

echo "Done."
