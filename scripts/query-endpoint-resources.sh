#!/bin/sh

#Iterate through endpoint source list json to query each url and store as properly named file
cd ../resources/prod_resources
jq -c '.[]' EndpointResourcesList.json | while read endpoint; do
   NAME=$(echo $endpoint | jq -c -r '.EndpointName')
   FILENAME=$(echo $endpoint | jq -c -r '.FileName')
   URL=$(echo $endpoint | jq -c -r '.URL')

   if [ -n "$URL" ];
   then
      echo "Downloading $NAME Endpoint Sources..."
      if [ "$NAME" = "CareEvolution" ] ||  [ "$NAME" = "1Up" ];
      then
         cd ../../endpointmanager/cmd/endpointwebscraper
         go run main.go $NAME $URL $FILENAME
         cd ../../../resources/prod_resources
      else
         curl -s -o $FILENAME $URL
      fi
      echo "done"
   fi
done