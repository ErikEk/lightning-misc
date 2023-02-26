#! /bin/bash

echo "Send amp payment from dave to alice..."

cd alice
echo "In alice directory"
lnclia="lncli --rpcserver=localhost:10001 --macaroonpath=data/chain/bitcoin/simnet/admin.macaroon addinvoice --amt=100 --amp"
addinvoice_command=$($lnclia)

echo $addinvoice_command | jq -r '.payment_request' > ../alice_pay_req.txt
#read tes <($addinvoice_command | jq -r '.payment_request')
#t=$kk | jq -r '.payment_request'
#echo $tes

cd ../dave
echo "in dave directory:"

alice_pay_req=$(< ../alice_pay_req.txt)
echo $alice_pay_req
lnclid="lncli --rpcserver=localhost:10004 --macaroonpath=data/chain/bitcoin/simnet/admin.macaroon"
eval "$lnclid sendpayment --dest=03622f978967a51015110d308bfa12e076a93f3bbae27ec24a209e9fc6ae104bca --amp --amt=20"

eval "$lnclid sendpayment --dest=03622f978967a51015110d308bfa12e076a93f3bbae27ec24a209e9fc6ae104bca --last_hop=02375d785c3ef60701a82c816741379eda146f3094eb9272daa3cae72df1ee8c63 --amp --outgoing_chan_id=29261302950068225 --amt=100"

#--last_hop=03ce43035c75ae972916e21f702326358e78539ffc284d10ce0864295e640d3e2b
#--outgoing_chan_id=29019410391957505
#--pay_req=$alice_pay_req

