#!/bin/sh

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

cd ..
export $(cat .env)
cd resources/prod_resources

EMAIL=

#Query CHPL endpoint resource list
echo "Downloading CHPL Endpoint List..."
URL="https://chpl.healthit.gov/rest/search/beta?api_key=${LANTERN_CHPLAPIKEY}&certificationCriteriaIds=182"
FILENAME="CHPLEndpointResourcesList.json"
cd ../../endpointmanager/cmd/CHPLupdatecheck
go run main.go $URL $FILENAME
cd ../../../resources/prod_resources

FILE="updatedEmails.json"
if [[ -f "$FILE" ]]; then
    UPDATEDURLS=""
    while read newURL; do
        UPDATEURL=$(echo $newURL)
        UPDATEDURLS="${UPDATEDURLS} ${UPDATEURL} \n"
    done <<< "$(jq -c '.[]' updatedEmails.json)"
    EMAILMESSAGE="CHPL Endpoint Resources List has been updated with the following URLs:\n ${UPDATEDURLS}"
    echo "$EMAILMESSAGE" | /usr/bin/mail -s "CHPL Endpoint Resources List Update" ${EMAIL}
    rm ./updatedEmails.json
fi
echo "done"
