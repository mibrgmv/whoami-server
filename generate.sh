#!/bin/bash

PROTO_DIR="./cmd/history/api" # change path to your .proto dir
PROTO_IMPORT_DIR="./third_party"
DOCS_DIR="./docs"

mkdir -p $DOCS_DIR

protoc -I=${PROTO_DIR} -I=${PROTO_IMPORT_DIR} \
  --go_out=.. \
  --go-grpc_out=.. \
  ${PROTO_DIR}/*.proto

protoc -I=${PROTO_DIR} -I=${PROTO_IMPORT_DIR} \
  --grpc-gateway_out=.. \
  --grpc-gateway_opt=logtostderr=true \
  ${PROTO_DIR}/*.proto

protoc -I=${PROTO_DIR} -I=${PROTO_IMPORT_DIR} \
  --openapiv2_out=${DOCS_DIR} \
  --openapiv2_opt=logtostderr=true \
  ${PROTO_DIR}/*.proto