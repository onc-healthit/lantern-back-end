package endpointwebscraper

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	http "net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func HTMLtablewebscraper(vendorURL string, vendor string, fileToWriteTo string) {

	var endpointEntryList EndpointList

	res, err := http.Get(vendorURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			var entry LanternEntry
			tableEntries := rowhtml.Find("td")
			if tableEntries.Length() > 0 {
				if vendor == "CareEvolution" {
					if indextr != 1 {
						entry.OrganizationName = strings.TrimSpace(tableEntries.Eq(0).Text())
						entry.URL = strings.TrimSpace(tableEntries.Eq(1).Text())
						endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, entry)
					}
				} else if vendor == "1Up" {
					endpointType := strings.TrimSpace(tableEntries.Eq(2).Text())
					if endpointType == "Health System" {
						entry.OrganizationName = strings.TrimSpace(tableEntries.Eq(0).Find("a").Text())
						entry.URL = strings.TrimSpace(tableEntries.Eq(1).Text())
						endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, entry)
					}
				}
			}
		})
	})

	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
