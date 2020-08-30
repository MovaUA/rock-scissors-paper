#!/bin/sh

set -ex

dir=$(dirname $0)

# protoc \
#   --proto_path "${dir}" \
#   --go_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:"${dir}" \
#   --go-grpc_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:"${dir}" \
#   --go_opt=paths=source_relative \
#   --go-grpc_opt=paths=source_relative \
#   "${dir}"/ssp.proto

protoc \
  --proto_path "${dir}" \
  --go_out="${dir}" \
  --go-grpc_out="${dir}" \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  "${dir}"/rsp.proto
