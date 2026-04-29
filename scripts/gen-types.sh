#!/bin/bash
# Generate TypeScript types from OpenAPI specification
# Usage: ./scripts/gen-types.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
OPENAPI_FILE="$PROJECT_ROOT/docs/api/openapi.yaml" # Changed from openapi.swagger.json to openapi.yaml
OUTPUT_FILE="$PROJECT_ROOT/web/src/types/api.ts"

echo "🔷 Generating TypeScript types from OpenAPI spec..."

if [ ! -f "$OPENAPI_FILE" ]; then
    echo "❌ OpenAPI file not found at: $OPENAPI_FILE"
    echo "   Run 'make gen-proto' first to generate OpenAPI documentation."
    exit 1
fi

cd "$PROJECT_ROOT/web"

# Check if openapi-typescript is installed
if ! command -v openapi-typescript &> /dev/null; then
    if [ ! -d "node_modules/openapi-typescript" ]; then
        echo "📦 Installing openapi-typescript..."
        npm install --save-dev openapi-typescript
    fi
fi

# Generate TypeScript types
npx openapi-typescript "../docs/api/openapi.yaml" -o "src/types/api.ts" # Changed from openapi.swagger.json to openapi.yaml

echo "✅ TypeScript types generated at: $OUTPUT_FILE"
