#! /bin/bash

echo "Send amp payment from alice to dave..."

cd dave
echo "In dave directory"
lnclid="lncli --rpcserver=localhost:10004 --macaroonpath=data/chain/bitcoin/simnet/admin.macaroon addinvoice --amt=50 --amp"
addinvoice_command=$($lnclid)

echo $addinvoice_command | jq -r '.payment_request' > ../dave_pay_req.txt
#read tes <($addinvoice_command | jq -r '.payment_request')
#t=$kk | jq -r '.payment_request'
#echo $tes

cd ../alice
echo "in alice directory:"

dave_pay_req=$(< ../dave_pay_req.txt)
echo $dave_pay_req
lnclia="lncli --rpcserver=localhost:10001 --macaroonpath=data/chain/bitcoin/simnet/admin.macaroon"
eval "$lnclia sendpayment --dest=03a6e8efb135441017d342461ff3ab98720185df905ed02f555e99775949c5fd1d --last_hop=03ce43035c75ae972916e21f702326358e78539ffc284d10ce0864295e640d3e2b --amp --amt=30"

eval "$lnclia sendpayment --dest=03a6e8efb135441017d342461ff3ab98720185df905ed02f555e99775949c5fd1d --last_hop=02375d785c3ef60701a82c816741379eda146f3094eb9272daa3cae72df1ee8c63 --amp --amt=20"

#--pay_req=$dave_pay_req

