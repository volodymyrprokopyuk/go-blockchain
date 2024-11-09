#!/usr/bin/env fish

set boot localhost:1122
set node localhost:1123
set authpass password
set ownerpass password

function txSignAndSend -a node from to value ownerpass
  set tx (./bcn tx sign --node $node --from $from --to $to --value $value \
    --ownerpass $ownerpass)
  echo SigTx $tx
  ./bcn tx send --node $node --sigtx $tx
end

function txProveAndVerify -a node hash mrkroot
  set mrkproof (./bcn tx prove --node $node --hash $hash)
  echo MerkleProof $mrkproof
  echo MerkleRoot $mrkroot
  ./bcn tx verify --node $node --hash $hash \
    --mrkproof $mrkproof --mrkroot $mrkroot
end

set acc1 66d614174909403746df7c3222cd74ca386995e4de11cfc99ca1efe548d33105
set acc2 0a6c57d451f561d6baefe35bba47f8dd682b31da27f0dfdedc646648ea5d12ba

# txSignAndSend $boot $acc1 $acc2 2 $ownerpass
# txSignAndSend $boot $acc2 $acc1 1 $ownerpass

set tx1 4312eb8f506a00c4f4f111ea8b318a871615115e5b1a49f14784c5f90a04baeb
set mrk1 c39f7787a0e1ad825964226031d1ede60f4a8546ce4a5f724321b22ffc3c7394
set tx2 bd849704122be82ee588c2abfacb8e12fb5bac0916356babcdb2b1683bbc684e
set mrk2 c39f7787a0e1ad825964226031d1ede60f4a8546ce4a5f724321b22ffc3c7394

txProveAndVerify $boot $tx1 $mrk1
txProveAndVerify $boot $tx2 $mrk2
