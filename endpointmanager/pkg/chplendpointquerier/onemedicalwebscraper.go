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

func oneMedicalWebscraper(oneMedicalURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string

	// Chromedp will wait for webpage to run javascript code to generate api search results before grapping HTML
	err := chromedp.Run(ctx,
		chromedp.Navigate(oneMedicalURL),
		chromedp.WaitVisible("root-url"),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("main").Each(func(index int, mainElem *goquery.Selection) {
		mainElem.Find(".gdoc-page").Each(func(index int, gdocPage *goquery.Selection) {
			gdocPage.Find("article").Each(func(index int, articleElem *goquery.Selection) {
				articleElem.Find("p").Each(func(index int, pElem *goquery.Selection) {
					var entry LanternEntry

					log.Info(pElem.Text())
			
					if pElem.Length() > 0 {
						if strings.Contains(pElem.Text(), "Production root URL:") {	
							aElems := pElem.Find("a")

							if aElems.Length() > 0 {

								entryURL, exists := pElem.Eq(0).Attr("href")

								if exists {
									entryURL = strings.TrimSpace(entryURL)
									entry.URL = entryURL
									
									lanternEntryList = append(lanternEntryList, entry)
									return
								}
							}
						}
					}
				})
			})
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
