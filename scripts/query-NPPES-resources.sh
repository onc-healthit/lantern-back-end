#!/bin/sh

#update source data from endpoint source list and NPPES
cd ../resources/prod_resources
YEAR=$(date +%Y)
PASTMONTH=$(date -v-1m +%B 2> /dev/null) || PASTMONTH=$(date -d '1 months ago' +%B)
MONTH=$(date +%B)

if [[ "${PASTMONTH}" -eq "December" ]]
then
  PASTYEAR=$(date -v-1y +%Y 2> /dev/null) || PASTYEAR=$(date -d '1 years ago' +%Y)
else
  PASTYEAR=$(date +%Y)
fi

NPPESFILE="https://download.cms.gov/nppes/NPPES_Data_Dissemination_${MONTH}_${YEAR}.zip"
PASTNPPESFILE="https://download.cms.gov/nppes/NPPES_Data_Dissemination_${PASTMONTH}_${PASTYEAR}.zip"

rm -f endpoint_pfile.csv
rm -f npidata_pfile.csv

echo "Downloading ${MONTH} NPPES Resources..."
curl -s -f -o temp.zip ${NPPESFILE} || echo "${MONTH} NPPES Resources not available, downloading ${PASTMONTH} NPPES Resources..." && curl -s -o temp.zip ${PASTNPPESFILE} 
echo "Extracting endpoint and npidata files from NPPES zip file..."
unzip -q temp.zip endpoint_pfile\*.csv
unzip -q temp.zip npidata_pfile\*.csv 
rm *FileHeader.csv
mv endpoint_pfile*.csv endpoint_pfile.csv
mv npidata_pfile*.csv npidata_pfile.csv
rm temp.zip

echo "Removing all entries from npidata_pfile that are not Entity Type 2 (Organization)..."
sed -E '/^[^,]*,[^,]*(\"1\"|\"\")/d' npidata_pfile.csv > npidata_pfile2.csv
rm npidata_pfile.csv
mv npidata_pfile2.csv npidata_pfile.csv

echo "Cutting down rows in npidata_pfile and endpoint_pfile for dev resources..."
rm -f ../dev_resources/npidata_pfile.csv 
rm -f ../dev_resources/endpoint_pfile.csv
sed '1000,$d' npidata_pfile.csv > ../dev_resources/npidata_pfile.csv 
sed '1000,$d' endpoint_pfile.csv > ../dev_resources/endpoint_pfile.csv

rm -f endpoint_pfile.csv
rm -f npidata_pfile.csv

echo "done"