#!/bin/sh

#Iterate through endpoint source list json to query each url and store as properly named file
cd ..
export $(cat .env)
cd resources/prod_resources

jq -c '.[]' EndpointResourcesList.json | while read endpoint; do
   NAME=$(echo $endpoint | jq -c -r '.EndpointName')
   FILENAME=$(echo $endpoint | jq -c -r '.FileName')
   URL=$(echo $endpoint | jq -c -r '.URL')

   if [ -n "$URL" ];
   then
      echo "Downloading $NAME Endpoint Sources..."
      if [ "$NAME" = "CareEvolution" ] ||  [ "$NAME" = "1Up" ] ||  [ "$NAME" = "AthenaHealth" ] ||  [ "$NAME" = "TechCare" ] ||  [ "$NAME" = "Carefluence" ];
      then
         cd ../../endpointmanager/cmd/endpointwebscraper
         go run main.go $NAME $URL $FILENAME
         cd ../../../resources/prod_resources
      else
         curl -s -o $FILENAME $URL
         
         if [ "$NAME" = "Cerner" ]
         then
            jq 'del(.endpoints[10:])' $FILENAME > ../dev_resources/$FILENAME
         else
            jq '.Endpoints.[10:]' $FILENAME > ../dev_resources/$FILENAME
         fi
      fi
      echo "done"
   fi
done

#Query CHPL endpoint resource list
echo "Downloading CHPL Endpoint List..."
URL="https://chpl.healthit.gov/rest/search/v2?api_key=${LANTERN_CHPLAPIKEY}&certificationCriteriaIds=182"
FILENAME="CHPLEndpointResourcesList.json"
cd ../../endpointmanager/cmd/CHPLpopulator
go run main.go $URL $FILENAME
cd ../../../resources/prod_resources
echo "done"

jq -c '.[]' CHPLEndpointResourcesList.json | while read endpoint; do
   NAME=$(echo $endpoint | jq -c -r '.EndpointName')
   FILENAME=$(echo $endpoint | jq -c -r '.FileName')
   URL=$(echo $endpoint | jq -c -r '.URL')

   if [ -n "$URL" ];
   then 
      cd ../../endpointmanager/cmd/chplendpointquerier
      echo "Downloading $NAME Endpoint Sources..."
      go run main.go $URL $FILENAME
      cd ../../../resources/prod_resources
      echo "done"
   fi
done