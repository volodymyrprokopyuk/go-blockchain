#!/usr/bin/env fish

# yay -S protobuf
# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

function compile -a proto
  set import (path dirname $proto)
  set out (path dirname $import)
  protoc --proto_path $import --go_out $out --go-grpc_out $out $proto
end

set node node/rpc/node.proto
set acc node/rpc/account.proto
set tx node/rpc/tx.proto
set blk node/rpc/block.proto

compile $node
compile $acc
compile $tx
compile $blk
