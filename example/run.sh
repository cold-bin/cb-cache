#!/bin/bash
trap "rm server;kill 0" EXIT

go build  -o server
./server -port=8001 -api=0 &
./server -port=8002 -api=1 &
./server -port=8003 -api=2 &
 # todo
  #	{
  #		"a": "1",
  #		"b": "2",
  #		"c": "3",
  #	}, 8001 9000
  #	{
  #		"d": "4",
  #		"e": "5",
  #		"f": "6",
  #	}, 8002 9001
  #	{
  #		"g": "7",
  #		"h": "8",
  #		"k": "9",
  #	}, 8003 9002
sleep 2
echo ">>> start test"
#curl "http://localhost:9000/api?key=b"
#curl "http://localhost:9001/api?key=e"
#curl "http://localhost:9002/api?key=a"

wait