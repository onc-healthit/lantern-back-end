package chplendpointquerier

import (
	"encoding/json"
	"io/ioutil"
	http "net/http"

	log "github.com/sirupsen/logrus"
)

type endpointList struct {
	Endpoints []lanternEntry `json:"Endpoints"`
}

type lanternEntry struct {
	URL              string `json:"URL"`
	OrganizationName string `json:"OrganizationName"`
	NPIID            string `json:"NPIID"`
}

func medHostQuerier(medhostURL string, fileToWriteTo string) {

	var lanternEntryList []lanternEntry
	var endpointEntryList endpointList

	client := &http.Client{}
	req, err := http.NewRequest("GET", medhostURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var medhostArr []interface{}
	err = json.Unmarshal(respBody, &medhostArr)
	if err != nil {
		log.Fatal(err)
	}

	for _, medhostEntry := range medhostArr {
		var entry lanternEntry
		medhostEntryInt, ok := medhostEntry.(map[string]interface{})
		if !ok {
			log.Fatal("Error converting medhost endpoint entry to type map[string]interface{}")
		}

		serviceBaseURL, ok := medhostEntryInt["serviceBaseUrl"].(string)
		if !ok {
			log.Fatal("Error converting serviceBaseUrl to type string")
		} else {
			entry.URL = serviceBaseURL
		}

		developerName, ok := medhostEntryInt["facilityName"].(string)
		if ok {
			entry.OrganizationName = developerName
		}

		npiID, ok := medhostEntryInt["npi"].(string)
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
