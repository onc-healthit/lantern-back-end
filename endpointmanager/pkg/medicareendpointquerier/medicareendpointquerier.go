package medicaidendpointquerier

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	logrus "github.com/sirupsen/logrus"
)

type EndpointEntry struct {
	FormatType   string `json:"FormatType"`
	URL          string `json:"URL"`
	EndpointName string `json:"EndpointName"`
	FileName     string `json:"FileName"`
}

type EndpointList struct {
	Endpoints []EndpointEntry `json:"Endpoints"`
}

func QueryMedicareEndpointList(fileToWriteTo string) {
	var lanternEntryList []EndpointEntry
	var endpointEntryList EndpointList
	var existingURLs []string

	csvFilePath := "../../../resources/prod_resources/payer-patient-access.csv"
	csvReader, file, err := QueryAndOpenCSV(csvFilePath, true)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logrus.Warnf("error closing file: %v", err)
		}
	}()
	for {
		var entry EndpointEntry
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		developer := strings.TrimSpace(rec[0])
		source := strings.TrimSpace(rec[1])

		developer = strings.ReplaceAll(developer, " ", "")

		if !helpers.StringArrayContains(existingURLs, source) {
			entry.FileName = "Medicare_" + developer + "EndpointSources.json"
			entry.FormatType = "Lantern"
			entry.URL = source
			entry.EndpointName = developer
			lanternEntryList = append(lanternEntryList, entry)
			existingURLs = append(existingURLs, source)
		}
	}

	endpointEntryList.Endpoints = lanternEntryList
	err = WritePayerFile(endpointEntryList, fileToWriteTo)
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
func WritePayerFile(endpointEntryList EndpointList, fileToWriteTo string) error {
	finalFormatJSON, err := json.MarshalIndent(endpointEntryList.Endpoints, "", "\t")
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

	reducedFinalFormatJSON, err := json.MarshalIndent(endpointEntryList.Endpoints, "", "\t")
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
