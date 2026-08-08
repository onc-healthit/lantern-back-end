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
		if _, err := f.WriteString("queried_at,list_source,error_message\n"); err != nil {
			log.Warnf("Failed to write to error file: %v", err)
			return
		}
	}

	escapedMsg := strings.ReplaceAll(errorMsg, "\"", "\"\"")
	escapedSource := strings.ReplaceAll(listSource, "\"", "\"\"")
	line := fmt.Sprintf("%s,\"%s\",\"%s\"\n",
		time.Now().UTC().Format(time.RFC3339),
		escapedSource,
		escapedMsg)
	if _, err := f.WriteString(line); err != nil {
		log.Warnf("Failed to write to error file: %v", err)
	}
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

var MedHostURL = "https://api.mhdi10xasayd.com/medhost-developer-composition/v1/fhir-base-urls.json"
var NextGenURL = "https://nextgen.com/api/practice-search"
var CanvasURL = "https://docs.canvasmedical.com/reference/service-base-urls"
var MeditechURL = "https://fhir.meditech.com/explorer/endpoints"
var DocsAthenaURL = "https://docs.athenahealth.com/api/base-fhir-urls"
var MyDataAthenaURL = "https://mydata.athenahealth.com/home"
var OneMedicalURL = "https://apidocs.onemedical.io/fhir/overview/"
var unifyURL = "https://unify-developer.chbase.com/?page=FHIRAPI"
var trimedtechURL = "https://www.trimedtech.com/Documentation/FHIRAPI/FHIRAPI.html"
var trimedtechv8URL = "https://www.trimedtech.com/Documentation/FHIRAPI/V8FHIRAPI.html"
var cernerSoarianR4URL = "https://github.com/cerner/ignite-endpoints/blob/main/soarian_patient_r4_endpoints.json"

// var techCareURL = "https://devportal.techcareehr.com/Serviceurls"
var carefluenceURL = "https://carefluence.com/carefluence-fhir-endpoints/"
var practiceSuiteURL = "https://academy.practicesuite.com/fhir-server-links/"
var bizmaticsURL = "https://prognocis.com/fhir/index.html"
var indianHealthServiceURL = "https://www.ihs.gov/cis/"
var geniusSolutionsURL = "http://www.media.geniussolutions.com/ehrTHOMAS/ehrWebApi/Help/html/ServiceUrl.html"
var assureCareURL = "https://ipatientcare.com/onc-acb-certified-2015-edition/"
var intelichartURL = "https://fhirtest.intelichart.com/Help/BaseUrl"
var healthCare2000URL = "https://www.provider.care/FHIR/MDVitaFHIRUrls.csv"
var healthSamuraiURL = "https://cmpl.aidbox.app/smart"
var triarqURL = "https://fhir.myqone.com/Endpoints"
var cyfluentURL = "https://app.swaggerhub.com/apis-docs/Cyfluent/ProviderPortalApi/3.3#/FHIR/fhir"
var meridianURL = "https://api-datamanager.carecloud.com:8081/fhirurl"
var qualifactsInsyncURL = "https://qualifacts.com/api-page/platform/insync/insync-fhir-org-list.html"
var qualifactsCredibleURL = "https://qualifacts.com/api-page/_downloads/credible-fhir-org-list.json"
var medinfoengineeringURL = "https://docs.webchartnow.com/resources/system-specifications/fhir-application-programming-interface-api/endpoints/"
var relimedsolutionsURL = "https://help.relimedsolutions.com/fhir/fhir-service-urls.csv"
var streamlinemdURL = "https://patientportal.streamlinemd.com/FHIRReg/Practice%20Service%20based%20URL%20List.csv"
var bridgepatientportalURL = "https://bridgepatientportal.docs.apiary.io/#/introduction/fhir-bridge-patient-portal/fhir-endpoints"
var medicalmineURL = "https://www.charmhealth.com/resources/fhir/index.html#api-endpoints"
var microfourURL = "https://oauth.patientwebportal.com/Fhir/Documentation#serviceBaseUrls"
var magilenenterprisesURL = "https://www.qsmartcare.com/api-documentation.html"
var interopxURL = "https://demo.interopx.com/ix-auth-server/#/endpoints"
var mphrxURL = "https://www.mphrx.com/fhir-service-base-url-directory/"
var varianmedicalURL = "https://varian.dynamicfhir.com/"
var caretrackerURL = "https://hag-fhir.amazingcharts.com/ac/endpoints"
var zhhealthcareURL = "https://blueehr.com/fhir-urls/"
var emedpracticeURL = "https://emedpractice.com/fhir/fhirhelpdocument.html"
var doc_torURL = "https://hag-fhir.amazingcharts.com/pc/endpoints"
var azaleahealthURL = "https://api.azaleahealth.com/fhir/R4/Endpoint"
var cloudcraftURL = "https://fhirapitest.naiacorp.net/fhir/r4/endpoints/"
var darenasolutionsURL = "https://api.meldrx.com/Directories/fhir/endpoints"
var glenwoodsystemsURL = "https://static.glaceemr.com/endpoints/urls.json"
var practicefusionURL = "https://www.practicefusion.com/assets/static_files/ServiceBaseURLs.json"
var universalEHRURL = "https://appstudio.interopengine.com/partner/fhirR4endpoints-universalehr.json"
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
var praxisemrURL = "https://www.praxisemr.com/applicationaccess/api/help/"
var escribeHOSTURL = "https://ehr.escribe.com/ehr/api/fhir/swagger-ui/"
var mdlogicEHRURL = "https://www.mdlogic.com/solutions/standard-api-documentation"
var altheaURL = "https://altheafhir.mdsynergy.com"
var webchartnowURL = "https://docs.webchartnow.com/resources/system-specifications/fhir-application-programming-interface-api/endpoints/"
var medifusionURL = "https://docs.medifusion.com/"
var smartemrURL = "https://smartemr.readme.io/reference/getting-started#base-url"
var tebraURL = "https://fhir.prd.cloud.tebra.com/fhir-request/swagger-ui/"
var landmarkhealthURL = "https://lmdmzprodws.landmarkhealth.org/docs/fhir-base-urls.csv"
var nthtechnologyURL = "https://admin.nthtechnology.com/fhir_endpoints.php/json"
var netsmartURL = "https://careconnect.netsmartcloud.com/csv/service-base-urls-20240905.csv"

