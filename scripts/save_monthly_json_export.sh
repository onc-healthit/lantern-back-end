#!/bin/sh

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

EMAIL=emichaud@mitre.org
PASTMONTHDATE=$(date -v-1m +%B%Y 2> /dev/null) || PASTMONTH=$(date -d '1 months ago' +%B%Y)
EXPORTFILE="/etc/lantern/exportfolder/${PASTMONTHDATE}JsonExport.json"
docker exec --workdir /go/src/app/cmd/jsonexport lantern-back-end_endpoint_manager_1 go run main.go ${EXPORTFILE} true || echo "Lantern failed to create the ${PASTMONTHDATE} JSON export file." | /usr/bin/mail -s "Monthly JSON export creation error." ${EMAIL}