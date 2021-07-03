#!/usr/bin/env bash

source options.txt
set -e
set -v

./install-dependencies.sh 
./build-fe.sh 
./serve.sh & 
SRV_PID=$!
sleep 3s
xdg-open "$URL"

echo "Web server is running under PID $SRV_PID"
fg 
