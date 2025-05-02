#!/bin/bash

PROTO_DIRS="./cmd/*/api"
PROTO_IMPORT_DIR="./third_party"
DOCS_DIR="./docs"

mkdir -p $DOCS_DIR

find $PROTO_DIRS -name "*.proto" -print0 | while IFS= read -r -d $'\0' PROTO_FILE; do
  PROTO_DIR=$(dirname "$PROTO_FILE")
  echo "Processing: $PROTO_FILE"

  protoc -I="${PROTO_DIR}" -I="${PROTO_IMPORT_DIR}" \
    --go_out=.. \
    --go-grpc_out=.. \
    "$PROTO_FILE"

  protoc -I="${PROTO_DIR}" -I="${PROTO_IMPORT_DIR}" \
    --grpc-gateway_out=.. \
    --grpc-gateway_opt=logtostderr=true \
    "$PROTO_FILE"

  protoc -I="${PROTO_DIR}" -I="${PROTO_IMPORT_DIR}" \
    --openapiv2_out="${DOCS_DIR}" \
    --openapiv2_opt=logtostderr=true \
    "$PROTO_FILE"
done

echo "Proto files and gateway definitions generated successfully!"