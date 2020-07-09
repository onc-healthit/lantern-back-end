#!/bin/sh

#extract npidata and endpoint pfiles from NPPES zip
cd ../resources/prod_resources
unzip -q temp.zip endpoint_pfile\*.csv
unzip -q temp.zip npidata_pfile\*.csv 
rm *FileHeader.csv
mv endpoint_pfile*.csv endpoint_pfile.csv
mv npidata_pfile*.csv npidata_pfile.csv
rm temp.zip