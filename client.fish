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

set acc1 8824f522bb131453c83225b276a3a3f8f145c99fb3518e3a7219b3f2f3bc0a0c
set acc2 715aa9e36740bce382a543b10fd70cad0bc1796b04fd7e49677a2fdcd1eac95c
txSignAndSend $boot $acc1 $acc2 2 $ownerpass
txSignAndSend $boot $acc2 $acc1 1 $ownerpass

# ./bcn account balance --node $boot --account $own
# ./bcn account balance --node $boot --account $ben

# ./bcn block search --node $boot --number 2
# ./bcn block search --node $boot --hash 96b3d1d
# ./bcn block search --node $boot --parent 76088e0

# ./bcn tx search --node $boot --hash 6fe5fff
# ./bcn tx search --node $boot --from d8a05ac
# ./bcn tx search --node $boot --to fd29d48
# ./bcn tx search --node $boot --account d8a05ac
