package chplendpointquerier

import (
	"encoding/json"
	"io/ioutil"
	http "net/http"

	log "github.com/sirupsen/logrus"
)

func MedHostQuerier(medhostURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	client := &http.Client{}
	req, err := http.NewRequest("GET", medhostURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var medhostArr []map[string]interface{}
	err = json.Unmarshal(respBody, &medhostArr)
	if err != nil {
		log.Fatal(err)
	}

	for _, medhostEntry := range medhostArr {
		var entry LanternEntry

		serviceBaseURL, ok := medhostEntry["serviceBaseUrl"].(string)
		if !ok {
			log.Fatal("Error converting serviceBaseUrl to type string")
		} else {
			entry.URL = serviceBaseURL
		}

		developerName, ok := medhostEntry["facilityName"].(string)
		if ok {
			entry.OrganizationName = developerName
		}

		npiID, ok := medhostEntry["npi"].(string)
		if ok {
			entry.NPIID = npiID
		}

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList
	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
