@echo off
REM Generate TypeScript types from OpenAPI v3 specification
REM Usage: scripts\gen-types.bat

setlocal enabledelayedexpansion

set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%.."
set "OPENAPI_FILE=%PROJECT_ROOT%\docs\api\openapi.yaml"
set "OUTPUT_FILE=%PROJECT_ROOT%\web\src\types\api.ts"

echo --- Generating TypeScript types from OpenAPI v3 spec...

if not exist "%OPENAPI_FILE%" (
    echo ERROR: OpenAPI file not found at: %OPENAPI_FILE%
    echo        Run 'make gen-proto' first to generate OpenAPI documentation.
    exit /b 1
)

cd /d "%PROJECT_ROOT%\web"

if not exist "node_modules\openapi-typescript" (
    echo INFO: Installing openapi-typescript...
    call npm install --save-dev openapi-typescript
)

echo --- Generating TypeScript types from OpenAPI v3...
npx openapi-typescript "..\docs\api\openapi.yaml" -o "src\types\api.ts"

echo --- TypeScript types generated at: %OUTPUT_FILE%

endlocal
