package helpers

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// MaxResponseBytes is the maximum size of an HTTP response body that will be read.
const MaxResponseBytes = 250 * 1024 * 1024

// ReadAllLimited reads r until EOF like io.ReadAll, but returns an error if more than
// MaxResponseBytes are available instead of reading them into memory.
func ReadAllLimited(r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, MaxResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(b) > MaxResponseBytes {
		return nil, fmt.Errorf("response body exceeds the %d byte limit", int64(MaxResponseBytes))
	}
	return b, nil
}

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

// ChromedpQueryEndpointList queries the given endpoint list using chromedp and returns the html document
func ChromedpQueryEndpointList(endpointListURL string, waitVisibleElement string) (*goquery.Document, error) {

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

// QueryEndpointList queries the given endpoint list using http client and returns the response body of the GET request
func QueryEndpointList(endpointListURL string) ([]byte, error) {
	client := &http.Client{Timeout: time.Minute}
	req, err := http.NewRequest("GET", endpointListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint list URL %q: %w", endpointListURL, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	respBody, err := ReadAllLimited(res.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

// QueryEndpointListWithTLSOption queries the given endpoint list with an option to skip TLS verification
func QueryEndpointListWithTLSOption(endpointListURL string, skipTLSVerify bool) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", endpointListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint list URL %q: %w", endpointListURL, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	respBody, err := ReadAllLimited(res.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

func QueryAndReadFile(URL string, filePath string) ([]byte, error) {

	err := downloadFile(filePath, URL)
	if err != nil {
		return nil, err
	}

	// read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func QueryAndOpenCSV(csvURL string, csvFilePath string, header bool) (*csv.Reader, *os.File, error) {

	err := downloadFile(csvFilePath, csvURL)
	if err != nil {
		return nil, nil, err
	}

	// open file
	f, err := os.Open(csvFilePath)
	if err != nil {
		return nil, nil, err
	}

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	csvReader.Comma = ','       // Set the delimiter (default is ',')
	csvReader.LazyQuotes = true // Enable handling of lazy quotes

	if header {
		// Read first line to skip over headers
		_, err = csvReader.Read()
		if err != nil {
			return nil, f, err
		}
	}

	return csvReader, f, nil
}

func downloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file, refusing anything larger than MaxResponseBytes
	n, err := io.Copy(out, io.LimitReader(resp.Body, MaxResponseBytes+1))
	if err != nil {
		return err
	}
	if n > MaxResponseBytes {
		os.Remove(filepath)
		return fmt.Errorf("response from %s exceeds the %d byte limit", url, int64(MaxResponseBytes))
	}
	return nil
}
