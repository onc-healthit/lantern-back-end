#!/bin/sh

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

EMAIL=
DIRECTORY=
YEAR=$(date +%Y)
MONTH=$(date +%B)
CURRENT_DATE=$(date +"%m_%d_%Y")
FILE_DIR="/lantern-project/onc-open-data/lantern-daily-data/$YEAR/$MONTH"
FILE_NAME=${CURRENT_DATE}endpointdata.csv
mkdir -p $FILE_DIR

EXPORTFILE="/etc/lantern/dailyexport/$FILE_NAME"
echo $EXPORTFILE
URL="http://127.0.0.1:8989/api/download"
curl -o "$EXPORTFILE" "$URL"
mv "$EXPORTFILE" "$FILE_DIR/$FILE_NAME"

cd $FILE_DIR

git pull
git checkout daily_export
git add $FILE_NAME
git commit -m "Adding daily export file ${FILE_NAME}"
git push --set-upstream origin daily_export
