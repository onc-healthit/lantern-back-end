package chplendpointquerier

import (
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func EmdsCloudWebscraper(CHPLURL string, fileToWriteTo string) {
	var entry LanternEntry
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	client := &http.Client{}
	req, err := http.NewRequest("GET", CHPLURL, nil)
	if err != nil {
		log.Fatalf("Error creating HTTP request: %v", err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making HTTP request: %v", err)
		return
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Warnf("Error closing response body: %v", err)
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
		return
	}

	var resources []map[string]interface{}
	if err := json.Unmarshal(body, &resources); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
		return
	}

	for _, resource := range resources {
		if url, ok := resource["ResourceUrl"].(string); ok {
			entry.URL = url
			lanternEntryList = append(lanternEntryList, entry)
		}
	}
	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}
}
