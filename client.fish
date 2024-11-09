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

function txProveAndVerify -a prover verifier hash mrkroot
  set mrkproof (./bcn tx prove --node $prover --hash $hash)
  echo MerkleProof $mrkproof
  echo MerkleRoot $mrkroot
  ./bcn tx verify --node $verifier --hash $hash \
    --mrkproof $mrkproof --mrkroot $mrkroot
end

set acc1 231c83f0a857cfb1e88f8adb92371e01aa1bdc80ef88ea443a2fccf02f444720
set acc2 cb68e5de26f72110e13e47b2519fcd48ca941a0f4f572bd9751654d01499b910

# txSignAndSend $boot $acc1 $acc2 2 $ownerpass
# txSignAndSend $boot $acc2 $acc1 1 $ownerpass

set tx1 6040ff5315af566ed974a737dbf460f04e73c9a713ef494e9baacfe7dd5dc8f1
set mrk1 6040ff5315af566ed974a737dbf460f04e73c9a713ef494e9baacfe7dd5dc8f1
set tx2 b87703f6bf0035613f638657293da795bc771ff414c000da894b05d22f5a70b8
set mrk2 b87703f6bf0035613f638657293da795bc771ff414c000da894b05d22f5a70b8

# txProveAndVerify $boot $boot $tx1 $mrk1
# txProveAndVerify $boot $boot $tx2 $mrk2

# txProveAndVerify $node $boot $tx1 $mrk1
# txProveAndVerify $node $boot $tx2 $mrk2

# txProveAndVerify $node $boot $tx1 $mrk2
