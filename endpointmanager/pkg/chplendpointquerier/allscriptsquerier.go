package chplendpointquerier

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"github.com/PuerkitoBio/goquery"
)

func AllScriptsQuerier(allscriptsURL string, fileToWriteTo string) {

	var endpointEntryList EndpointList

	var DSTU2URL string

	doc := helpers.ChromedpQueryEndpointList(allscriptsURL, "")

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
	respBody := helpers.QueryEndpointList(allscriptsURL + "/" + strings.Join(strings.Split(DSTU2URL, "/")[2:], "/"))

	// convert bundle data to lantern format
	endpointEntryList.Endpoints = BundleToLanternFormat(respBody)

	WriteCHPLFile(endpointEntryList, fileToWriteTo)

}
