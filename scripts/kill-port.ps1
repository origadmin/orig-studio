param(
    [Parameter(Mandatory=$true)]
    [int]$Port
)

$connections = netstat -ano | Select-String ":$Port\s" | Select-String "LISTENING"
if (-not $connections) {
    Write-Host "No process found on port $Port"
    exit 0
}

$pids = @{}
foreach ($conn in $connections) {
    $fields = $conn.Line.Trim() -split '\s+'
    $pid = [int]$fields[-1]
    if ($pid -ne 0 -and $pid -ne $PID) {
        $pids[$pid] = $true
    }
}

if ($pids.Count -eq 0) {
    Write-Host "No process found on port $Port"
    exit 0
}

foreach ($pid in $pids.Keys) {
    Write-Host "Killing process $pid on port $Port..."
    Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
    if ($?) {
        Write-Host "Killed process $pid"
    } else {
        Write-Host "Failed to kill process $pid"
    }
}

Write-Host "Done."
