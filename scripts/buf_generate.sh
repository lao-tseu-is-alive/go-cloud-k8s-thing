#!/bin/bash
echo "📣 about to run buf dep update"
buf dep update api/proto
echo "📣 about to run buf lint"
buf lint api/proto
echo "📣 about to run buf generate"
buf generate api/proto
