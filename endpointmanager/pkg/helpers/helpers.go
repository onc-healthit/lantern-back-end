package helpers

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// StringArrayContains checks if the string array contains the provided string.
func StringArrayContains(l []string, s string) bool {
	for _, s2 := range l {
		if s == s2 {
			return true
		}
	}
	return false
}

// IntArrayContains checks if the integer array contains the provided integer.
func IntArrayContains(l []int, i int) bool {
	for _, i2 := range l {
		if i == i2 {
			return true
		}
	}
	return false
}

// StringArraysEqual checks if l1 and l2 have the same contents regardless of order.
func StringArraysEqual(l1 []string, l2 []string) bool {
	if len(l1) != len(l2) {
		return false
	}

	// don't care about order
	a := make([]string, len(l1))
	b := make([]string, len(l2))
	copy(a, l1)
	copy(b, l2)
	sort.Strings(a)
	sort.Strings(b)
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// IntArraysEqual checks if l1 and l2 have the same contents regardless of order.
func IntArraysEqual(l1 []int, l2 []int) bool {
	if len(l1) != len(l2) {
		return false
	}

	// don't care about order
	a := make([]int, len(l1))
	b := make([]int, len(l2))
	copy(a, l1)
	copy(b, l2)
	sort.Ints(a)
	sort.Ints(b)
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// FailOnError checks if err is not equal to nil and if it isn't, logs failure and exits the program
func FailOnError(errString string, err error) {
	if err != nil {
		if errString == "" {
			log.Fatalf("%s", err)
		} else {
			log.Fatalf("%s %s", errString, err)
		}
	}
}

func ChromedpQueryEndpointList(endpointListURL string, waitVisibleElement string) *goquery.Document {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string
	var err error

	if len(waitVisibleElement) > 0 {
		// Chromedp will wait for webpage to run javascript code to generate api search results before grapping HTML
		err = chromedp.Run(ctx,
			chromedp.Navigate(endpointListURL),
			chromedp.WaitVisible(waitVisibleElement, chromedp.ByQuery),
			chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
		)
	} else {
		err = chromedp.Run(ctx,
			chromedp.Navigate(endpointListURL),
			chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
		)
	}

	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func QueryEndpointList(endpointListURL string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("GET", endpointListURL, nil)
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
	return respBody
}
