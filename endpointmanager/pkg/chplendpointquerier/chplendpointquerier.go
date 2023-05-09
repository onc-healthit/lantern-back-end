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
	URL                 string `json:"URL"`
	OrganizationName    string `json:"OrganizationName"`
	NPIID               string `json:"NPIID"`
	OrganizationZipCode string `json:"OrganizationZipCode"`
}

var MedHostURL = "https://api.mhdi10xasayd.com/medhost-developer-composition/v1/fhir-base-urls.json"
var NextGenURL = "https://nextgen.com/api/practice-search/"
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
var cernerGitHubURL = "https://github.com/cerner/ignite-endpoints"
var cernerSoarianR4URL = "https://github.com/cerner/ignite-endpoints/blob/main/soarian_patient_r4_endpoints.json"
var techCareURL = "https://devportal.techcareehr.com/Serviceurls"
var carefluenceURL = "https://carefluence.com/carefluence-fhir-endpoints/"
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
var goldblattURL = "https://www.goldblattsystems.com/apis"
var cyfluentURL = "https://app.swaggerhub.com/apis-docs/Cyfluent/ProviderPortalApi/3.3#/FHIR/fhir"
var meridianURL = "https://api-datamanager.carecloud.com:8081/fhirurl"
var qualifactsInsyncURL = "https://qualifacts.com/api-page/platform/insync/insync-fhir-org-list.html"
var qualifactsCredibleURL = "https://qualifacts.com/api-page/_downloads/credible-fhir-org-list.json"
var medinfoengineeringURL = "https://docs.webchartnow.com/resources/system-specifications/fhir-application-programming-interface-api/endpoints/"
var relimedsolutionsURL = "https://help.relimedsolutions.com/fhir/fhir-service-urls.csv"
var eclinicalworksURL = "https://fhir.eclinicalworks.com/ecwopendev"
var integraconnectURL = "https://www.integraconnect.com/certifications/"
var streamlinemdURL = "https://patientportal.streamlinemd.com/FHIRReg/Practice%20Service%20based%20URL%20List.csv"
var bridgepatientportalURL = "https://bridgepatientportal.docs.apiary.io/#/introduction/fhir-bridge-patient-portal/fhir-endpoints"
var medicalmineURL = "https://www.charmhealth.com/resources/fhir/index.html#api-endpoints"
var microfourURL = "https://oauth.patientwebportal.com/Fhir/Documentation#serviceBaseUrls"
var magilenenterprisesURL = "https://www.qsmartcare.com/api-documentation.html"
var interopxURL = "https://demo.interopx.com/ix-auth-server/#/endpoints"
var mphrxURL = "https://www.mphrx.com/fhir-service-base-url-directory/"
var correctekURL = "https://ulrichmedicalconcepts.com/home/the-ehr/meaningful-use/disclosure-and-transparency/"
var varianmedicalURL = "https://variandev.dynamicfhir.com/"
var caretrackerURL = "https://hag-fhir.amazingcharts.com/ac/endpoints"
var zhhealthcareURL = "https://blueehr.com/fhir-urls/"
var emedpracticeURL = "https://emedpractice.com/Fhir/FhirHelpDocument.html"
var modernizingmedicineURL = "https://mm-fhir-endpoint-display.qa.fhir.ema-api.com/"
var doc_torURL = "https://hag-fhir.amazingcharts.com/pc/endpoints"
var azaleahealthURL = "https://api.azaleahealth.com/fhir/R4/Endpoint"
var cloudcraftURL = "https://fhirapitest.naiacorp.net/fhir/r4/endpoints/"
var darenasolutionsURL = "https://hub.meldrx.com"
var glenwoodsystemsURL = "https://static.glaceemr.com/endpoints/urls.json"
var practicefusionURL = "https://www.practicefusion.com/assets/static_files/ServiceBaseURLs.json"
var universalEHRURL = "https://appstudio.interopengine.com/partner/fhirR4endpoints-universalehr.json"
var welligentURL = "https://mu3test.welligent.com/fhir/r4/endpoints/"
var astronautURL = "https://astronautehr.com/index.php/fhir-base-urls/"
var bestpracticesacademyURL = "https://ipatientcare.com/onc-acb-certified-2015-edition"
var californiamedicalsystemsURL = "https://cal-med.com/fhir/Fhir-base-urls.csv"
var claimpowerURL = "https://www.claimpowerehr.com/2015ECURES/documents/CP_FHIR_URLS.csv"
var dextersolutionsURL = "https://img1.wsimg.com/blobby/go/f698f3eb-0d14-4f25-a21e-9ac5944696fe/downloads/ezdocs-fhir-base-urls.csv"
var mendelsonURL = "https://orthoplex.mkoss.com/Fhirdocs"
var netsmarttechnologiesURL = "https://careconnect-uat.netsmartcloud.com/baseUrls/"
var patagoniahealthURL = "https://patagoniahealth.com/wp-content/uploads/2022/12/fhir-base-urls.csv"
var webedoctorURL = "https://www.webedoctor.com/docs/fhir-base-urls.csv"
var medicscloudURL = "https://staging.medicscloud.com/MCExtAPI/FHIRMedicsCloud.htm"
var advancedmdURL = "https://developer.advancedmd.com/fhir/base-urls"
var agasthaURL = "http://agastha.com/production-links.html"
var allegiancemdURL = "https://fhir.allegiancemd.io/R4/"
var elationURL = "https://elationfhir.readme.io/reference/service-base-urls"
var betterdayhealthURL = "https://betterdayhealth.net/fhir-docs"
var carecloudURL = "https://api-datamanager.carecloud.com/"
var ethizoURL = "https://fhir-api.ethizo.com/#55b1b3d2-fd9a-4afa-8d17-5bf78943702d"
var hmsfirstURL = "https://fhir-api.hmsfirst.com/r4/EndPoints"
var praxisemrURL = "https://www.praxisemr.com/applicationaccess/api/help/"
var escribeHOSTURL = "https://ehr.escribe.com/ehr/api/fhir"
var mdlogicEHRURL = "https://www.mdlogic.com/solutions/standard-api-documentation"
var altheaURL = "https://altheafhir.mdsynergy.com"
var webchartnowURL = "https://docs.webchartnow.com/resources/system-specifications/fhir-application-programming-interface-api/endpoints/"
var medifusionURL = "https://docs.medifusion.com/"
var smartemrURL = "https://smartemr.readme.io/reference/getting-started#base-url"
var tebraURL = "https://fhir.prd.cloud.tebra.com/fhir-request/swagger-ui/"

