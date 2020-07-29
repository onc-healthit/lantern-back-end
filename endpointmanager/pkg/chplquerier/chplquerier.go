package chplquerier

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var chplDomain string = "https://chpl.healthit.gov"
var chplAPIPath string = "/rest"

// creates the base chpl url using the provided path, a list of query arguments,
// and the chpl api key.
func makeCHPLURL(path string, queryArgs map[string]string) (*url.URL, error) {
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

	chplURL.RawQuery = queryArgsToSend.Encode()
	chplURL.Path = chplAPIPath + path

	return chplURL, nil
}

func getJSON(ctx context.Context, client *http.Client, chplURL *url.URL, userAgent string) ([]byte, error) {
	// request ceritified products list
	req, err := http.NewRequest("GET", chplURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating http request failed")
	}
	req.Header.Set("User-Agent", userAgent)
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "making the GET request to the CHPL server failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("CHPL request responded with status: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading the CHPL response body failed")
	}

	return body, nil
}
