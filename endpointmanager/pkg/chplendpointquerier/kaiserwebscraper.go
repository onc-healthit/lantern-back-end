package chplendpointquerier

import (
	"context"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

func KaiserURLWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := KaiserChromedpQueryEndpointList(CHPLURL, ".opblock-tag-section")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".language-json").Each(func(index int, codehtml *goquery.Selection) {
		found := false
		processed := false
		codehtml.Find("span").Each(func(index int, spanhtml *goquery.Selection) {
			if strings.Contains(spanhtml.Text(), "CapabilityStatement") {
				found = true
			}
			if found {
				if strings.HasSuffix(spanhtml.Text(), "/FHIR/api\"") {
					var entry LanternEntry
					URL := strings.TrimSpace(spanhtml.Text())
					URL = strings.ReplaceAll(URL, "\"", "")
					entry.URL = URL
					processed = true
					lanternEntryList = append(lanternEntryList, entry)

				}
			}
			if processed {
				return
			}

		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}

// KaiserChromedpQueryEndpointList queries the given endpoint list and clicks buttons using chromedp and returns the html document
func KaiserChromedpQueryEndpointList(endpointListURL string, waitVisibleElement string) (*goquery.Document, error) {

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	timeoutContext, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	var htmlContent string
	var err error

	if len(waitVisibleElement) > 0 {
		// Chromedp will wait a max of 30 seconds for webpage to run javascript code to generate api search results before grapping HTML
		err = chromedp.Run(timeoutContext,
			chromedp.Navigate(endpointListURL),
			chromedp.WaitVisible(waitVisibleElement, chromedp.ByQuery),

			// Expand the Metadata section
			chromedp.WaitVisible(`.expand-operation`),
			chromedp.Click(`.expand-operation`, chromedp.ByQuery),

			// Expand the Metadata endpoint section
			chromedp.WaitVisible(`.opblock-summary-control`),
			chromedp.Click(`.opblock-summary-control`, chromedp.ByQuery),

			// Wait till the code snippet is rendered
			chromedp.WaitVisible(`.language-json`),

			chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
		)
	} else {
		err = chromedp.Run(timeoutContext,
			chromedp.Navigate(endpointListURL),
			chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
		)
	}

	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	return doc, nil
}
