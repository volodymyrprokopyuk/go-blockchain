#!/usr/bin/env fish

set boot localhost:1122
set node localhost:1123
set authpass password
set ownerpass password

function txSignAndSend -a node from to value ownerpass
  set tx (./bcn tx sign --node $node --from $from --to $to --value $value \
    --ownerpass $ownerpass)
  echo $tx
  ./bcn tx send --node $node --sigtx $tx
end

function txProveAndVerify -a node hash
  set proof (./bcn tx prove --node $node --hash $hash)
  echo Merkle proof $proof[1]
  echo Merkle root $proof[2]
  ./bcn tx verify --node $node --hash $hash \
    --mrkproof $proof[1] --mrkroot $proof[2]
end

set acc1 1034335fe3f62d16fdfb7a30d872234ef8e9e1899a68c07bbb9c94f594508fa9
set acc2 39e8aee509a53d1720c12ba4fb17de1f0778beaa5e7a8f06a5a831d6256efbce

# txSignAndSend $boot $acc1 $acc2 2 $ownerpass
# txSignAndSend $boot $acc2 $acc1 1 $ownerpass

set tx1 fe7d63dfd4c7cdb479f611d340e07e4bb7c4f89f4dd8c2afc3c224c0f0408bc8
set tx2 b7eea8c6220cc901caf0916ddd2faf03d6148d1aa8e31c7d6a01010c11c45f99

txProveAndVerify $boot $tx1
txProveAndVerify $boot $tx2
