#!/usr/bin/env bash
source options.txt

set -e
# needed for build tools like sass
PATH="$PATH:$PWD/node_modules/.bin"

# apply Bootstrap Theme IFF the target older 
CCSS="$WEBROOT/css/custom.css"
CCSS_SRC="$SOURCE/sass/custom.scss"
[[ "$CCSS_SRC" -nt "$CCSS" ]] && sass "$CCSS_SRC" "$CCSS" && echo "Generated $CCSS"

# load fonts
for d in node_modules/@openfonts/*
do
    cat $d/index.css
done >> $WEBROOT/css/fonts.css
sed -i 's,\./files,/fonts,g' $WEBROOT/css/fonts.css

# add css and js compression here

# remove unnecessary files
find $WEBROOT -name "README.md" -delete

set +v
echo "Post-processing complete"
exit 0
