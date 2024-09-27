package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func main() {
	doc, err := QueryEndpointList("http://localhost:8090/?tab=dashboard_tab")
	if err != nil {
		fmt.Print("Error!!!")
		log.Fatal(err)
	}

	if doc != nil {
		fmt.Print("Success!!!")
	}
	doc1, err1 := ChromedpQueryEndpointList("http://localhost:8090/?tab=dashboard_tab", "#httpvendor")
	if err1 != nil {
		fmt.Print("Error!!!")
		log.Fatal(err1)
	}

	if doc1 != nil {
		fmt.Print("Success!!!")
	}

}

func loadPage(url, selector string, resultCh chan<- string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady(selector, chromedp.ByQuery),
		chromedp.EvaluateAsDevTools(fmt.Sprintf(`document.querySelector('%s').innerText`, selector), resultCh),
	}
}

func QueryEndpointList(endpointListURL string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", endpointListURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

func ChromedpQueryEndpointList(endpointListURL string, waitVisibleElement string) (*goquery.Document, error) {

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	timeoutContext, cancel := context.WithTimeout(ctx, 60*time.Minute)
	defer cancel()

	var htmlContent string
	var err error

	if len(waitVisibleElement) > 0 {
		// Chromedp will wait a max of 30 seconds for webpage to run javascript code to generate api search results before grapping HTML
		err = chromedp.Run(timeoutContext,
			chromedp.Navigate(endpointListURL),
			chromedp.WaitVisible(waitVisibleElement, chromedp.ByQuery),
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
