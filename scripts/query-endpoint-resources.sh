#!/bin/sh

log_file="/etc/lantern/logs/query-endpoint-resources_logs.txt"
current_datetime=$(date +"%Y-%m-%d %H:%M:%S")

#Iterate through endpoint source list json to query each url and store as properly named file
cd ..
export $(cat .env)
cd resources/prod_resources

echo "$current_datetime - Downloading Medicaid state Endpoint List..." >> $log_file
file_path="MedicaidState_EndpointSources.json"
csv_file_path="medicaid-state-endpoints.csv"
if [ -f "$csv_file_path" ]; then
   cd ../../endpointmanager/cmd/medicaidendpointquerier
   echo "$current_datetime - Querying Medicaid state endpoints..." >> $log_file
   go run main.go $file_path
   cd ../../../resources/prod_resources
   echo "$current_datetime - done" >> $log_file
fi

echo "$current_datetime - Parsing State Payer Endpoint List..." >> $log_file
file_path="MedicareStateEndpointResourcesList.json"
csv_file_path="payer-patient-access.csv"
if [ -f "$csv_file_path" ]; then
   cd ../../endpointmanager/cmd/medicareendpointquerier
   echo "$current_datetime - Querying Medicare state endpoints..." >> $log_file
   go run main.go $file_path
   cd ../../../resources/prod_resources
   echo "$current_datetime - done" >> $log_file
fi

jq -c '.[]' EndpointResourcesList.json | while read endpoint; do
   NAME=$(echo $endpoint | jq -c -r '.EndpointName')
   FILENAME=$(echo $endpoint | jq -c -r '.FileName')
   URL=$(echo $endpoint | jq -c -r '.URL')

   if [ -n "$URL" ];
   then
      echo "$current_datetime - Downloading $NAME Endpoint Sources..." >> $log_file
      if [ "$NAME" = "CareEvolution" ] ||  [ "$NAME" = "1Up" ];
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
            jq 'del(.Endpoints.[10:])' $FILENAME > ../dev_resources/$FILENAME
         fi
      fi
      echo "$current_datetime - done" >> $log_file
   fi
done

#Query CHPL endpoint resource list
echo "$current_datetime - Downloading Medicare State Endpoint List..." >> $log_file
URL="https://chpl.healthit.gov/rest/search/v3?api_key=${LANTERN_CHPLAPIKEY}&certificationCriteriaIds=182&certificationStatuses=Active,Suspended%20by%20ONC,Suspended%20by%20ONC-ACB"
FILENAME="CHPLEndpointResourcesList.json"
cd ../../endpointmanager/cmd/CHPLpopulator
go run main.go $URL $FILENAME
cd ../../../resources/prod_resources
echo "$current_datetime - done" >> $log_file

jq -c '.[]' MedicareStateEndpointResourcesList.json | while read endpoint; do
   NAME=$(echo $endpoint | jq -c -r '.EndpointName')
   FILENAME=$(echo $endpoint | jq -c -r '.FileName')
   URL=$(echo $endpoint | jq -c -r '.URL')
   if [ -n "$URL" ];
   then
      cd ../../endpointmanager/cmd/chplendpointquerier
      echo "$current_datetime - Downloading $NAME Endpoint Sources..." >> $log_file
      go run main.go $URL $FILENAME
      cd ../../../resources/prod_resources
      echo "$current_datetime - done" >> $log_file
   fi
done

echo "$current_datetime - Downloading CHPL Endpoint List..." >> $log_file
jq -c '.[]' CHPLEndpointResourcesList.json | while read endpoint; do
   NAME=$(echo $endpoint | jq -c -r '.EndpointName')
   FILENAME=$(echo $endpoint | jq -c -r '.FileName')
   URL=$(echo $endpoint | jq -c -r '.URL')

   if [ -n "$URL" ];
   then 
      cd ../../endpointmanager/cmd/chplendpointquerier
      echo "$current_datetime - Downloading $NAME Endpoint Sources..." >> $log_file
      go run main.go $URL $FILENAME
      cd ../../../resources/prod_resources
      echo "$current_datetime - done" >> $log_file
   fi
done