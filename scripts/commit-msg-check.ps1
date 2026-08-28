#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Commit message gate: English-only + Conventional Commits (Angular style).

.DESCRIPTION
    Enforced on the whole message, not just the subject:
      1. subject must be English (no CJK)
      2. subject must match <type>(<scope>): <subject>
      3. subject length <= 100
      4. body must be English too -- CJK is only tolerated inside a file path,
         because paths must keep pointing at the real file on disk

    Why rule 4 exists: the previous version only inspected the first line,
    which let 43 commits accumulate Chinese inside their bodies.
#>
param(
    [Parameter(Mandatory = $true)]
    [string]$CommitMsgFile
)

$ErrorActionPreference = "Stop"

$commitMessage = Get-Content $CommitMsgFile -Raw -Encoding UTF8
if ($null -eq $commitMessage) { $commitMessage = "" }

# CJK ranges: Han, extension A, compatibility, kana, hangul
$Cjk = '[\u4e00-\u9fff\u3400-\u4dbf\uf900-\ufaff\u3040-\u309f\u30a0-\u30ff\uac00-\ud7af]'
# A file path may legitimately contain Chinese; translating it breaks the link.
$Path = '(?:[\w\u4e00-\u9fff.\-]+/)*[\w\u4e00-\u9fff.\-]+\.(?:md|ts|tsx|js|jsx|go|json|yml|yaml|proto|sql|sh|ps1|html|css)'
$Conventional = '^(feat|fix|docs|style|refactor|perf|test|chore|build|ci|revert)(\([a-z0-9_.\/-]+\))?: .+'
$Revert = '^Revert ".+"'

$lines = $commitMessage -split "`n"
$firstLine = $lines[0].Trim()

$errors = @()

if ([string]::IsNullOrWhiteSpace($firstLine)) {
    Write-Host ""
    Write-Host "================================================" -ForegroundColor Red
    Write-Host "  COMMIT REJECTED: Empty commit message" -ForegroundColor Red
    Write-Host "================================================" -ForegroundColor Red
    exit 1
}

# --- subject -------------------------------------------------------------
if ($firstLine -match $Cjk) {
    $errors += "SUBJECT MUST BE ENGLISH: CJK characters detected.`n   Found: `"$firstLine`""
}

if ($firstLine.Length -gt 100) {
    $errors += "SUBJECT TOO LONG: $($firstLine.Length) characters (max 100)."
}

if (($firstLine -notmatch $Conventional) -and ($firstLine -notmatch $Revert)) {
    $errors += "NOT CONVENTIONAL COMMITS FORMAT: `"$firstLine`"`n   Expected: <type>(<scope>): <subject>`n   Types: feat, fix, docs, style, refactor, perf, test, chore, build, ci, revert`n   Example: fix(media): correct route registration order and Nginx proxy config"
}

# --- body (the gap the old hook left wide open) --------------------------
for ($i = 1; $i -lt $lines.Count; $i++) {
    $line = $lines[$i]
    # Skip comment lines; git strips them before the commit is stored.
    if ($line.StartsWith("#")) { continue }
    $stripped = $line -replace $Path, ''
    if ($stripped -match $Cjk) {
        $preview = $line.Trim()
        if ($preview.Length -gt 90) { $preview = $preview.Substring(0, 90) + "..." }
        $errors += "BODY MUST BE ENGLISH: CJK outside a file path on line $($i + 1).`n   Found: `"$preview`"`n   Chinese is tolerated only inside file paths, which must keep their real names."
        break
    }
}

if ($errors.Count -gt 0) {
    Write-Host ""
    Write-Host "================================================" -ForegroundColor Red
    Write-Host "  COMMIT REJECTED: Invalid commit message" -ForegroundColor Red
    Write-Host "================================================" -ForegroundColor Red
    Write-Host ""
    foreach ($e in $errors) {
        Write-Host "  [ERROR] $e" -ForegroundColor Red
        Write-Host ""
    }
    Write-Host "Write commit messages in English, Conventional Commits format." -ForegroundColor Yellow
    Write-Host "Format: <type>(<scope>): <subject>" -ForegroundColor Yellow
    exit 1
}

Write-Host "  [OK] Commit message check passed (subject + body)" -ForegroundColor Green
exit 0
