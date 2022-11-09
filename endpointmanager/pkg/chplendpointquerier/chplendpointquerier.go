package chplendpointquerier

import (
	"encoding/json"
	"io/ioutil"
)

type EndpointList struct {
	Endpoints []LanternEntry `json:"Endpoints"`
}

type LanternEntry struct {
	URL              string `json:"URL"`
	OrganizationName string `json:"OrganizationName"`
	NPIID            string `json:"NPIID"`
}

var MedHostURL = "https://api.mhdi10xasayd.com/medhost-developer-composition/v1/fhir-base-urls.json"
var NextGenURL = "https://nextgen.com/api/practice-search"
var CanvasURL = "https://docs.canvasmedical.com/reference/service-base-urls"
var AllScriptsURL = "https://open.allscripts.com/fhirendpoints"
var EpicURL = "https://open.epic.com/MyApps/Endpoints"
var MeditechURL = "https://home.meditech.com/en/d/restapiresources/pages/apidoc.htm"
var DocsAthenaURL = "https://docs.athenahealth.com/api/base-fhir-urls"
var MyDataAthenaURL = "https://mydata.athenahealth.com/home"
var OneMedicalURL = "https://apidocs.onemedical.io/fhir/overview/"
var unifyURL = "https://unify-developer.chbase.com/?page=FHIRAPI"
var trimedtechURL = "https://www.trimedtech.com/Documentation/FHIRAPI/FHIRAPI.html"
var trimedtechv8URL = "https://www.trimedtech.com/Documentation/FHIRAPI/V8FHIRAPI.html"
var cernerR4URL = "https://github.com/cerner/ignite-endpoints/blob/main/soarian_patient_r4_endpoints.json"
var techCareURL = "https://devportal.techcareehr.com/Serviceurls"
var carefluenceURL = "https://carefluence.com/carefluence-fhir-endpoints/"
var abeoSolutionsURL = "https://www.crystalpm.com/FHIRServiceURLs.csv"
var practiceSuiteURL = "https://academy.practicesuite.com/fhir-server-links/"
var bizmaticsURL = "https://prognocis.com/fhir/index.html"
var indianHealthServiceURL = "https://www.ihs.gov/cis/"
var geniusSolutionsURL = "https://gsehrwebapi.geniussolutions.com/Help/html/ServiceUrl.html"
var assureCareURL = "https://ipatientcare.com/onc-acb-certified-2015-edition/"
var intelichartURL = "https://fhirtest.intelichart.com/Help/BaseUrl"
var healthCare2000URL = "https://www.provider.care/FHIR/MDVitaFHIRUrls.csv"
var firstInsightURL = "https://www.first-insight.com/maximeyes_fhir_base_url_endpoints/"
var healthSamuraiURL = "https://cmpl.aidbox.app/smart"
var triarqURL = "https://fhir.myqone.com/Endpoints"
var napchareURL = "https://devportal.techcareehr.com/Serviceurls"

func QueryCHPLEndpointList(chplURL string, fileToWriteTo string) {

	if chplURL == MedHostURL {
		MedHostQuerier(chplURL, fileToWriteTo)
	} else if chplURL == NextGenURL {
		NextGenwebscraper(chplURL, fileToWriteTo)
	} else if chplURL == CanvasURL {
		Canvaswebscraper(chplURL, fileToWriteTo)
	} else if chplURL == AllScriptsURL {
		AllScriptsQuerier(chplURL, fileToWriteTo)
	} else if chplURL == EpicURL {
		EpicQuerier(chplURL, fileToWriteTo)
	} else if chplURL == MeditechURL {
		Meditechwebscraper("https://fhir.meditech.com/explorer/endpoints", fileToWriteTo)
	} else if chplURL == DocsAthenaURL {
		AthenaCSVParser("https://fhir.athena.io/athena-fhir-urls/athenanet-fhir-base-urls.csv", fileToWriteTo)
	} else if chplURL == MyDataAthenaURL {
		Athenawebscraper("https://mydata.athenahealth.com/aserver", fileToWriteTo)
	} else if chplURL == OneMedicalURL {
		oneMedicalWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == unifyURL {
		UnifyWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == trimedtechURL {
		TriMedTechWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == trimedtechv8URL {
		TriMedTechV8Webscraper(chplURL, fileToWriteTo)
	} else if chplURL == cernerR4URL {
		chplURL = strings.ReplaceAll(chplURL, "github.com", "raw.githubusercontent.com")
		chplURL = strings.Replace(chplURL, "/blob", "", 1)
		BundleQuerierParser(chplURL, fileToWriteTo)
	} else if chplURL == techCareURL {
		Techcarewebscraper(chplURL, fileToWriteTo)
	} else if chplURL == carefluenceURL {
		Carefluenceebscraper(chplURL, fileToWriteTo)
	} else if chplURL == abeoSolutionsURL {
		AbeoSolutionsCSVParser(chplURL, fileToWriteTo)
	} else if chplURL == bizmaticsURL {
		BundleQuerierParser("https://prognocis.com/fhir/FHIR_FILES/fhirtest.json", fileToWriteTo)
	} else if chplURL == assureCareURL {
		AssureCareCSVParser("https://ipatientcare.com/wp-content/uploads/2022/10/fhir-base-urls.csv", fileToWriteTo)
	} else if chplURL == practiceSuiteURL {
		PracticeSuiteWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == indianHealthServiceURL {
		IndianHealthWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == geniusSolutionsURL {
		GeniusSolutionsWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == intelichartURL {
		IntelichartWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == healthCare2000URL {
		HealthCare2000SVParser(chplURL, fileToWriteTo)
	} else if chplURL == firstInsightURL {
		FirstInsightBundleParser(chplURL, fileToWriteTo)
	} else if chplURL == healthSamuraiURL {
		HealthSamuraiWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == triarqURL {
		TRIARQPracticeWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == napchareURL {
		NaphCareWebscraper(chplURL, fileToWriteTo)
	}
}

// WriteCHPLFile writes the given endpointEntryList to a json file and stores it in the prod resources directory
func WriteCHPLFile(endpointEntryList EndpointList, fileToWriteTo string) error {
	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}
