package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	http "net/http"
	"os"
	"strings"
)

type endpointEntry struct {
	FormatType   string `json:"FormatType"`
	URL          string `json:"URL"`
	EndpointName string `json:"EndpointName"`
	FileName     string `json:"FileName"`
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

	var endpointEntryList []endpointEntry

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
		log.Fatal("Error converting CHPL endpoint list JSON is type []interface{}")
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
		developerName = strings.TrimSpace(developerName)

		// serviceBaseUrlList is an array, so loop through list and add each url with developer name to endpoint list
		endpointURLList, ok := chplEntry["serviceBaseUrlList"].([]interface{})
		if !ok {
			log.Fatal("Error converting serviceBasedUrlList to type []interface{}")
		}

		for _, url := range endpointURLList {
			var entry endpointEntry

			urlString, ok := url.(string)
			if !ok {
				log.Fatal("Error converting CHPL url to type string")
			}
			urlString = strings.TrimSpace(urlString)

			// Remove all characters before the 'h' in http in the url
			index := strings.Index(urlString, "h")
			entry.URL = urlString[index:]

			entry.EndpointName = developerName

			// Get fileName from URL domain name
			index = strings.Index(urlString, ".")
			fileName := urlString[index+1:]
			index = strings.Index(fileName, ".")
			fileName = fileName[:index]

			entry.FileName = fileName + "EndpointSources.json"
			entry.FormatType = ""

			endpointEntryList = append(endpointEntryList, entry)
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
