#!/bin/bash
trap "rm server;kill 0" EXIT

go build  -o server
./server -port=8001 -api=0 &
./server -port=8002 -api=1 &
./server -port=8003 -api=2 &

sleep 2
echo ">>> start test"
curl "http://localhost:9000/api?key=a" &
curl "http://localhost:9001/api?key=e" &
curl "http://localhost:9002/api?key=h" &

wait