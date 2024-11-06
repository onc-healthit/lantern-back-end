package chplendpointquerier

import (
	"os"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	log "github.com/sirupsen/logrus"
)

type WebScraperFunc func(string, string)

type ScraperTestCase struct {
	scraperFunc WebScraperFunc
	url         string
	fileName    string
}

func TestWebScrapers(t *testing.T) {
	log.Info("common web scraper test file")
	testCases := []ScraperTestCase{
		{
			scraperFunc: AspMDeWebscraper,
			url:         "https://fhirapi.asp.md:3030/aspmd/fhirserver/fhir_aspmd.asp",
			fileName:    "ASPMD_Inc_EndpointSources.json",
		},
		{
			scraperFunc: OntadaWebscraper,
			url:         "https://g2fhir-int.mckesson.com/docs/index.html",
			fileName:    "McKesson_Specialty_Health_Technology_Products_LLC_Ontada_EndpointSources.json",
		},
		{
			scraperFunc: MdlandWebscraper,
			url:         "https://api.mdland.com/Mdland%20SMART%20on%20FHIR%20OAuth%202.0%20Guide.htm",
			fileName:    "MDLAND_EndpointSources.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.fileName, func(t *testing.T) {
			runWebScraperTest(t, tc.scraperFunc, tc.url, tc.fileName)
		})
	}
}

func runWebScraperTest(t *testing.T, scraperFunc WebScraperFunc, url, fileName string) {
	scraperFunc(url, fileName)

	fileExists, err := doesfileExist(fileName)
	th.Assert(t, err == nil, err)
	th.Assert(t, fileExists, "File does not exist")

	fileEmpty, err := isFileEmpty(fileName)
	th.Assert(t, err == nil, err)
	th.Assert(t, !fileEmpty, "File is empty")

	err = os.Remove("../../../resources/prod_resources/" + fileName)
	th.Assert(t, err == nil, err)

	err = os.Remove("../../../resources/dev_resources/" + fileName)
	th.Assert(t, err == nil, err)
}