var landmarkhealthURL = "https://lmdmzprodws.landmarkhealth.org/docs/fhir-base-urls.csv"
var nthtechnologyURL = "https://admin.nthtechnology.com/fhir_endpoints.php/json"
var netsmartURL = "https://careconnect-uat.netsmartcloud.com/"

func QueryCHPLEndpointList(chplURL string, fileToWriteTo string) {

	if URLsEqual(chplURL, MedHostURL) {
		MedHostQuerier(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, NextGenURL) {
		NextGenwebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, CanvasURL) {
		Canvaswebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, AlteraURL) {
		AlteraQuerier(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, EpicURL) {
		EpicQuerier(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, MeditechURL) {
		MeditechWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, DocsAthenaURL) {
		AthenaCSVParser("https://fhir.athena.io/athena-fhir-urls/athenanet-fhir-base-urls.csv", fileToWriteTo)
	} else if URLsEqual(chplURL, MyDataAthenaURL) {
		Athenawebscraper("https://mydata.athenahealth.com/aserver", fileToWriteTo)
	} else if URLsEqual(chplURL, OneMedicalURL) {
		oneMedicalWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, unifyURL) {
		UnifyWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, trimedtechURL) {
		TriMedTechWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, trimedtechv8URL) {
		TriMedTechV8Webscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, cernerSoarianR4URL) {
		chplURL = strings.ReplaceAll(chplURL, "github.com", "raw.githubusercontent.com")
		chplURL = strings.Replace(chplURL, "/blob", "", 1)
		BundleQuerierParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, cernerGitHubURL) {
		CernerBundleParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, techCareURL) {
		Techcarewebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, carefluenceURL) {
		CarefluenceWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, bizmaticsURL) {
		BundleQuerierParser("https://prognocis.com/fhir/FHIR_FILES/fhirtest.json", fileToWriteTo)
	} else if URLsEqual(chplURL, assureCareURL) {
		CSVParser("https://ipatientcare.com/wp-content/uploads/2022/10/fhir-base-urls.csv", fileToWriteTo, "./fhir-base-urls.csv", 1, 2, true, 1, -1)
	} else if URLsEqual(chplURL, practiceSuiteURL) {
		PracticeSuiteWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, indianHealthServiceURL) {
		IndianHealthWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, geniusSolutionsURL) {
		GeniusSolutionsWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, intelichartURL) {
		IntelichartWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, healthCare2000URL) {
		HealthCare2000SVParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, firstInsightURL) {
		FirstInsightBundleParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, healthSamuraiURL) {
		HealthSamuraiWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, triarqURL) {
		TRIARQPracticeWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, goldblattURL) {
		BundleQuerierParser("https://fhir-test.csn.health/gs-fhir-domain-server/public-base-service-endpoints.json", fileToWriteTo)
	} else if URLsEqual(chplURL, cyfluentURL) {
		SwaggerUIWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, meridianURL) {
		MeridianWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, qualifactsInsyncURL) {
		QualifactsWebscraper("https://qualifacts.com/api-page/_downloads/insync-fhir-org-list.json", fileToWriteTo)
	} else if URLsEqual(chplURL, qualifactsCredibleURL) {
		QualifactsWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, medinfoengineeringURL) {
		MedicalInformaticsEngineeringWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, relimedsolutionsURL) {
		CSVParser(chplURL, fileToWriteTo, "./fhir_service_urls.csv", 1, 3, true, 1, -1)
	} else if URLsEqual(chplURL, eclinicalworksURL) {
		eClinicalWorksBundleParser("https://fhir.eclinicalworks.com/ecwopendev/external/practiceList", fileToWriteTo)
	} else if URLsEqual(chplURL, integraconnectURL) {
		IntegraConnectWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, streamlinemdURL) {
		StreamlineMDCSVParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, bridgepatientportalURL) {
		BridgePatientPortalWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, medicalmineURL) {
		MedicalMineWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, microfourURL) {
		MicroFourWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, magilenenterprisesURL) {
		MagilenEnterprisesWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == interopxURL {
		InteropxWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == mphrxURL {
		SwaggerUIWebscraper("https://atdevsandbox.mphrx.com/", fileToWriteTo)
	} else if chplURL == correctekURL {
		CorrecTekWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == varianmedicalURL {
		VarianMedicalWebscraper("https://variandev.dynamicfhir.com/dhit/basepractice/r4/Home/ApiDocumentation", fileToWriteTo)
	} else if chplURL == caretrackerURL {
		BundleQuerierParser("https://hag-fhir.amazingcharts.com/ac/endpoints/r4", fileToWriteTo)
	} else if chplURL == zhhealthcareURL {
		ZHHealthcareWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == medinfoengineeringURL {
		MedicalInformaticsEngineeringWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == emedpracticeURL {
		eMedPracticeWebscraper("https://servicebackup.emedpractice.com:8443/helpdoc/fhir_helpdoc.html", fileToWriteTo)
	} else if chplURL == modernizingmedicineURL {
		ModernizingMedicineQuerier("qa.fhir.ema-api.com/fhir/r4/Endpoint?connection-type=hl7-fhir-rest", fileToWriteTo)
	} else if chplURL == doc_torURL {
		BundleQuerierParser(chplURL+"/r4", fileToWriteTo)
	} else if URLsEqual(chplURL, azaleahealthURL) {
		BundleQuerierParser(chplURL+"?_format=application/json", fileToWriteTo)
	} else if URLsEqual(chplURL, cloudcraftURL) {
		BundleQuerierParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, darenasolutionsURL) {
		BundleQuerierParser("https://api.meldrx.com/Directories/fhir/endpoints", fileToWriteTo)
	} else if URLsEqual(chplURL, glenwoodsystemsURL) {
		BundleQuerierParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, practicefusionURL) {
		BundleQuerierParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, universalEHRURL) {
		BundleQuerierParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, welligentURL) {
		BundleQuerierParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, astronautURL) {
		CSVParser("https://astronautehr.com/wp-content/uploads/2022/12/Astronaut-fhir-base-urls.csv", fileToWriteTo, "./astronaut_fhir_base_urls.csv", 1, 2, true, 1, -1)
	} else if URLsEqual(chplURL, bestpracticesacademyURL) {
		CSVParser("https://ipatientcare.com/wp-content/uploads/2022/10/fhir-base-urls.csv", fileToWriteTo, "./fhir_base_urls.csv", 1, 2, true, 1, -1)
	} else if URLsEqual(chplURL, californiamedicalsystemsURL) {
		CSVParser(chplURL, fileToWriteTo, "./fhir_base_urls.csv", 1, 0, true, 1, -1)
	} else if URLsEqual(chplURL, claimpowerURL) {
		CSVParser(chplURL, fileToWriteTo, "./cp_fhir_urls.csv", 1, 2, true, 1, -1)
	} else if URLsEqual(chplURL, mendelsonURL) {
		CSVParser("https://orthoplex.mkoss.com/FhirDocs/DownloadCSV", fileToWriteTo, "./baseurl.csv", 1, 0, true, 1, 0)
	} else if URLsEqual(chplURL, patagoniahealthURL) {
		CSVParser(chplURL, fileToWriteTo, "./fhir_base_urls.csv", 1, 2, true, 1, -1)
	} else if URLsEqual(chplURL, webedoctorURL) {
		CSVParser(chplURL, fileToWriteTo, "./fhir_base_urls.csv", 1, 2, true, 1, -1)
	} else if URLsEqual(chplURL, dextersolutionsURL) {
		CSVParser(chplURL, fileToWriteTo, "./ezdocs_fhir_base_urls.csv", 1, 0, true, 3, 1)
	} else if URLsEqual(chplURL, netsmarttechnologiesURL) {
		CSVParser(chplURL, fileToWriteTo, "./fhir_base_urls.csv", -1, 0, false, 1, 0)
	} else if URLsEqual(chplURL, medicscloudURL) {
		MedicsCloudWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, advancedmdURL) {
		AdvancedMdWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, agasthaURL) {
		AgasthaWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, allegiancemdURL) {
		AllegianceMDWebscraper("https://fhir.allegiancemd.io/R4/swagger-ui/", fileToWriteTo)
	} else if URLsEqual(chplURL, elationURL) {
		ElationWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, betterdayhealthURL) {
		BetterdayHealthWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, carecloudURL) {
		CareCloudWebscraper("https://api-datamanager.carecloud.com/fhirurl", fileToWriteTo)
	} else if URLsEqual(chplURL, ethizoURL) {
		EthizoWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, hmsfirstURL) {
		HMSfirstWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, praxisemrURL) {
		PraxisEMRWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, escribeHOSTURL) {
		EscribeHOSTWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, mdlogicEHRURL) {
		MDLogicEHRWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, altheaURL) {
		AltheaWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, webchartnowURL) {
		WebchartNowWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, medifusionURL) {
		MedifusionWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, smartemrURL) {
		SmarteMRWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, tebraURL) {
		TebraWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, landmarkhealthURL) {
		CSVParser(chplURL, fileToWriteTo, "./landmark-fhir-base-urls.csv", 1, 2, true, 1, -1)
	} else if URLsEqual(chplURL, landmarkhealthURL) {
		LandmarkHealthCSVParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, nthtechnologyURL) {
		BundleQuerierParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, netsmartURL) {
		NetsmartCSVParser("https://careconnect-uat.netsmartcloud.com/baseUrls", fileToWriteTo)
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

func URLsEqual(chplURL string, savedURL string) bool {
	savedURLNorm := strings.TrimSuffix(savedURL, "/")
	chplURLNorm := strings.TrimSuffix(chplURL, "/")

	savedURLNorm = strings.TrimPrefix(savedURLNorm, "https://")
	chplURLNorm = strings.TrimPrefix(chplURLNorm, "https://")
	savedURLNorm = strings.TrimPrefix(savedURLNorm, "http://")
	chplURLNorm = strings.TrimPrefix(chplURLNorm, "http://")

	return savedURLNorm == chplURLNorm
}
