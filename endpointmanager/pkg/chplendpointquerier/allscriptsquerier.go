package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func AllScriptsQuerier(allscriptsURL string, fileToWriteTo string) {

	var endpointEntryList EndpointList

	var DSTU2URL string

	doc, err := helpers.ChromedpQueryEndpointList(allscriptsURL, "")
	if err != nil {
		log.Fatal(err)
	}

	// find <a> tags and iterateover them
	doc.Find("a").Each(func(index int, linkhtml *goquery.Selection) {
		// select only the 6th link on the page for DSTU2 endpoints
		if index == 5 {
			// get href from link
			href, _ := linkhtml.Attr("href")
			DSTU2URL = strings.TrimSpace(href)
		}
	})

	// concatenate dstu2 link onto base allscripts url to get full url and make request
	respBody, err := helpers.QueryEndpointList(allscriptsURL + "/" + strings.Join(strings.Split(DSTU2URL, "/")[2:], "/"))
	if err != nil {
		log.Fatal(err)
	}

	// convert bundle data to lantern format
	endpointEntryList.Endpoints = BundleToLanternFormat(respBody)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
