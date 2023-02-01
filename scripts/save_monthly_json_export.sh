#!/bin/sh

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

EMAIL=
DIRECTORY=
YEAR=$(date +%Y)
PASTMONTH=$(date -v-1m +%B 2> /dev/null) || PASTMONTH=$(date -d '1 months ago' +%B)
if [ "${PASTMONTH}" = "December" ]
then
  YEAR=$(date -v-1y +%Y 2> /dev/null) || YEAR=$(date -d '1 years ago' +%Y)
fi
EXPORTFILE="/etc/lantern/exportfolder/${PASTMONTH}${YEAR}JsonExport.json"
docker exec --workdir /go/src/app/cmd/jsonexport lantern-back-end_endpoint_manager_1 go run main.go ${EXPORTFILE} "month" || echo "Lantern failed to create the ${PASTMONTH} JSON export file." | /usr/bin/mail -s "Monthly JSON export creation error." ${EMAIL}
docker cp lantern-back-end_endpoint_manager_1:${EXPORTFILE} ${DIRECTORY}
cd ${DIRECTORY}
zip ${PASTMONTH}${YEAR}JsonExport ${PASTMONTH}${YEAR}JsonExport.json
rm ${PASTMONTH}${YEAR}JsonExport.json
git add ${PASTMONTH}${YEAR}JsonExport.zip
git commit -m "Added ${PASTMONTH} ${YEAR} file"
git push