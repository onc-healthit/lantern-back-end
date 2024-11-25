package chplendpointquerier

import (
	"context"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

func HcscURLWebscraper(chplURL string, fileToWriteTo string) error {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	timeoutContext, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	var htmlContent string
	var err error

	var fhirEndpoint string

	err = chromedp.Run(timeoutContext,
		chromedp.Navigate(chplURL),
		chromedp.WaitVisible("ul.hcsc-p:last-of-type", chromedp.ByQuery),
		chromedp.OuterHTML(`document.querySelector('ul.hcsc-p:last-of-type')`, &htmlContent, chromedp.ByJSPath))

	if err != nil {
		log.Info(err)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Info(err)
		return err
	}

	fhirEndpoint = doc.Find("b").Text()

	endpointDomPath := `document.querySelector('api-documentation').shadowRoot`
	endpointDomPath += `.querySelector('api-summary').shadowRoot.`
	endpointDomPath += `querySelector('api-url').shadowRoot.querySelector('.url-value')`

	err = chromedp.Run(timeoutContext,
		chromedp.Navigate(chplURL),
		chromedp.WaitVisible("api-documentation", chromedp.ByQuery),
		chromedp.OuterHTML(endpointDomPath, &htmlContent, chromedp.ByJSPath))

	if err != nil {
		log.Info(err)
		return err
	}

	doc, err = goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Info(err)
		return err
	}

	fhirEndpoint = strings.TrimSpace(fhirEndpoint)
	fhirEndpoint += strings.Split(doc.Text(), "{environmentBaseUrl}")[1]

	endpointDomPath = `document.querySelector('api-documentation').shadowRoot`
	endpointDomPath += `.querySelector('api-summary').shadowRoot.`
	endpointDomPath += `querySelector('.endpoint-path')`

	err = chromedp.Run(timeoutContext,
		chromedp.Navigate(chplURL),
		chromedp.WaitVisible("api-documentation", chromedp.ByQuery),
		chromedp.OuterHTML(endpointDomPath, &htmlContent, chromedp.ByJSPath))

	if err != nil {
		log.Info(err)
		return err
	}

	doc, err = goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Info(err)
		return err
	}

	fhirEndpoint = strings.TrimSpace(fhirEndpoint)
	fhirEndpoint += doc.Text()

	var entry LanternEntry
	entry.URL = strings.TrimSpace(fhirEndpoint)

	lanternEntryList = append(lanternEntryList, entry)
	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Info(err)
		return err
	}

	return nil
}
