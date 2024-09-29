#!/usr/bin/env fish

set node localhost:1122
set pass password

# ./bcn node start --node $node --bootstrap --authpass $pass \
#   --ownerpass $pass --balance 1000
# ./bcn account create --node $node --ownerpass $pass
# ./bcn node start --node localhost:1123 --seed localhost:1122

function txSignAndSend -a node from to value pass
  set tx (./bcn tx sign --node $node --from $from --to $to --value $value \
    --ownerpass $pass)
  echo $tx
  set hash (./bcn tx send --node $node --sigtx $tx)
  echo $hash
end

set own 1e99b05ea4c43c1b928b0f2b028ea099bb72fcb624dfa5bbbd99128f5e670946
set ben 00e2f46d9e7c3e42eb69f41c2fe63c096c2114ca5de9f51c0540a1d02215b087

txSignAndSend $node $own $ben 2 $pass
txSignAndSend $node $ben $own 1 $pass

# ./bcn account balance --node $node --account $own
# ./bcn account balance --node $node --account $ben

# ./bcn block search --node $node --number 2
# ./bcn block search --node $node --hash 96b3d1d
# ./bcn block search --node $node --parent 76088e0

# ./bcn tx search --hash 6fe5fff
# ./bcn tx search --from d8a05ac
# ./bcn tx search --to fd29d48
# ./bcn tx search --account d8a05ac
