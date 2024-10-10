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

set acc1 4f3748d4d46b695a85f1773b6cb86aa0837818d5df33550180c5b8da7c966a6f
set acc2 bba08a59c80977b2bbf5df4f9d09471ddf1592aa7b0133377c5df865e73a8b12
# txSignAndSend $boot $acc1 $acc2 2 $ownerpass
# txSignAndSend $node $acc2 $acc1 1 $ownerpass

# ./bcn account balance --node $boot --account $own
# ./bcn account balance --node $boot --account $ben

# ./bcn block search --node $boot --number 2
# ./bcn block search --node $boot --hash 96b3d1d
# ./bcn block search --node $boot --parent 76088e0

# ./bcn tx search --node $boot --hash 6fe5fff
# ./bcn tx search --node $boot --from d8a05ac
# ./bcn tx search --node $boot --to fd29d48
# ./bcn tx search --node $boot --account d8a05ac
