#! /bin/bash

echo "Send amp payment from dave to alice..."

cd alice
echo "In alice directory"
lnclia="lncli --rpcserver=localhost:10001 --macaroonpath=data/chain/bitcoin/simnet/admin.macaroon addinvoice --amt=50"
addinvoice_command=$($lnclia)

echo $addinvoice_command | jq -r '.payment_request' > ../alice_pay_req.txt

cd ../dave
echo "in dave directory:"

alice_pay_req=$(< ../alice_pay_req.txt)
echo $alice_pay_req
lnclid="lncli --rpcserver=localhost:10004 --macaroonpath=data/chain/bitcoin/simnet/admin.macaroon"
eval "$lnclid sendpayment --pay_req=$alice_pay_req"

#--dest=03a6e8efb135441017d342461ff3ab98720185df905ed02f555e99775949c5fd1d --last_hop=03ce43035c75ae972916e21f702326358e78539ffc284d10ce0864295e640d3e2b
