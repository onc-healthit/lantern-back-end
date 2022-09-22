package chplendpointquerier

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/chromedp/chromedp"
)

func UnifyWebscraper(unifyURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string

	// Chromedp will wait for webpage to run javascript code to generate api search results before grapping HTML
	err := chromedp.Run(ctx,
		chromedp.Navigate(unifyURL),
		chromedp.WaitVisible(".main-container", chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".box").Each(func(index int, boxElems *goquery.Selection) {
		h3Elem := boxElems.Find("h3")
		if h3Elem.Length() > 0 && h3Elem.Text() == "Request" {
			pEntries := boxElems.Find("p")

			if pEntries.Length() > 0 && strings.Contains(pEntries.Text(), "FHIR Base URL: ") {
				var entry LanternEntry

				aElem := pEntries.Find("a")

				entryURL := strings.TrimSpace(aElem.Text())
				entry.URL = entryURL

				lanternEntryList = append(lanternEntryList, entry)

				return
			}
		}
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