var omnimdURL = "https://fhirregistration.omnimd.com/#/specification"
var pcesystemsURL = "https://www.pcesystems.com/g10APIInfo.html"
var medicsdaextURL = "https://staging.medicscloud.com/MedicsDAExtAPI/FHIRMedicsDocAssistant.htm"
var azaleahealthr4URL = "https://app.azaleahealth.com/fhir/R4/Endpoint"
var dssincURL = "https://dssjuno-dev-web.dssinc.com/dss/01ho/r4/Home/ApiDocumentation#Api_Urls"
var kodjinURL = "https://docs.kodjin.com/service-base-urls"
var firelyURL = "https://docs.fire.ly/projects/Firely-Server/en/latest/_static/g10/EndpointBundleFirely.json"
var azurewebsitesURL = "https://sfp-proxy9794.azurewebsites.net/fhir/base-url"
var viewmymedURL = "https://portal.viewmymed.com/fhir/Endpoint"
var imedemrURL = "https://icom.imedemr.com/icom50/html/emr/mvc/pages/fhir_endpoints.php?format=csv"
var imedemrURL2 = "https://icom.imedemr.com/icom50/html/emr/mvc/pages/fhir_endpoints.php"
var moyaeURL = "https://documenter.getpostman.com/view/15917486/UyxojQMd#a24aa40c-fe15-478e-a555-3c2cb10d56c9"
var myheloURL = "https://www.myhelo.com/api/"
var nextechURL = "https://www.nextech.com/hubfs/Nextech%20FHIR%20Base%20URL.csv"
var novomediciURL = "https://www.novomedici.com/api-documents/"
var patientpatternURL = "https://patientpattern-static.s3.us-west-2.amazonaws.com/static/documents/fhir-base-urls.csv"
var pcisgoldURL = "https://fhir.pcisgold.com/fhirdocs/practices.json"

