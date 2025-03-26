package medicaidendpointquerier

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	logrus "github.com/sirupsen/logrus"
)

type EndpointList struct {
	Endpoints []LanternEntry `json:"Endpoints"`
}

type LanternEntry struct {
	URL                 string `json:"URL"`
	OrganizationName    string `json:"OrganizationName"`
	NPIID               string `json:"NPIID"`
	OrganizationZipCode string `json:"OrganizationZipCode"`
}

func QueryMedicaidEndpointList(fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	csvFilePath := "../../../resources/prod_resources/medicaid-state-endpoints.csv"
	csvReader, file, err := QueryAndOpenCSV(csvFilePath, true)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logrus.Warnf("Error closing file: %v", err)
		}
	}()
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		var entry LanternEntry

		organizationName := strings.TrimSpace(rec[0])
		URL := strings.TrimSpace(rec[1])

		entry.OrganizationName = organizationName
		entry.URL = URL

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList
	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}

func QueryAndOpenCSV(csvFilePath string, header bool) (*csv.Reader, *os.File, error) {

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

// WriteCHPLFile writes the given endpointEntryList to a json file and stores it in the prod resources directory
func WriteCHPLFile(endpointEntryList EndpointList, fileToWriteTo string) error {
	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		return err
	}

	if len(endpointEntryList.Endpoints) > 10 {
		endpointEntryList.Endpoints = endpointEntryList.Endpoints[0:10]
	}

	reducedFinalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile("../../../resources/dev_resources/"+fileToWriteTo, reducedFinalFormatJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}

func URLsEqual(chplURL string, savedURL string) bool {
	savedURLNorm := strings.TrimSuffix(savedURL, "/")
	chplURLNorm := strings.TrimSuffix(chplURL, "/")

	savedURLNorm = strings.TrimPrefix(savedURLNorm, "https://")
	chplURLNorm = strings.TrimPrefix(chplURLNorm, "https://")
	savedURLNorm = strings.TrimPrefix(savedURLNorm, "http://")
	chplURLNorm = strings.TrimPrefix(chplURLNorm, "http://")

	return savedURLNorm == chplURLNorm
}
