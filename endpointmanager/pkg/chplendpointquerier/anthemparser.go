package chplendpointquerier

import (
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

type AntemOrganization struct {
	Prefix       string `json:"Prefix"`
	SiteBase     string `json:"SiteBase"`
	EndPoint     string `json:"EndPoint"`
	Version      string `json:"Version"`
	EndPointType string `json:"EndPointType"`
	LogoUrl      string `json:"LogoUrl"`
	Name         string `json:"Name"`
}

func AnthemURLParser(willowURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(willowURL)
	if err != nil {
		log.Fatal(err)
	}
	var organizations []AntemOrganization

	err = json.Unmarshal(respBody, &organizations)
	if err != nil {
		log.Fatal(err)
	}

	for _, org := range organizations {
		var entry LanternEntry

		entry.URL = org.EndPoint
		entry.OrganizationName = org.Name

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList
	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