var healthieURL = "https://app-52512.on-aptible.com/service-base-urls"
var medConnectURL = "https://api.medconnecthealth.com/fhir/r4/endpoints"
var citiusTechURL = "https://8759937.fs1.hubspotusercontent-na1.net/hubfs/8759937/assets/pdfs/Perform+ConnectServerEndpoints.json"
var enableHealthcareURL = "https://ehifire.ehiconnect.com/fhir/r4/endpoints"
var drchronoURL = "https://drchrono-fhirpresentation.everhealthsoftware.com/fhir/r4/endpoints"
var visionWebURL = "https://dhpresentation.youruprise.com/fhir/r4/endpoints"
var streamlineURL = "https://dhfhirpresentation.smartcarenet.com/fhir/r4/endpoints"
var procentiveURL = "https://fhir-dev.procentive.com/fhir/r4/endpoints"
var tenElevenURL = "https://fhir-dev.10e11.com/fhir/r4/endpoints"
var henryScheinURL = "https://micromddev.dynamicfhir.com/fhir/r4/endpoints"
var iSALUSURL = "https://isalus-fhirpresentation.everhealthsoftware.com/fhir/r4/endpoints"
var healthInnovationURL = "https://revolutionehrdev.dynamicfhir.com/fhir/r4/endpoints"
var mPNSoftwareURL = "https://mpnproxyfhirstore.blob.core.windows.net/serviceurl/ServiceBaseURLs.csv"
var NexusURL = "https://www.nexusclinical.net/nexusehr-fhirapi-base-urls.csv"
var MEDENTURL = "https://www.medent.com/std_api/ServiceBaseURL.csv"
var CarepathsURL = "https://carepaths.com/uploads/org_endpoint_bundle.json"
var athenaClinicalsURL = "https://docs.athenahealth.com/api/guides/base-fhir-urls"
var canvasMedicalURL = "https://docs.canvasmedical.com/api/service-base-urls/"
var veradigmURL = "https://open.platform.veradigm.com/fhirendpoints"
var broadStreetURL = "https://broadstreetcare.com/docs"
var officePracticumURL = "https://fhir-documentation.patientmedrecords.com/endpoints"
var welligentURL = "https://fhir.qa.welligent.com/"
var willowURL = "https://www.willowgladetechnologies.com/requirements"

// var aidboxURL = "https://aidbox.cx360.net/service-base-urls"
var medicaURL = "https://code.medicasoft.us/fhir_r4_endpoints.html"
var dss2URL = "https://dssjess-dev-web.dssinc.com/fhir/r4/endpoints"
var cozevaURL = "https://fhir.cozeva.com/endpoints"
var fhirjunoURL = "https://fhirjuno-prod-web.dssinc.com/fhir/r4/endpoints"
var hcsincURL = "https://hcswebportal.corporate.hcsinc.net/HCSClinicals_FHIR/api/Endpoint?connection-type=hl7-fhir-rest"
var greenwayURL = "https://fhir-servicebaseurl.fhirhlprod.greenwayhealth.com/servicebundle.json"
var criterionsURL = "https://criterions.com/fhir-end-points/"
var maximusURL = "https://documents.maximus.care"
var tenzingURL = "https://tenzing.docs.apiary.io/#introduction/fhir-endpoints"
var inpracsysURL = "https://inpracsys.com/fhir/"

var meldrxURL = "https://app.meldrx.com/api/Directories/fhir/endpoints"
var emr4MDURL = "https://appstudio.interopengine.com/partner/fhirR4endpoints-mednetmedical.json"
var smartCareURL = "https://dhfhirpresentation.smartcarenet.com/"
var dssEmergencyURL = "https://dssjess-dev-web.dssinc.com"
var e11URL = "https://fhir.10e11.com/"
var practicegatewayURL = "https://fhir.practicegateway.net/smart/Endpoint?_format=application/json"
var procentiveFhirURL = "https://fhir.procentive.com/"
var fhirDssjunoURL = "https://fhirjuno-prod-web.dssinc.com"
var officeallyURL = "https://fhirpt.officeally.com/"
var epicURL = "https://open.epic.com/Endpoints/R4"
var qualifactsURL = "https://qualifacts.com/api-page/_downloads/carelogic-fhir-org-list.json"
var myeyecarerecordsURL = "https://smartonfhir.myeyecarerecords.com/fhir/Endpoint?_format=application/fhir+json&status=active"
var nextgenAPIURL = "https://www.nextgen.com/patient-access-api"
var sabiamedURL = "https://www.sabiamed.com/api-endpoints"
var zoommdURL = "https://www.zoommd.com/zoommd-file-api-endpoints"

