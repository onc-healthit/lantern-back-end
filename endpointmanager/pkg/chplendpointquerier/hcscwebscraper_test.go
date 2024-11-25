package chplendpointquerier

import (
	"os"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	log "github.com/sirupsen/logrus"
)

func doesfileExist(filename string) (bool, error) {
	_, err := os.Stat("../../../resources/prod_resources/" + filename)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err // File exists or Error occured
}

func isFileEmpty(filename string) (bool, error) {
	data, err := os.ReadFile("../../../resources/prod_resources/" + filename)
	if err != nil {
		return true, err
	}

	return string(data) == "", nil // Convert byte slice to string
}

func Test_HcscURLWebscraper(t *testing.T) {

	log.Info("hcsc test file")
	// 1. Happy case: Valid url, valid file format
	err := HcscURLWebscraper("https://interoperability.hcsc.com/s/provider-directory-api", "TEST_Medicare_HCSCEndpointSources.json")

	if err == nil {
		fileExists, err := doesfileExist("TEST_Medicare_HCSCEndpointSources.json")
		th.Assert(t, err == nil, err)
		th.Assert(t, fileExists, "JSON file does not exist")

		fileEmpty, err := isFileEmpty("TEST_Medicare_HCSCEndpointSources.json")
		th.Assert(t, err == nil, err)
		th.Assert(t, !fileEmpty, "Empty JSON file")

		err = os.Remove("../../../resources/prod_resources/TEST_Medicare_HCSCEndpointSources.json")
		th.Assert(t, err == nil, err)

		err = os.Remove("../../../resources/dev_resources/TEST_Medicare_HCSCEndpointSources.json")
		th.Assert(t, err == nil, err)
	}

	// 2. Empty inputs
	err = HcscURLWebscraper("", "")

	if err == nil {
		fileExists, err := doesfileExist("TEST_Medicare_HCSCEndpointSources.json")
		th.Assert(t, err == nil, err)
		th.Assert(t, !fileExists, "File exists for invalid inputs")

		fileEmpty, err := isFileEmpty("TEST_Medicare_HCSCEndpointSources.json")
		th.Assert(t, err != nil, "File data read successfully for invalid inputs")
		th.Assert(t, fileEmpty, "File contains data for invalid inputs")
	}

	// 3. Different file format
	err = HcscURLWebscraper("https://interoperability.hcsc.com/s/provider-directory-api", "TEST_Medicare_HCSCEndpointSources.csv")

	if err == nil {
		fileExists, err := doesfileExist("TEST_Medicare_HCSCEndpointSources.csv")
		th.Assert(t, err == nil, err)
		th.Assert(t, fileExists, "CSV file does not exist")

		fileEmpty, err := isFileEmpty("TEST_Medicare_HCSCEndpointSources.csv")
		th.Assert(t, err == nil, err)
		th.Assert(t, !fileEmpty, "Empty CSV file")
	}

	// 4. Invalid URL
	err = HcscURLWebscraper("https://non-existent-url.com/dummy-api", "TEST_Medicare_HCSCEndpointSources.json")

	if err == nil {
		fileExists, err := doesfileExist("TEST_Medicare_HCSCEndpointSources.json")
		th.Assert(t, err == nil, err)
		th.Assert(t, !fileExists, "File exists for invalid URL")

		fileEmpty, err := isFileEmpty("TEST_Medicare_HCSCEndpointSources.json")
		th.Assert(t, err != nil, "File read successful for invalid URL")
		th.Assert(t, fileEmpty, "File contains data for invalid URL")
	}
}
