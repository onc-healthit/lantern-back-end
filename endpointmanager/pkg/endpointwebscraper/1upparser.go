package endpointwebscraper

import (
	"github.com/spf13/viper"

	"encoding/json"

	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func OneUpQuerier(oneUpURL string, fileToWriteTo string) {
	clientSecret := viper.GetString("1up_client_secret")
	clientID := viper.GetString("1up_client_id")

	if clientSecret == "" && clientID == "" {
		return
	}

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	oneUpURL = oneUpURL + "?client_id=" + clientID + "&client_secret=" + clientSecret + "&systemType=HealthSystem"

	respBody, err := helpers.QueryEndpointList(oneUpURL)
	if err != nil {
		log.Fatal(err)
	}

	var oneUpArr []map[string]interface{}
	err = json.Unmarshal(respBody, &oneUpArr)
	if err != nil {
		log.Fatal(err)
	}

	for _, oneUpEntry := range oneUpArr {
		var entry LanternEntry

		serviceBaseURL, ok := oneUpEntry["resource_url"].(string)
		if !ok {
			log.Fatal("Error converting resource_url to type string")
		} else {
			entry.URL = strings.TrimSpace(serviceBaseURL)
		}

		developerName, ok := oneUpEntry["name"].(string)
		if ok {
			entry.OrganizationName = strings.TrimSpace(developerName)
		}

		locationArr, ok := oneUpEntry["locations"].([]interface{})
		if ok && len(locationArr) > 0 {
			locationObj, ok := locationArr[0].(map[string]interface{})
			if ok {
				addressObj, ok := locationObj["address"].(map[string]interface{})
				if ok {
					postalCode, ok := addressObj["postalCode"].(string)
					if ok {
						entry.OrganizationZipCode = strings.TrimSpace(postalCode)
					}
				}
			}
		}

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList
	err = WriteEndpointListFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
