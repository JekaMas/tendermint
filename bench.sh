#!/bin/bash

url="http://localhost:26657/broadcast_tx_async?tx="
END=20000
start=$(date +%s%N)

for i in $(seq 1 $END); do
    content="$(curl -s "$url\"$start$i\"")"
    #content="$(curl -s "$url\"x$i\"")"

	remainder=$(( i % 5000 ))
	if [ "$remainder" -eq 0 ]; then
	    echo "$content" >> output.txt

	    content1="$(curl -s "http://localhost:26657/num_unconfirmed_txs")"
        echo "$content1" >> output_2.txt
	fi
done

echo "Waiting for txs"
sleep 2

endtime=$(date +%s%N)
timediff=$(( endtime - start ))

content1="$(curl -s "http://localhost:26657/num_unconfirmed_txs")"

echo "Waiting for txs" >> output_2.txt
echo "$timediff" >> output_2.txt
echo "$content1" >> output_2.txt
