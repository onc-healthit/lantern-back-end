package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	http "net/http"
	"os"
	"strings"
)

type endpointList struct {
	Endpoints []endpointEntry `json:"Endpoints"`
}
type endpointEntry struct {
	URL              string `json:"URL"`
	OrganizationName string `json:"OrganizationName"`
}

func main() {

	var chplURL string
	var fileToWriteTo string

	if len(os.Args) >= 1 {
		chplURL = os.Args[1]
		fileToWriteTo = os.Args[2]
	} else {
		log.Fatalf("ERROR: Missing command-line arguments")
	}

	var endpointEntryList endpointList

	client := &http.Client{}
	req, err := http.NewRequest("GET", chplURL, nil)
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

	var chplJSON map[string]interface{}
	err = json.Unmarshal(respBody, &chplJSON)
	if err != nil {
		log.Fatal(err)
	}

	chplResults := chplJSON["results"]
	if chplResults == nil {
		log.Fatal("CHPL endpoint list is empty")
	}

	chplResultsList, ok := chplResults.([]interface{})
	if !ok {
		log.Fatal("Error asserting CHPL endpoint list JSON is type []interface{}")
	}

	for _, chplEntry := range chplResultsList {
		chplEntry, ok := chplEntry.(map[string]interface{})
		if !ok {
			log.Fatal("Error converting CHPL endpoint entry to type map[string]interface{}")
		}

		developerName, ok := chplEntry["developer"].(string)
		if !ok {
			log.Fatal("Error converting CHPL developer name to type string")
		}

		endpointURLList, ok := chplEntry["serviceBaseUrlList"].([]interface{})
		if !ok {
			log.Fatal("Error converting serviceBasedUrlList to type []interface{}")
		}

		for _, url := range endpointURLList {
			var entry endpointEntry
			entry.OrganizationName = developerName

			urlString, ok := url.(string)
			if !ok {
				log.Fatal("Error converting CHPL developer name to type string")
			}
			index := strings.Index(urlString, "h")
			entry.URL = urlString[index:]
			endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, entry)
		}
	}

	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
