#!/bin/sh

#Iterate through endpoint source list json to query each url and store as properly named file
cd ../resources/prod_resources
jq -c '.[]' EndpointResourcesList.json | while read endpoint; do
   NAME=$(echo $endpoint | jq -c -r '.EndpointName')
   FILENAME=$(echo $endpoint | jq -c -r '.FileName')
   URL=$(echo $endpoint | jq -c -r '.URL')
   
   echo "Downloading $NAME Endpoint Sources..."
   curl -s -o $FILENAME $URL
   echo "done"
done