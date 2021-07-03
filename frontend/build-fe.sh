#!/usr/bin/env bash

source options.txt
mkdir -p "$WEBROOT"
mkdir -p tmp

# make sure necessary directories exist in the source folder
for d in js css fonts icons img bt; do 
    mkdir -p "$SOURCE/$d"
done

while IFS= read -r line || [[ -n "$line" ]]; do
    if [ "" != "$line" ] && [ "#" != "${line:0:1}" ]
    then
        subst="$line"
        subst=${subst//\$WEBROOT/$WEBROOT}
        subst=${subst//\$SOURCE/$SOURCE}
        cp -r -v $subst || exit 1
    else
        tput smso
        echo $line
        tput rmso
    fi
done < "$RESOURCES_FILE"

echo "Copying files complete. Stand by while post-processing..."
./post-process.sh
