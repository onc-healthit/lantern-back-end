package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func InpracsysURLWebscraper(CHPLURL string, fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".elementor-element.elementor-element-832f39a.elementor-widget.elementor-widget-text-editor")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".elementor-element.elementor-element-832f39a.elementor-widget.elementor-widget-text-editor").Each(func(index int, mainhtml *goquery.Selection) {
		mainhtml.Find("div").Each(func(indextr int, divhtml *goquery.Selection) {
			mainhtml.Find("span").Each(func(indextr int, spanhtml *goquery.Selection) {
				preElem := spanhtml.Find("pre").First()
				parts := strings.SplitN(preElem.Text(), ".com", 2)
				url := parts[0] + ".com"

				parts = strings.SplitN(url, "GET", 2)

				var entry LanternEntry

				entry.URL = strings.TrimSpace(parts[1])
				lanternEntryList = append(lanternEntryList, entry)
			})
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
