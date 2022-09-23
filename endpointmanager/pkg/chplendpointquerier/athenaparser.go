package chplendpointquerier

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"encoding/csv"
	"os"

	log "github.com/sirupsen/logrus"
)

func AthenaCSVParser(CHPLURL string, fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	csvFilePath := "./athenanet-fhir-base-urls.csv"

	err := DownloadFile(csvFilePath, CHPLURL)
	if err != nil {
		log.Fatal(err)
	}

	// open file
	f, err := os.Open(csvFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)

	// Read first line to skip over headers
	_, err = csvReader.Read()
	if err != nil {
		log.Fatal(err)
	}

	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		var entry LanternEntry

		organizationName := strings.TrimSpace(rec[1])
		URL := strings.TrimSpace(rec[3])

		entry.OrganizationName = organizationName
		entry.URL = URL

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList

	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(csvFilePath)
	if err != nil {
		log.Fatal(err)
	}
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
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

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
