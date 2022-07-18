package endpointwebscraper

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/chromedp/chromedp"
)

func Techcarewebscraper(vendorURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string

	// Chromedp will wait for webpage to run javascript code to generate api search results before grapping HTML
	err := chromedp.Run(ctx,
		chromedp.Navigate(vendorURL),
		chromedp.WaitVisible(".WordSection1"),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	count := 0

	doc.Find(".WordSection1").Each(func(index int, wordSectionElem *goquery.Selection) {
		wordSectionElem.Find("p").Each(func(indextr int, phtml *goquery.Selection) {
			// Only the first two entries are production server endpoints
			if count < 2 {
				var entry LanternEntry
				fhirURLLink := phtml.Find("a")
				if fhirURLLink.Length() > 0 {

					fhirURL, ok := fhirURLLink.Attr("href")
					if ok {
						fhirURL = strings.TrimSpace(fhirURL)
						entry.URL = fhirURL
						lanternEntryList = append(lanternEntryList, entry)
					}

					count++;
				}
			}
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
