#!/usr/bin/env fish

set boot localhost:1122
set node localhost:1123
set authpass password
set ownerpass password

# ./bcn node start --node $boot --bootstrap --authpass $authpass \
#   --ownerpass $ownerpass --balance 1000
# ./bcn account create --node $boot --ownerpass $ownerpass
# ./bcn node start --node $node --seed $boot

function txSignAndSend -a node from to value ownerpass
  set tx (./bcn tx sign --node $node --from $from --to $to --value $value \
    --ownerpass $ownerpass)
  echo $tx
  ./bcn tx send --node $node --sigtx $tx
end

set own 1dc67739c409b169d8f981525366355694c7de9e24188d1814a7e2159857a878
set ben 0b283b314c12c66ce7ad65da7d5ab3008d28e25a988308721f010e5a04f23247

# txSignAndSend $boot $own $ben 2 $ownerpass
# txSignAndSend $boot $ben $own 1 $ownerpass

# ./bcn account balance --node $boot --account $own
# ./bcn account balance --node $boot --account $ben

# ./bcn block search --node $boot --number 2
# ./bcn block search --node $boot --hash 96b3d1d
# ./bcn block search --node $boot --parent 76088e0

# ./bcn tx search --node $boot --hash 6fe5fff
# ./bcn tx search --node $boot --from d8a05ac
# ./bcn tx search --node $boot --to fd29d48
# ./bcn tx search --node $boot --account d8a05ac
