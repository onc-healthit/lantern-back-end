package chplquerier

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var chplDomain string = "https://chpl.healthit.gov"
var chplAPIPath string = "/rest"

// creates the base chpl url using the provided path, a list of query arguments,
// and the chpl api key.
func makeCHPLURL(path string, queryArgs map[string]string, pageSize int, pageNumber int) (*url.URL, error) {
	queryArgsToSend := url.Values{}
	chplURL, err := url.Parse(chplDomain)
	if err != nil {
		return nil, err
	}

	apiKey := viper.GetString("chplapikey")
	if apiKey == "" {
		return nil, fmt.Errorf("the CHPL API Key is not set")
	}
	queryArgsToSend.Set("api_key", apiKey)
	for k, v := range queryArgs {
		queryArgsToSend.Set(k, v)
	}
	if pageSize != -1 && pageNumber != -1 {
		queryArgsToSend.Set("pageSize", strconv.Itoa(pageSize))
		queryArgsToSend.Set("pageNumber", strconv.Itoa(pageNumber))
	}

	chplURL.RawQuery = queryArgsToSend.Encode()
	chplURL.Path = chplAPIPath + path

	return chplURL, nil
}

func getJSON(ctx context.Context, client *http.Client, chplURL *url.URL, userAgent string) ([]byte, error) {
	// request ceritified products list
	// Adds a short delay between request
	// LANTERN-721: Increased the delay to 1.2 seconds since the CHPL API rate limit is one request per second
	time.Sleep(time.Duration(5000 * time.Millisecond))
	req, err := http.NewRequest("GET", chplURL.String(), nil)
	if err != nil {
		log.Errorf("CHPL ERROR: Failed to create request for URL=%s | Error=%v", chplURL.String(), err)
		return nil, errors.Wrap(err, "creating http request failed")
	}

	req.Header.Set("User-Agent", userAgent)
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("CHPL NETWORK ERROR: GET request failed for URL=%s | Error=%v", chplURL.String(), err)

		// DNS failures
		if strings.Contains(err.Error(), "no such host") {
			log.Errorf("CHPL DNS ERROR: Host could not be resolved for URL=%s", chplURL.String())
		}

		// Timeout
		if strings.Contains(err.Error(), "timeout") {
			log.Errorf("CHPL TIMEOUT: Request timed out for URL=%s", chplURL.String())
		}

		// Connection failures
		if strings.Contains(err.Error(), "refused") || strings.Contains(err.Error(), "reset") {
			log.Errorf("CHPL CONNECTION ERROR: Could not connect to URL=%s", chplURL.String())
		}

		return nil, errors.Wrap(err, "making the GET request to the CHPL server failed")
	}
	defer resp.Body.Close()

	// Server returned non-200
	if resp.StatusCode != http.StatusOK {
		log.Errorf("CHPL SERVER ERROR: URL=%s responded with status %s", chplURL.String(), resp.Status)
		return nil, errors.New("CHPL request responded with status: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("CHPL READ ERROR: Could not read response body for URL=%s | Error=%v", chplURL.String(), err)
		return nil, errors.Wrap(err, "reading the CHPL response body failed")
	}

	// Detect empty response
	if len(body) == 0 {
		log.Warnf("CHPL EMPTY BODY: URL=%s returned an empty response", chplURL.String())
	}

	trimmed := strings.TrimSpace(string(body))
	// Detect HTML response instead of JSON
	if strings.HasPrefix(trimmed, "<") {
		log.Warnf("CHPL NON-JSON RESPONSE: URL=%s returned HTML instead of JSON.", chplURL.String())
		if len(trimmed) > 50 {
			trimmed = trimmed[:50]
		}
		log.Warnf("CHPL FIRST BYTES: %s", trimmed)
	}

	return body, nil
}
