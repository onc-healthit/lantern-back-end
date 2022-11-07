#!/bin/sh

EMAIL=

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

chmod +rx populatedb_prod.sh; ./populatedb_prod.sh || echo "Lantern failed to save endpoint information in database and download and save NPPES information." | /usr/bin/mail -s "Automatic prod database population error." ${EMAIL}