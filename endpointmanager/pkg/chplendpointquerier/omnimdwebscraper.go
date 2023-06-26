package chplendpointquerier

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
)

func OmniMDWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".sl-stack--1")
	if err != nil {
		log.Fatal(err)
	}

	found := false
	doc.Find("div").Each(func(index int, spanhtml *goquery.Selection) {
		spanbodyhtml := spanhtml.Find("span").Last()
		attr, _ := spanbodyhtml.Attr("aria-label")

		if strings.Contains(attr, "Live Server") && !found {
			entryURL := strings.TrimSpace(spanbodyhtml.Text())
			var entry LanternEntry
			entry.URL = entryURL

			lanternEntryList = append(lanternEntryList, entry)
			found = true
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
