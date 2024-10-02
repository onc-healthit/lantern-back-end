package chplendpointquerier

import (
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
)

func entryExists(lanternEntryList []LanternEntry, lanternEntry LanternEntry) bool {
	for _, entry := range lanternEntryList {
		if entry == lanternEntry {
			return true
		}
	}
	return false
}

func GuidewellURLWebscraper(CHPLURL string, fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	patientAccessChplUrl := "https://developer.bcbsfl.com/interop/interop-developer-portal/product/306/api/285#/CMSInteroperabilityPatientAccessMetadata_100/operation/%2FR4%2Fmetadata/get"
	payer2payerChplUrl := "https://developer.bcbsfl.com/interop/interop-developer-portal/product/309/api/288#/CMSInteroperabilityPayer2PayerOutboundMetadata_100/operation/%2FP2P%2FR4%2Fmetadata/get"
	// providerDirectoryChplUrl := "https://developer.bcbsfl.com/interop/interop-developer-portal/product/530/api/300#/ProviderDirectoryAPI_108/overview"
	CHPLURLs := []string{patientAccessChplUrl, payer2payerChplUrl}

	for i := 0; i < len(CHPLURLs); i++ {
		doc, err := helpers.ChromedpQueryEndpointList(CHPLURLs[i], "div.apiEndpointUrl")
		if err != nil {
			log.Fatal(err)
		}
		doc.Find("div.apiEndpointUrl").Each(func(index int, urlElements *goquery.Selection) {

			var lanternEntry LanternEntry

			fhirURL := urlElements.Text()
			lanternEntry.URL = fhirURL
			if !entryExists(lanternEntryList, lanternEntry) {
				lanternEntryList = append(lanternEntryList, lanternEntry)
			}
		})
	}

	endpointEntryList.Endpoints = lanternEntryList
	err := WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
