#!/bin/sh

cd ..
export $(cat .env)

#Iterate through endpoint source list json to query each url and store as properly named file
cd resources/prod_resources

jq -c '.[]' EndpointResourcesList.json | while read endpoint; do
   NAME=$(echo $endpoint | jq -c -r '.EndpointName')
   FILENAME=$(echo $endpoint | jq -c -r '.FileName')
   URL=$(echo $endpoint | jq -c -r '.URL')
 
   if [ "$NAME" = "1Up" ];
   then
      URL="${URL}?client_id=${LANTERN_1UPCLIENTID}&client_secret=${LANTERN_1UPCLIENTSECRET}"
   fi

   if [ -n "$URL" ];
   then
      echo "Downloading $NAME Endpoint Sources..."
      curl -s -o $FILENAME $URL
      echo "done"
   fi
done