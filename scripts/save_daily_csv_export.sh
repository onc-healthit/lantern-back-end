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
log_file="/etc/lantern/logs/save_daily_csv_export_logs.txt"
EXPORTFILE="/etc/lantern/dailyexport/$FILE_NAME"
echo $EXPORTFILE
#URL="http://127.0.0.1:8989/api/download"
URL="https://lantern.healthit.gov/api/daily/download"
expected_response_code=200
max_attempts=10  # Number of times to attempt the request
current_datetime=$(date +"%Y-%m-%d %H:%M:%S")

attempts=0
echo "$current_datetime - Starting daily download.." >> $log_file
while [ $attempts -lt $max_attempts ]; do
    response_code=$(curl -s -o "$EXPORTFILE" -w "%{http_code}" "$URL")

     echo "$current_datetime - response code is $response_code" >> $log_file
    if [ $response_code -eq $expected_response_code ]; then
        current_datetime=$(date +"%Y-%m-%d %H:%M:%S")
        echo "$current_datetime - Received a 200 OK response. Writing to GitHub repo." >> $log_file
        mv "$EXPORTFILE" "$FILE_DIR/$FILE_NAME"
        cd $FILE_DIR
        git pull
        git checkout main
        git add $FILE_NAME
        git commit -m "Adding daily export file ${FILE_NAME}"
        git push --set-upstream origin main
        echo "$current_datetime - Done writing to GitHub repo" >> $log_file
        break
    else
        echo "$current_datetime - Received response code $response_code. Retrying in 3 mins..." >> $log_file
        sleep 180
    fi

    ((attempts++))
done

if [ $attempts -eq $max_attempts ]; then
    echo "Maximum number of attempts reached. Exiting." >> $log_file
fi