func QueryCHPLEndpointList(chplURL string, fileToWriteTo string, errorFilePath string) {

	var err error

	if URLsEqual(chplURL, MedHostURL) {
		MedHostQuerier(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, NextGenURL) {
		NextGenwebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, CanvasURL) {
		Canvaswebscraper(chplURL, fileToWriteTo)
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
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, carefluenceURL) {
		CarefluenceWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, bizmaticsURL) {
		err = BundleQuerierParser("https://prognocis.com/fhir/FHIR_FILES/fhirtest.json", fileToWriteTo, errorFilePath, chplURL)
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
	} else if URLsEqual(chplURL, healthSamuraiURL) {
		CustomBundleQuerierParser("https://smartbox.aidbox.app/service-base-urls", fileToWriteTo)
	} else if URLsEqual(chplURL, triarqURL) {
		TRIARQPracticeWebscraper(chplURL, fileToWriteTo)
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
	} else if chplURL == varianmedicalURL {
		VarianMedicalWebscraper(chplURL+"dhit/basepractice/r4/Home/ApiDocumentation", fileToWriteTo)
	} else if chplURL == caretrackerURL {
		err = BundleQuerierParser("https://hag-fhir.amazingcharts.com/ac/endpoints/r4", fileToWriteTo, errorFilePath, chplURL)
	} else if chplURL == zhhealthcareURL {
		ZHHealthcareWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == medinfoengineeringURL {
		MedicalInformaticsEngineeringWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == emedpracticeURL {
		eMedPracticeWebscraper(chplURL, fileToWriteTo)
	} else if chplURL == doc_torURL {
		err = BundleQuerierParser(chplURL+"/r4", fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, azaleahealthURL) {
		err = BundleQuerierParser(chplURL+"?_format=application/json", fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, cloudcraftURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, darenasolutionsURL) {
		err = BundleQuerierParser(darenasolutionsURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, glenwoodsystemsURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, practicefusionURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, universalEHRURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, welligentURL) {
		err = BundleQuerierParser("https://fhir.qa.welligent.com/fhir/r4/endpoints", fileToWriteTo, errorFilePath, chplURL)
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
	} else if URLsEqual(chplURL, athenaClinicalsURL) {
		CSVParser("https://fhir.athena.io/athena-fhir-urls/athenanet-fhir-base-urls.csv", fileToWriteTo, "./athenanet-fhir-base-urls.csv", 17136, 2, true, 3, 1)
	} else if URLsEqual(chplURL, criterionsURL) {
		CriterionsWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, tenzingURL) {
		TenzingURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, inpracsysURL) {
		InpracsysURLWebscraper(chplURL, fileToWriteTo)
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
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, netsmartURL) {
		NetsmartCSVParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, omnimdURL) {
		OmniMDWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, pcesystemsURL) {
		PCESystemsWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, medicsdaextURL) {
		MedicsDAExtAPIWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, azaleahealthr4URL) {
		err = BundleQuerierParser("https://app.azaleahealth.com/fhir/R4/Endpoint?_format=application/json", fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, dssincURL) {
		DssIncWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, kodjinURL) {
		KodjinWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, firelyURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, azurewebsitesURL) {
		AzureWebsitesURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, viewmymedURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, imedemrURL) {
		ImedemrWebscraper("https://icom.imedemr.com/icom50/html/emr/mvc/pages/fhir_endpoints.php", fileToWriteTo)
	} else if URLsEqual(chplURL, imedemrURL2) {
		ImedemrWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, moyaeURL) {
		MoyaeURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, myheloURL) {
		MyheloURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, nextechURL) {
		NextechURLCSVParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, novomediciURL) {
		NovomediciURLWebscraper("https://www.novomedici.com/wp-content/uploads/2022/11/fhir-base-urls.csv", fileToWriteTo)
	} else if URLsEqual(chplURL, patientpatternURL) {
		PatientpatternURLCSVParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, pcisgoldURL) {
		PCISgoldURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, healthieURL) {
		CustomBundleQuerierParser(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, medConnectURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, citiusTechURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, enableHealthcareURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, drchronoURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, visionWebURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, streamlineURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, procentiveURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, tenElevenURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, henryScheinURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, iSALUSURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, healthInnovationURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, CarepathsURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, greenwayURL) {
		err = BundleQuerierParser(chplURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, mPNSoftwareURL) {
		CSVParser("https://mpnproxyfhirstore.blob.core.windows.net/serviceurl/ServiceBaseURLs.csv", fileToWriteTo, "./ServiceBaseURLs.csv", 1, 0, true, 3, 2)
	} else if URLsEqual(chplURL, NexusURL) {
		CSVParser("https://www.nexusclinical.net/nexusehr-fhirapi-base-urls.csv", fileToWriteTo, "./nexusehr-fhirapi-base-urls.csv", 1, 0, true, 2, 1)
	} else if URLsEqual(chplURL, MEDENTURL) {
		CSVParser(MEDENTURL, fileToWriteTo, "./ServiceBaseURL.csv", -1, 2, true, 1, 0)
	} else if URLsEqual(chplURL, canvasMedicalURL) {
		CanvasMedicalURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, maximusURL) {
		MaximusURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, broadStreetURL) {
		BroadStreetURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, officePracticumURL) {
		OfficePracticumURLWebscraper(chplURL, fileToWriteTo)
	} else if URLsEqual(chplURL, willowURL) {
		WillowQuerierParser("https://ccdoc.phn.care/service-base-urls", fileToWriteTo)
		// } else if URLsEqual(chplURL, aidboxURL) {
		// 	AidboxQuerierParser(aidboxURL, fileToWriteTo)
	} else if URLsEqual(chplURL, dss2URL) {
		err = BundleQuerierParser(dss2URL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, cozevaURL) {
		err = BundleQuerierParser("https://fhir.cozeva.com/r4Endpoints.json", fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, medicaURL) {
		err = BundleQuerierParser("https://code.medicasoft.us/fhir_r4_endpoints.json", fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, hcsincURL) {
		err = BundleQuerierParser(hcsincURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, fhirjunoURL) {
		err = BundleQuerierParser(fhirjunoURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, veradigmURL) {
		err = BundleQuerierParser("https://open.platform.veradigm.com/fhirendpoints/download/R4?endpointFilter=All", fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, meldrxURL) {
		err = BundleQuerierParser(meldrxURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, emr4MDURL) {
		err = BundleQuerierParser(emr4MDURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, smartCareURL) {
		err = BundleQuerierParser(smartCareURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, dssEmergencyURL) {
		err = BundleQuerierParser(dssEmergencyURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, e11URL) {
		err = BundleQuerierParser(e11URL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, practicegatewayURL) {
		err = BundleQuerierParser(practicegatewayURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, procentiveFhirURL) {
		err = BundleQuerierParser(procentiveFhirURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, fhirDssjunoURL) {
		err = BundleQuerierParser(fhirDssjunoURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, officeallyURL) {
		err = BundleQuerierParser(officeallyURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, epicURL) {
		err = BundleQuerierParser(epicURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, myeyecarerecordsURL) {
		err = BundleQuerierParser(myeyecarerecordsURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, sabiamedURL) {
		err = BundleQuerierParser(sabiamedURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, myeyecarerecordsURL) {
		err = BundleQuerierParser(myeyecarerecordsURL, fileToWriteTo, errorFilePath, chplURL)
	} else if URLsEqual(chplURL, zoommdURL) {
		ZoomMDCSVParser("https://www.zoommd.com/FHIRServerURLs_ZoomMD.csv", fileToWriteTo)
	} else if URLsEqual(chplURL, qualifactsURL) {
		QualifactsWebscraper(qualifactsURL, fileToWriteTo)
	} else if URLsEqual(chplURL, nextgenAPIURL) {
		NextgenAPIWebscraper(nextgenAPIURL, fileToWriteTo)
	} else if URLsEqual(chplURL, aetnaURL) {
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
