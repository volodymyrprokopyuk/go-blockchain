#!/usr/bin/env fish

# yay -S protobuf
# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

function compile -a proto
  set -l import (path dirname $proto)
  set -l out (path dirname $import)
  protoc --proto_path $import --go_out $out --go-grpc_out $out $proto
end

set -l account node/raccount/account.proto
set -l store node/rstore/store.proto
set -l tx node/rtx/tx.proto

compile $account
compile $store
compile $tx
