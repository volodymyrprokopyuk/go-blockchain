#!/usr/bin/env fish

# yay -S protobuf
# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

function compile -a proto
  set -l import (path dirname $proto)
  set -l out (path dirname $import)
  protoc --proto_path $import --go_out $out --go-grpc_out $out $proto
end

set -l node node/rpc/node.proto
set -l acc node/rpc/account.proto
set -l tx node/rpc/tx.proto
set -l blk node/rpc/block.proto

compile $node
compile $acc
compile $tx
compile $blk
