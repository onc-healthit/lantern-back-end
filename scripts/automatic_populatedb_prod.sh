#!/bin/sh

EMAIL=

# Commenting out SHELL and PATH variables as they are causing Go version error during the execution of query-endpoint-resources.sh
#SHELL=/bin/sh
#PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin
log_file="/etc/lantern/logs/automatic_populatedb_prod_logs.txt"
current_datetime=$(date +"%Y-%m-%d %H:%M:%S")
LOGFILE=populatedb_logs_$(date +%Y%m%d%H%M%S).txt

#update source data from endpoint source list and NPPES
#cd ../resources/prod_resources
# YEAR=$(date +%Y)
# PASTMONTH=$(date -v-1m +%B 2> /dev/null) || PASTMONTH=$(date -d '1 months ago' +%B)
# MONTH=$(date +%B)

# if [[ "${PASTMONTH}" == "December" ]]
# then
#   PASTYEAR=$(date -v-1y +%Y 2> /dev/null) || PASTYEAR=$(date -d '1 years ago' +%Y)
# else
#   PASTYEAR=$(date +%Y)
# fi

# NPPESFILE="https://download.cms.gov/nppes/NPPES_Data_Dissemination_${MONTH}_${YEAR}.zip"
# PASTNPPESFILE="https://download.cms.gov/nppes/NPPES_Data_Dissemination_${PASTMONTH}_${PASTYEAR}.zip"

# rm -f endpoint_pfile.csv
# rm -f npidata_pfile.csv

# echo "$current_datetime - Downloading ${MONTH} NPPES Resources..." >> $log_file
# curl -s -f -o temp.zip ${NPPESFILE} || {
#   echo "$current_datetime - ${MONTH} NPPES Resources not available, downloading ${PASTMONTH} NPPES Resources..." >> $log_file && curl -s -o temp.zip ${PASTNPPESFILE} 
# }
# echo "$current_datetime - Extracting endpoint and npidata files from NPPES zip file..." >> $log_file
# unzip -q temp.zip endpoint_pfile\*.csv
# unzip -q temp.zip npidata_pfile\*.csv 
# rm *fileheader.csv
# mv endpoint_pfile*.csv endpoint_pfile.csv
# mv npidata_pfile*.csv npidata_pfile.csv
# rm temp.zip

# echo "$current_datetime - Removing all entries from npidata_pfile that are not Entity Type 2 (Organization)..." >> $log_file
# sed -E '/^[^,]*,[^,]*(\"1\"|\"\")/d' npidata_pfile.csv > npidata_pfile2.csv
# rm npidata_pfile.csv
# mv npidata_pfile2.csv npidata_pfile.csv

chmod +rx query-endpoint-resources.sh; ./query-endpoint-resources.sh

cd ../resources
cp -r prod_resources resources
docker cp resources lantern-back-end_endpoint_manager_1:/etc/lantern

echo "$current_datetime - Populating db with endpoint information..." >> $log_file
#cd ../../scripts
docker exec lantern-back-end_endpoint_manager_1 /etc/lantern/populatedb.sh || {
  echo "$current_datetime - Lantern failed to save endpoint information in database." >> $log_file
  echo "Lantern failed to save endpoint information in database." | /usr/bin/mail -s "Automatic prod database population error." ${EMAIL}
}
# cd ../resources/prod_resources
# rm -f endpoint_pfile.csv
# rm -f endpoint_pfile
# rm -f npidata_pfile.csv
# rm -f npidata_pfile

rm -r resources

echo "$current_datetime - done" >> $log_file

docker cp lantern-back-end_endpoint_manager_1:/etc/lantern/populatedb_logs.txt /etc/lantern/logs/populatedb_logs/${LOGFILE}