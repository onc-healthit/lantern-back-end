package chplendpointquerier

import (
	"context"
	"encoding/json"
	"io/ioutil"
	http "net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func AllScriptsQuerier(allscriptsURL string, fileToWriteTo string) {

	var endpointEntryList EndpointList

	var DSTU2URL string

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string

	// Chromedp will wait for webpage to run javascript code to generate api search results before grapping HTML
	err := chromedp.Run(ctx,
		chromedp.Navigate(allscriptsURL),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("a").Each(func(index int, linkhtml *goquery.Selection) {
		href, _ := linkhtml.Attr("href")
		DSTU2URL = strings.TrimSpace(href)
	})

	client := &http.Client{}
	req, err := http.NewRequest("GET", allscriptsURL+"/"+strings.Join(strings.Split(DSTU2URL, "/")[2:], "/"), nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	endpointEntryList.Endpoints = BundleToLanternFormat(respBody)
	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
