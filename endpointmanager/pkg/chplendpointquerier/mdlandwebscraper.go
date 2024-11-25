package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func MdlandWebscraper(chplURL string, fileToWriteTo string) error {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, ".MsoNormal")
	if err != nil {
		log.Info(err)
		return err
	}

	doc.Find("span").Each(func(index int, spanElem *goquery.Selection) {
		if strings.Contains(spanElem.Text(), "https://") && strings.Contains(spanElem.Text(), "metadata") {
			str := spanElem.Text()
			str = strings.Replace(str, "GET ", "", -1)
			str = strings.Replace(str, "/metadata", "", -1)
			str = strings.Replace(str, "\nHTTP/1.1", "", -1)

			var entry LanternEntry
			entry.URL = str
			lanternEntryList = append(lanternEntryList, entry)
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Info(err)
		return err
	}

	return nil
}
