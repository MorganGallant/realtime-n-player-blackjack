#!/bin/bash

# This script is used to compile the rpc.proto file for the two projects.

mkdir -p {client/rpc,server/rpc}
protoc -I . rpc.proto --go_out=plugins=grpc:.

cp rpc.pb.go client/rpc
cp rpc.pb.go server/rpc

rm -f rpc.pb.go