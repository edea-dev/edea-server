#!/usr/bin/env bash

source options.txt

if [ -d "$WEBROOT" ]
then
    python3 -m http.server --directory "$WEBROOT" $PORT || echo "Failed to start webserver."
else
    echo "The webroot $WEBROOT does not exist! Run ./build-fe.sh first."
fi

