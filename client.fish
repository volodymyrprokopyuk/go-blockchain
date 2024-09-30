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

set own 42e61ae200e77b00533f0faa54b536711298fd656aa8ae9b2cd491a8eac437c3
set ben f607fd36d6ed871db2a6021382452f54225d0cff8354698a0584f287019afec9

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
