#!/bin/bash
echo "📣 about to run buf dep update"
buf dep update api/proto
echo "📣 about to run buf lint"
buf lint api/proto
echo "📣 about to run buf generate"
buf generate api/proto

echo "📣 copying generated OpenAPI JSON to documentation and frontend folders"
# Rename the generated JSON file that defaults to the weird extension
mv api/openapi/thing.swagger.json api/openapi/thing.json

# Copy to docs
rsync -av api/openapi/thing.json docs/thing.json
# Copy to frontend assets
rsync -av api/openapi/thing.json cmd/goCloudK8sThingServer/goCloudK8sThingFront/public/oapidoc/thing.json

echo "✅ Generation complete."
