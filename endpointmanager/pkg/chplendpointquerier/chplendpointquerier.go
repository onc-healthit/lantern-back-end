package chplendpointquerier

import (
	"encoding/json"
	"io/ioutil"
	"strings"
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
var AlteraURL = "https://open.allscripts.com/fhirendpoints"
var EpicURL = "https://open.epic.com/MyApps/Endpoints"
var MeditechURL = "https://fhir.meditech.com/explorer/endpoints"
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
var goldblattURL = "https://www.goldblattsystems.com/apis"
var cyfluentURL = "https://app.swaggerhub.com/apis-docs/Cyfluent/ProviderPortalApi/3.3#/FHIR/fhir"
var mphrxURL = "https://www.mphrx.com/fhir-service-base-url-directory/"
var meridianURL = "https://api-datamanager.carecloud.com:8081/fhirurl"
var qualifactsURL = "https://qualifacts.com/api-documentation/"
var medinfoengineeringURL = "https://docs.webchartnow.com/resources/system-specifications/fhir-application-programming-interface-api/endpoints/"
var relimedsolutionsURL = "https://help.relimedsolutions.com/fhir/fhir-service-urls.csv"
var eclinicalworksURL = "https://fhir.eclinicalworks.com/ecwopendev"
var integraconnectURL = "https://www.integraconnect.com/certifications/"
var streamlinemdURL = "https://patientportal.streamlinemd.com/FHIRReg/Practice%20Service%20based%20URL%20List.csv"
var bridgepatientportalURL = "https://bridgepatientportal.docs.apiary.io/#/introduction/fhir-bridge-patient-portal/fhir-endpoints"
var medicalmineURL = "https://www.charmhealth.com/resources/fhir/index.html#api-endpoints"
var microfourURL = "https://oauth.patientwebportal.com/Fhir/Documentation#serviceBaseUrls"
var magilenenterprisesURL = "https://www.qsmartcare.com/api-documentation.html"

func QueryCHPLEndpointList(chplURL string, fileToWriteTo string) {

	if chplURL == MedHostURL {
		MedHostQuerier(chplURL, fileToWriteTo)
	} else if chplURL == NextGenURL {
		NextGenwebscraper(chplURL, fileToWriteTo)
	} else if chplURL == CanvasURL {
		Canvaswebscraper(chplURL, fileToWriteTo)
	} else if chplURL == AlteraURL {
		AlteraQuerier(chplURL, fileToWriteTo)
	} else if chplURL == EpicURL {
		EpicQuerier(chplURL, fileToWriteTo)
	} else if chplURL == MeditechURL {
		MeditechWebscraper(MeditechURL, fileToWriteTo)
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
		CarefluenceWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == abeoSolutionsURL {
		CSVParser(chplURL, fileToWriteTo, "./FHIRServiceURLs.csv", 1, 0)
	} else if chplURL == bizmaticsURL {
		BundleQuerierParser("https://prognocis.com/fhir/FHIR_FILES/fhirtest.json", fileToWriteTo)
	} else if chplURL == assureCareURL {
		CSVParser("https://ipatientcare.com/wp-content/uploads/2022/10/fhir-base-urls.csv", fileToWriteTo, "./fhir-base-urls.csv", 1, 2)
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
	} else if chplURL == goldblattURL {
		BundleQuerierParser("https://fhir-test.csn.health/gs-fhir-domain-server/public-base-service-endpoints.json", fileToWriteTo)
	} else if chplURL == cyfluentURL {
		SwaggerUIWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == mphrxURL {
		SwaggerUIWebscraper("https://atdevsandbox.mphrx.com/", fileToWriteTo)
	} else if chplURL == meridianURL {
		MeridianWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == qualifactsURL {
		QualifactsWebscraper("https://qualifacts.com/api-page/_downloads/insync-fhir-org-list.json", fileToWriteTo)
	} else if chplURL == medinfoengineeringURL {
		MedicalInformaticsEngineeringWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == relimedsolutionsURL {
		CSVParser(chplURL, fileToWriteTo, "./fhir_service_urls.csv", 1, 3)
	} else if chplURL == eclinicalworksURL {
		eClinicalWorksBundleParser("https://fhir.eclinicalworks.com/ecwopendev/external/practiceList", fileToWriteTo)
	} else if chplURL == integraconnectURL {
		IntegraConnectWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == streamlinemdURL {
		StreamlineMDCSVParser(chplURL, fileToWriteTo)
	} else if chplURL == bridgepatientportalURL {
		BridgePatientPortalWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == medicalmineURL {
		MedicalMineWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == microfourURL {
		MicroFourWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == magilenenterprisesURL {
		MagilenEnterprisesWebscraper(chplURL, fileToWriteTo)
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

	if len(endpointEntryList.Endpoints) > 10 {
		endpointEntryList.Endpoints = endpointEntryList.Endpoints[0:10]
	}

	reducedFinalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("../../../resources/dev_resources/"+fileToWriteTo, reducedFinalFormatJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}
