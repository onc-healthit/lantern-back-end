package chplendpointquerier

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func EzemrxWebscraper(CHPLURL string, fileToWriteTo string) error {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	var entry LanternEntry

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "#comp-lb6njyhb")
	if err != nil {
		log.Info(err)
		return err
	}

	divElem := doc.Find("#comp-lb6njyhb").First()
	pElem := divElem.Find("p").First()
	spanElem := pElem.Find("span").First()

	parts := strings.Split(spanElem.Text(), "\n")

	entry.URL = strings.TrimSpace(parts[1])

	lanternEntryList = append(lanternEntryList, entry)

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Info(err)
		return err
	}

	return nil
}
