package chplendpointquerier

import (
	"os"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	log "github.com/sirupsen/logrus"
)

func Test_GuidewellURLWebscraper(t *testing.T) {

	// Patient Access API Test Cases
	// 1. Happy case: Valid url, valid file format
	err := GuidewellURLWebscraper("https://developer.bcbsfl.com/interop/interop-developer-portal/product/306/api/285#/CMSInteroperabilityPatientAccessMetadata_100/operation/%2FR4%2Fmetadata/get", "TEST_Medicare_GuidewellPatientAccessEndpointSources.json")

	if err == nil {
		fileExists, err := doesfileExist("TEST_Medicare_GuidewellPatientAccessEndpointSources.json")
		th.Assert(t, err == nil, err)
		th.Assert(t, fileExists, "JSON file does not exist")

		fileEmpty, err := isFileEmpty("TEST_Medicare_GuidewellPatientAccessEndpointSources.json")
		th.Assert(t, err == nil, err)
		th.Assert(t, !fileEmpty, "Empty JSON file")

		err = os.Remove("../../../resources/prod_resources/TEST_Medicare_GuidewellPatientAccessEndpointSources.json")
		th.Assert(t, err == nil, err)

		err = os.Remove("../../../resources/dev_resources/TEST_Medicare_GuidewellPatientAccessEndpointSources.json")
		th.Assert(t, err == nil, err)
	} else {
		log.Info(err)
	}

	// 2. Different file format
	err = GuidewellURLWebscraper("https://developer.bcbsfl.com/interop/interop-developer-portal/product/306/api/285#/CMSInteroperabilityPatientAccessMetadata_100/operation/%2FR4%2Fmetadata/get", "TEST_Medicare_GuidewellPatientAccessEndpointSources.csv")

	if err == nil {
		fileExists, err := doesfileExist("TEST_Medicare_GuidewellPatientAccessEndpointSources.csv")
		th.Assert(t, err == nil, err)
		th.Assert(t, fileExists, "CSV file does not exist")

		fileEmpty, err := isFileEmpty("TEST_Medicare_GuidewellPatientAccessEndpointSources.csv")
		th.Assert(t, err == nil, err)
		th.Assert(t, !fileEmpty, "Empty CSV file")

		err = os.Remove("../../../resources/prod_resources/TEST_Medicare_GuidewellPatientAccessEndpointSources.csv")
		th.Assert(t, err == nil, err)

		err = os.Remove("../../../resources/dev_resources/TEST_Medicare_GuidewellPatientAccessEndpointSources.csv")
		th.Assert(t, err == nil, err)
	} else {
		log.Info(err)
	}

	// Payer2Payer API Test Cases
	// 1. Happy case: Valid url, valid file format
	err = GuidewellURLWebscraper("https://developer.bcbsfl.com/interop/interop-developer-portal/product/309/api/288#/CMSInteroperabilityPayer2PayerOutboundMetadata_100/operation/%2FP2P%2FR4%2Fmetadata/get", "TEST_Medicare_GuidewellPayer2PayerEndpointSources.json")

	if err == nil {
		fileExists, err := doesfileExist("TEST_Medicare_GuidewellPayer2PayerEndpointSources.json")
		th.Assert(t, err == nil, err)
		th.Assert(t, fileExists, "JSON file does not exist")

		fileEmpty, err := isFileEmpty("TEST_Medicare_GuidewellPayer2PayerEndpointSources.json")
		th.Assert(t, err == nil, err)
		th.Assert(t, !fileEmpty, "Empty JSON file")

		err = os.Remove("../../../resources/prod_resources/TEST_Medicare_GuidewellPayer2PayerEndpointSources.json")
		th.Assert(t, err == nil, err)

		err = os.Remove("../../../resources/dev_resources/TEST_Medicare_GuidewellPayer2PayerEndpointSources.json")
		th.Assert(t, err == nil, err)
	} else {
		log.Info(err)
	}

	// 2. Different file format
	err = GuidewellURLWebscraper("https://developer.bcbsfl.com/interop/interop-developer-portal/product/309/api/288#/CMSInteroperabilityPayer2PayerOutboundMetadata_100/operation/%2FP2P%2FR4%2Fmetadata/get", "TEST_Medicare_GuidewellPayer2PayerEndpointSources.csv")

	if err == nil {
		fileExists, err := doesfileExist("TEST_Medicare_GuidewellPayer2PayerEndpointSources.csv")
		th.Assert(t, err == nil, err)
		th.Assert(t, fileExists, "CSV file does not exist")

		fileEmpty, err := isFileEmpty("TEST_Medicare_GuidewellPayer2PayerEndpointSources.csv")
		th.Assert(t, err == nil, err)
		th.Assert(t, !fileEmpty, "Empty CSV file")

		err = os.Remove("../../../resources/prod_resources/TEST_Medicare_GuidewellPayer2PayerEndpointSources.csv")
		th.Assert(t, err == nil, err)

		err = os.Remove("../../../resources/dev_resources/TEST_Medicare_GuidewellPayer2PayerEndpointSources.csv")
		th.Assert(t, err == nil, err)
	} else {
		log.Info(err)
	}
}
