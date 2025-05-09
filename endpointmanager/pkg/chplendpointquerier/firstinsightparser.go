package chplendpointquerier

import (
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func FirstInsightBundleParser(CHPLURL string, fileToWriteTo string) {
	var maximEyesEHRURL = "https://fhirdehr.maximeyes.com/api/maxsql/R4/PracticeBundle"
	var maximEyesCOMURL = "https://fhirwehr.maximeyes.com/api/maximeyes.com/R4/PracticeBundle"

	ehrFilePath := "MaximEyesPracticeBundleEHR.json"

	comFilePath := "MaximEyesPracticeBundleCOM.json"

	var endpointEntryList EndpointList

	respBodyEHR, err := helpers.QueryAndReadFile(maximEyesEHRURL, ehrFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// convert bundle data to lantern format
	EHRLanternFormat := BundleToLanternFormat(respBodyEHR, CHPLURL)

	respBodyCOM, err := helpers.QueryAndReadFile(maximEyesCOMURL, comFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// convert bundle data to lantern format
	COMLanternFormat := BundleToLanternFormat(respBodyCOM, CHPLURL)

	endpointEntryList.Endpoints = append(EHRLanternFormat, COMLanternFormat...)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(ehrFilePath)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(comFilePath)
	if err != nil {
		log.Fatal(err)
	}
}
