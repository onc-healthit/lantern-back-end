#!/bin/sh

#update source data from endpoint source list and NPPES
cd ../resources/prod_resources
YEAR=$(date +%Y)
PASTMONTH=$(date -v-1m +%B)
MONTH=$(date +%B)	
DATE=$(date +%Y%m%d)
PASTDATE=$(date -v-1m +%Y%m%d)	
NPPESFILE="https://download.cms.gov/nppes/NPPES_Data_Dissemination_${MONTH}_${YEAR}.zip"
PASTNPPESFILE="https://download.cms.gov/nppes/NPPES_Data_Dissemination_${PASTMONTH}_${YEAR}.zip"

rm -f endpoint_pfile.csv
rm -f npidata_pfile.csv
cd ../../scripts; chmod +rx query-endpoint-resources.sh; ./query-endpoint-resources.sh
cd ../resources/prod_resources
echo "Downloading ${MONTH} NPPES Resources..."
curl -s -f -o temp.zip ${NPPESFILE} || echo "${MONTH} NPPES Resources not available, downloading ${PASTMONTH} NPPES Resources..." && curl -s -o temp.zip ${PASTNPPESFILE} 
echo "Extracting endpoint and npidata files from NPPES zip file..."
unzip -q temp.zip endpoint_pfile\*.csv
unzip -q temp.zip npidata_pfile\*.csv 
rm *FileHeader.csv
mv endpoint_pfile*.csv endpoint_pfile.csv
mv npidata_pfile*.csv npidata_pfile.csv
rm temp.zip
echo "done"