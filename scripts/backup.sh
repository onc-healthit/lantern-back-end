#!/bin/sh

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

EMAIL=
DB_NAME=
DB_USER=
BACKUP_DIR=

BACKUP=lantern_backup_$(date +%Y%m%d%H%M%S).sql
docker exec lantern-back-end_postgres_1 pg_dump -Fc -U ${DB_USER} -d ${DB_NAME} > "${BACKUP}"
echo "New Lantern Staging Backup ${BACKUP} available" | /usr/bin/mail -s "Lantern Staging Backup ${BACKUP} Available" "${EMAIL}"
rm -f ${BACKUP_DIR}/*.sql
mv "${BACKUP}" ${BACKUP_DIR}