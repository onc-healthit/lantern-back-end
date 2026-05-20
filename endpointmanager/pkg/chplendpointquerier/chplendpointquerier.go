package chplendpointquerier

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type EndpointList struct {
	Endpoints []LanternEntry `json:"Endpoints"`
}

type LanternEntry struct {
	URL                     string   `json:"URL"`
	OrganizationName        string   `json:"OrganizationName"`
	OrganizationURL         string   `json:"OrganizationURL"`
	NPIID                   string   `json:"NPIID"`
	OrganizationZipCode     string   `json:"OrganizationZipCode"`
	OrganizationIdentifiers []string `json:"OrganizationIdentifiers"`
	OrganizationAddresses   []string `json:"OrganizationAddresses"`
	OrganizationActive      string   `json:"OrganizationActive"`
}

func AppendQueryError(errorFilePath, listSource, errorMsg string) {
	if errorFilePath == "" {
		return
	}

	if err := os.MkdirAll(filepath.Dir(errorFilePath), 0755); err != nil {
		log.Warnf("Failed to create error file directory: %v", err)
		return
	}

	_, statErr := os.Stat(errorFilePath)
	fileExists := statErr == nil

	f, err := os.OpenFile(errorFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Warnf("Failed to write to error file: %v", err)
		return
	}
	defer f.Close()

	if !fileExists {
		f.WriteString("queried_at,list_source,error_message\n")
	}

	escapedMsg := strings.ReplaceAll(errorMsg, "\"", "\"\"")
	escapedSource := strings.ReplaceAll(listSource, "\"", "\"\"")
	line := fmt.Sprintf("%s,\"%s\",\"%s\"\n",
		time.Now().UTC().Format(time.RFC3339),
		escapedSource,
		escapedMsg)
	f.WriteString(line)
}

// State Payer list
var aetnaURL = "https://developerportal.aetna.com/fhirapis"
var centeneURL = "https://partners.centene.com/apiDetail/2718669d-6e2e-42b5-8c90-0a82f13a30ba"
var cignaURL = "https://developer.cigna.com/docs/service-apis/patient-access/implementation-guide#Implementation-Guide-Base-URL"
var anthemURL = "https://patient360.anthem.com/P360Member/fhir"
var hcscURL = "https://interoperability.hcsc.com/s/provider-directory-api"
var guidewellPatAccURL = "https://developer.bcbsfl.com/interop/interop-developer-portal/product/396/api/375#/CMSInteroperabilityPatientAccessMetadata_100/operation/%2FR4%2Fmetadata/get"
var guidewellP2PURL = "https://developer.bcbsfl.com/interop/interop-developer-portal/product/399/api/378#/CMSInteroperabilityPayer2PayerOutboundMetadata_100/operation/%2FP2P%2FR4%2Fmetadata/get"
var humanaURL = "https://developers.humana.com/apis/patient-api/doc"
var kaiserURL = "https://developer.kp.org/#/apis/639c015049655aa96ab5b2f1"

// var molinaURL = "https://developer.interop.molinahealthcare.com/api-details#api=patient-access&operation=5f72ab665269f310ef58b361"
var unitedHealthURL = "https://www.uhc.com/legal/interoperability-apis/patient-access-api"

func QueryCHPLEndpointList(chplURL string, fileToWriteTo string, errorFilePath string) {

	var err error

	if URLsEqual(chplURL, aetnaURL) {
		AetnaURLWebscraper("https://developerportal.aetna.com/fhir/apis/swagger/_v2_patientaccess_Endpoint_id.yaml", fileToWriteTo)
	} else if URLsEqual(chplURL, centeneURL) {
		CenteneURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, cignaURL) {
		CignaURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, anthemURL) {
		AnthemURLParser("https://patient360.anthem.com/P360Member/fhir/endpoints", fileToWriteTo)
	} else if URLsEqual(chplURL, guidewellPatAccURL) || URLsEqual(chplURL, guidewellP2PURL) {
		err = GuidewellURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, hcscURL) {
		err = HcscURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, humanaURL) {
		HumanaURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, kaiserURL) {
		KaiserURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, unitedHealthURL) {
		UnitedHealthURLWebscraper(chplURL, fileToWriteTo)
	} else {
		log.Warnf(
			"CHPL ENDPOINT QUERIER PARSER FALLBACK: No explicit handler matched. Using BundleQuerierParser. url=%s",
			chplURL,
		)
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	}

	if err != nil {
		log.Info(err)
	}
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
