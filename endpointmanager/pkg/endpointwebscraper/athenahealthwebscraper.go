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

func Athenawebscraper(vendorURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string

	log.Info("Getting the changes")

	// Chromedp will wait for webpage to run javascript code to generate api search results before grapping HTML
	err := chromedp.Run(ctx,
		chromedp.Navigate(vendorURL),
		chromedp.WaitVisible(".content"),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	log.Fatalf("%v", htmlContent)

	doc.Find("table").Each(func(index int, rowhtml *goquery.Selection) {
		rowhtml.Find("tr").Each(func(indextr int, rowbodyhtml *goquery.Selection) {
			var entry LanternEntry
			tableEntries := rowbodyhtml.Find("td")
			if tableEntries.Length() > 0 {
				organizationName := strings.TrimSpace(tableEntries.Eq(0).Text())
				fhirURL := strings.TrimSpace(tableEntries.Eq(1).Text())

				entry.OrganizationName = organizationName
				entry.URL = fhirURL

				log.Infof("%v", fhirURL)

				lanternEntryList = append(lanternEntryList, entry)
			}
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("%v", endpointEntryList)
	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
