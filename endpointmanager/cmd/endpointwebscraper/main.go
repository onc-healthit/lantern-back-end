package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	http "net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

type endpointList struct {
	Entries []endpointEntry `json:"Entries"`
}
type endpointEntry struct {
	FHIRPatientFacingURI string `json:"FHIRPatientFacingURI"`
	OrganizationName     string `json:"OrganizationName"`
}

func main() {

	var vendor string
	var vendorURL string
	var fileToWriteTo string

	if len(os.Args) >= 1 {
		vendor = os.Args[1]
		vendorURL = os.Args[2]
		fileToWriteTo = os.Args[3]
	} else {
		log.Fatalf("ERROR: Missing command-line arguments")
	}

	var endpointEntryList endpointList

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
			var entry endpointEntry
			tableEntries := rowhtml.Find("td")
			if tableEntries.Length() > 0 {
				if vendor == "CareEvolution" {
					entry.OrganizationName = tableEntries.Eq(0).Text()
					entry.FHIRPatientFacingURI = tableEntries.Eq(1).Text()
					endpointEntryList.Entries = append(endpointEntryList.Entries, entry)
				} else if vendor == "1Up" {
					endpointType := tableEntries.Eq(3).Text()
					if endpointType == "Health System" {
						entry.OrganizationName = tableEntries.Eq(1).Find("a").Text()
						entry.FHIRPatientFacingURI = tableEntries.Eq(2).Text()
						endpointEntryList.Entries = append(endpointEntryList.Entries, entry)
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
