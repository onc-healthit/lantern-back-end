package chplendpointquerier

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func NovomediciURLWebscraper(CHPLURL string, fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	csvFilePath := "./novomedic-fhir-base-urls.csv"
	csvReader, file, err := helpers.QueryAndOpenCSV(CHPLURL, csvFilePath, true)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Warnf("Error closing file: %v", err)
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
		orgName := strings.TrimSpace(rec[0])
		if !strings.HasPrefix(orgName, "Prod FHIR ") {
			continue
		}
		URL := strings.TrimSpace(rec[1])
		if strings.TrimSpace(URL) == "" {
			continue
		}

		//entry.OrganizationName = orgName
		entry.URL = URL
		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(csvFilePath)
	if err != nil {
		log.Fatal(err)
	}
}
