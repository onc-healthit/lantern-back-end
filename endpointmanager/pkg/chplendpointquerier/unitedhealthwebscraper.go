package chplendpointquerier

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func UnitedHealthURLWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	CHPLURL = strings.TrimSpace(strings.Trim(CHPLURL, "\"'"))

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".cmp-text__paragraph")
	if err != nil {
		log.Fatal(err)
	}

	reURL := regexp.MustCompile(`^https?://\S+$`)

	expectURLBlock := false
	found := false

	doc.Find("div.cmp-text__paragraph").Each(func(_ int, div *goquery.Selection) {
		if found {
			return
		}

		// First block: label
		if strings.Contains(div.Text(), "Base request URL") {
			expectURLBlock = true
			return
		}

		// Next block: contains two <p> URLs; take the first one
		if expectURLBlock {
			div.Find("p").Each(func(_ int, p *goquery.Selection) {
				if found {
					return
				}
				txt := strings.TrimSpace(p.Text())
				if reURL.MatchString(txt) {
					url := strings.TrimRight(txt, "/") + "/R4"
					lanternEntryList = append(lanternEntryList, LanternEntry{URL: url})
					found = true
				}
			})

			expectURLBlock = false
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
