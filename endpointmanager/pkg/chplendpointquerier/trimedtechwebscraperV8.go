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

func TriMedTechV8Webscraper(trimedtechURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string

	// Chromedp will wait for webpage to run javascript code to generate api search results before grapping HTML
	err := chromedp.Run(ctx,
		chromedp.Navigate(trimedtechURL),
		chromedp.WaitVisible("get-smartconfiguration"),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("h4").Each(func(index int, h4Elems *goquery.Selection) {
		if strings.Contains(h4Elems.Text(), "main service base endpoint") {
			h4Elems.Find("a").Each(func(index int, aElems *goquery.Selection) {
				if aElems.Length() > 0 {
					hrefText, exists := aElems.Attr("href")
					if exists && !strings.Contains(hrefText, "#") {
						var entry LanternEntry

						entryURL := strings.TrimSpace(hrefText)
						entry.URL = entryURL

						lanternEntryList = append(lanternEntryList, entry)

						return
					}
				}
			})
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
