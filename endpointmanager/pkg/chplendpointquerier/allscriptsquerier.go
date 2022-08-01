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

	// get document from html string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	// find <a> tags and iterateover them
	doc.Find("a").Each(func(index int, linkhtml *goquery.Selection) {
		// select only the 6th link on the page for DSTU2 endpoints
		if index == 5 {
			// get href from link
			href, _ := linkhtml.Attr("href")
			DSTU2URL = strings.TrimSpace(href)
		}
	})

	client := &http.Client{}
	// concatenate dstu2 link onto base allscripts url to get full url and make request
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

	// convert bundle data to lantern format
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
