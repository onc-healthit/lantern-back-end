package chplendpointquerier

import (
	"io"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"os"

	log "github.com/sirupsen/logrus"
)

func StreamlineMDCSVParser(CHPLURL string, fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	csvFilePath := "./Practice_Service_based_URL_List.csv"

	csvReader, file, err := helpers.QueryAndOpenCSV(CHPLURL, csvFilePath, true)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Warnf("error closing file: %v", err)
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

		organizationName := strings.TrimSpace(rec[1])
		URL := strings.TrimSpace(rec[2])

		entry.OrganizationName = organizationName
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
