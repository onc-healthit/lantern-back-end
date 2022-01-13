package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	http "net/http"
	"path/filepath"
	"strings"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type endpointEntry struct {
	FormatType   string `json:"FormatType"`
	URL          string `json:"URL"`
	EndpointName string `json:"EndpointName"`
	FileName     string `json:"FileName"`
}

var chplEndpointList []endpointEntry

func main() {
	var chplURL string
	var fileToWriteTo string

	if len(os.Args) >= 2 {
		chplURL = os.Args[1]
		fileToWriteTo = os.Args[2]
	}else {
		log.Fatalf("ERROR: Missing command-line arguments")
	}

	// Get CHPL Endpoint List from CHPL URL
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

	// Get CHPL Endpoint list stored in Lantern resources folder
	path := filepath.Join("../../../resources/prod_resources/", "CHPLEndpointResourcesList.json")
	CHPLFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(CHPLFile, &chplEndpointList)
	if err != nil {
		log.Fatal(err)
	}

	newURLs := CHPLEndpointListUpdateCheck(chplResultsList)
	if (len(newURLs) <= 0) {
		log.Info("CHPL list does not need to be updated.")
		return
	} else {
		finalFormatJSON, err := json.MarshalIndent(chplEndpointList, "", "\t")
		if err != nil {
			log.Fatal(err)
		}
	
		err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
		if err != nil {
			log.Fatal(err)
		}

		finalFormatJSON, err = json.MarshalIndent(newURLs, "", "\t")
		if err != nil {
			log.Fatal(err)
		}
	
		err = ioutil.WriteFile("../../../resources/prod_resources/updatedEmails.json", finalFormatJSON, 0644)
		if err != nil {
			log.Fatal(err)
		}

		log.Info(fmt.Sprintf("CHPLEndpointList has been updated with new entries: %v", newURLs))
	
	}
}

func CHPLEndpointListUpdateCheck(chplResultsList []interface{}) []string {

	var existingURLs []string
	var newURLs []string
	for _, chplEntry := range chplEndpointList {
		existingURLs = append(existingURLs, chplEntry.URL)
	}


	for _, chplEntries := range chplResultsList {
		
		chplEntry, ok := chplEntries.(map[string]interface{})
		if !ok {
			log.Fatal("Error converting CHPL result entry to type map[string]interface{}")
		}

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
			entryURL := urlString[index:]

			if !stringArrayContains(existingURLs, entryURL) {
				if !contains(chplEndpointList, entryURL) {

					developerName, ok := chplEntry["developer"].(string)
					if !ok {
						log.Fatal("Error converting CHPL developer name to type string")
					}
					developerName = strings.TrimSpace(developerName)

					entry.URL = entryURL
					entry.EndpointName = developerName

					// Get fileName from URL domain name
					index = strings.Index(urlString, ".")
					fileName := urlString[index+1:]
					index = strings.Index(fileName, ".")
					fileName = fileName[:index]

					entry.FileName = fileName + "EndpointSources.json"
					entry.FormatType = ""

					chplEndpointList = append(chplEndpointList, entry)
					
					// Add URL to list of new entries that were updated
					newURLs = append(newURLs, entryURL)
				}
			}
		}
	}

	return newURLs
}

func contains(endpointEntryList []endpointEntry, url string) bool {
	for _, e := range endpointEntryList {
		if e.URL == url {
			return true
		}
	}
	return false
}

func stringArrayContains(urlList []string, url string) bool {
	for _, urlExist := range urlList {
		if urlExist == url {
			return true
		}
	}
	return false
}