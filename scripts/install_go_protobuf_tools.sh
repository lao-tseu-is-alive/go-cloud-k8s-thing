#!/bin/bash
go install github.com/bufbuild/buf/cmd/buf@v1.61.0
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
buf config init
