// +build e2e

package integration_tests

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/capabilityhandler"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointlinker"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fetcher"
	endptQuerier "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fhirendpointquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/nppesquerier"
	se "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/sendendpoints"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	aq "github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	Assert "github.com/stretchr/testify/assert"
)

type Endpoint struct {
	url               string
	organization_name string
	mapped_npi_ids    []string
}

var store *postgresql.Store
var qUser, qPassword, qHost, qPort, testQName string
var endptList = "./testdata/TestEndpointSources.json"
var shortEndptList = "./testdata/TestEndpointSources_1.json"
var LanternEndptList = "./testdata/TestLanternEndpointSources.json"

var conn *amqp.Connection
var channel *amqp.Channel

func TestMain(m *testing.M) {
	config.SetupConfigForTests()
	var err error
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	if err != nil {
		panic(err)
	}

	teardown, err := th.IntegrationDBTestSetupMain(store.DB)
	testQueueSetup()

	populateTestNPIData()
	populateTestEndpointData(endptList, "Test")
	go setupTestServer()
	// Give time for the querier to query the test server we just setup
	time.Sleep(30 * time.Second)

	code := m.Run()

	teardown(store.DB)
	channel.Close()
	conn.Close()

	os.Exit(code)
}

func populateTestNPIData() {
	var err error
	fname := "./testdata/npidata_min.csv"
	ctx := context.Background()
	err = store.DeleteAllNPIOrganizations(ctx)
	_, err = nppesquerier.ParseAndStoreNPIFile(ctx, fname, store)
	helpers.FailOnError("", err)
}

func populateTestEndpointData(testEndpointList string, source string) {
	var listOfEndpoints fetcher.ListOfEndpoints
	var knownSource fetcher.Source
	content, err := ioutil.ReadFile(testEndpointList)
	helpers.FailOnError("", err)

	if source == "Test" {
		listOfEndpoints, err = fetcher.GetListOfEndpoints(content, source)
		helpers.FailOnError("", err)
	} else {
		knownSource = "LanternEndpointSourcesJson"
		listOfEndpoints, err = fetcher.GetListOfEndpointsKnownSource(content, knownSource)
		helpers.FailOnError("", err)
	}

	ctx := context.Background()

	dbErr := endptQuerier.AddEndpointData(ctx, store, &listOfEndpoints)
	helpers.FailOnError("", dbErr)
}

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	contents, err := ioutil.ReadFile("testdata/DSTU2CapabilityStatement.xml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/fhir+xml")
	_, err = w.Write(contents)
	// Don't fail on error in this case since test connection will drop out when test ends
	if err != nil {
		log.Printf("%s", err)
	}
}

func setupTestServer() {
	http.HandleFunc("/metadata", metadataHandler)
	var err = http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal("HTTP Server Creation Error: ", err.Error())
	}
}

func testQueueSetup() {
	var err error

	qUser = viper.GetString("quser")
	qPassword = viper.GetString("qpassword")
	qHost = viper.GetString("qhost")
	qPort = viper.GetString("qport")
	testQName = viper.GetString("qname")

	hap := th.HostAndPort{Host: qHost, Port: qPort}
	err = th.CheckResources(hap)
	if err != nil {
		log.Fatal("Check Resources Error: ", err.Error())
	}

	fmt.Printf("amqp://%s:%s@%s:%s/", qUser, qPassword, qHost, qPort)
	// setup specific queue info so we can test what's in the queue
	s := fmt.Sprintf("amqp://%s:%s@%s:%s/", qUser, qPassword, qHost, qPort)
	conn, err = amqp.Dial(s)
	if err != nil {
		log.Fatal("Database Connection Error: ", err.Error())
	}

	channel, err = conn.Channel()
	if err != nil {
		log.Fatal("Channel Connection Error: ", err.Error())
	}
}

func sendEndpointsOverQueue(ctx context.Context, t *testing.T, queueName string, mq lanternmq.MessageQueue, chID lanternmq.ChannelID) {
	var wg sync.WaitGroup
	wg.Add(1)
	errs := make(chan error)

	go se.GetEnptsAndSend(ctx, &wg, queueName, 10, store, &mq, &chID, errs)
	time.Sleep(30 * time.Second)
}

func queueIsEmpty(t *testing.T, queueName string) {
	count, err := aq.QueueCount(queueName, channel)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 0, "should be no messages in queue.")
}

func checkCleanQueue(t *testing.T, queueName string, channel *amqp.Channel) {
	err := aq.CleanQueue(queueName, channel)
	th.Assert(t, err == nil, err)
}

func Test_EndpointDataIsAvailable(t *testing.T) {
	var err error
	response_time_row := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints;")
	var link_count int
	err = response_time_row.Scan(&link_count)
	helpers.FailOnError("", err)

	if link_count != 30 {
		t.Fatalf("Only 30 endpoint should have been parsed out of TestEndpointSources.json, Got: " + strconv.Itoa(link_count))
	}
}

func Test_EndpointLinksAreAvailable(t *testing.T) {
	var err error
	expected_link_count := 31
	endpoint_orgs_row := store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;")
	var link_count int
	err = endpoint_orgs_row.Scan(&link_count)
	helpers.FailOnError("", err)

	if link_count != 0 {
		t.Fatalf("Empty database should not have had any links made yet. Has: " + strconv.Itoa(link_count))
	}

	ctx := context.Background()
	//This will add one link to the endpoint_organization table
	endpointlinker.LinkAllOrgsAndEndpoints(ctx, store, "./testdata/fakeWhitelist.json", "./testdata/fakeBlacklist.json", false)

	endpoint_orgs_row = store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;")
	err = endpoint_orgs_row.Scan(&link_count)
	helpers.FailOnError("", err)

	if link_count != expected_link_count {
		t.Fatalf("Database should only have made 30 links given the fake NPPES data that was loaded. Has: " + strconv.Itoa(link_count))
	}

	// endpoint maps to multiple orgs
	ep1 := Endpoint{
		url:               "https://epicproxy.et1094.epichosted.com/FHIRProxy/api/FHIR/DSTU2/",
		organization_name: "Cape Fear Valley Health",
		mapped_npi_ids:    []string{"1111111111", "1497758544", "1639172869", "1790784999", "1588667794"},
	}

	// endpoint maps to one org
	ep2 := Endpoint{
		url:               "https://FHIR.valleymed.org/FHIR-PRD/api/FHIR/DSTU2/",
		organization_name: "Valley Medical Center",
		mapped_npi_ids:    []string{"1245230598"},
	}

	// endpoint maps to no orgs
	ep3 := Endpoint{
		url:               "https://mcproxyprd.med.umich.edu/FHIR-PRD/api/FHIR/DSTU2/",
		organization_name: "Michigan Medicine",
		mapped_npi_ids:    []string{},
	}

	ep_list := []Endpoint{ep1, ep2, ep3}

	for _, ep := range ep_list {

		// Get endpoint id
		var endpoint_id string
		query_str := "SELECT id FROM fhir_endpoints WHERE url=$1;"
		err = store.DB.QueryRow(query_str, ep.url).Scan(&endpoint_id)
		if err != nil {
			t.Fatalf("failed org url is "+ep.url+"\nError %v\n", err)
		}
		helpers.FailOnError("", err)

		// Assert that endpoint id has correct url
		var endpoint_url string
		query_str = "SELECT url FROM fhir_endpoints WHERE id=$1;"
		err = store.DB.QueryRow(query_str, endpoint_id).Scan(&endpoint_url)
		helpers.FailOnError("", err)
		if endpoint_url != ep.url {
			t.Fatalf("Endpoint id mapped to wrong endpoint url")
		}
		// Assert that the correct endpoint has correct number of npi organizations mapped
		var num_npi_ids int
		query_str = "SELECT count(*) FROM endpoint_organization WHERE url =$1;"
		err = store.DB.QueryRow(query_str, ep.url).Scan(&num_npi_ids)
		helpers.FailOnError("", err)
		if num_npi_ids != len(ep.mapped_npi_ids) {
			t.Fatalf("Expected number of npi organizations mapped to endpoint is " + strconv.Itoa(len(ep.mapped_npi_ids)) + " Got: " + strconv.Itoa(num_npi_ids))
		}

		for _, npi_id := range ep.mapped_npi_ids {
			// Assert that each npi organization is mapped to correct endpoint
			var linked_endpoint_url string
			query_str = "SELECT url FROM endpoint_organization WHERE organization_npi_id =$1;"
			err = store.DB.QueryRow(query_str, npi_id).Scan(&linked_endpoint_url)
			helpers.FailOnError("", err)
			if linked_endpoint_url != ep.url {
				t.Fatalf("Endpoint url mapped to wrong npi organization")
			}
		}
	}
}

func Test_GetCHPLCriteria(t *testing.T) {
	var err error
	var actualCriteriaStored int

	if viper.GetString("chplapikey") == "" {
		t.Skip("Skipping Test_GetCHPLCriteria because the CHPL API key is not set.")
	}

	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 35,
	}

	// as of 7/30/20, at least 182 entries are expected to be added to the database
	minNumExpCriteriaStored := 182

	err = chplquerier.GetCHPLCriteria(ctx, store, client, "")
	th.Assert(t, err == nil, err)
	rows := store.DB.QueryRow("SELECT COUNT(*) FROM certification_criteria;")
	err = rows.Scan(&actualCriteriaStored)
	th.Assert(t, err == nil, err)
	th.Assert(t, actualCriteriaStored >= minNumExpCriteriaStored, fmt.Sprintf("Expected at least %d criteria stored. Actually had %d criteria stored.", minNumExpCriteriaStored, actualCriteriaStored))

	// expect to see this entry in the database:
	// {
	// CertificationID: 44,
	// CertificationNumber: "170.315 (f)(2)",
	// Title: "Transmission to Public Health Agencies - Syndromic Surveillance",
	// CertificationEditionID: 3,
	// CertificationEdition: "2015",
	// Description: null,
	// Removed: false
	// }
	criteria, err := store.GetCriteriaByCertificationID(ctx, 44)
	th.Assert(t, err == nil, err)
	th.Assert(t, criteria.CertificationID == 44, "CertificationID not as expected")
	th.Assert(t, criteria.CertificationNumber == "170.315 (f)(2)", "CertificationNumber not as expected")
	th.Assert(t, criteria.Title == "Transmission to Public Health Agencies - Syndromic Surveillance", "Title not as expected")
	th.Assert(t, criteria.CertificationEditionID == 3, "CertificationEditionID not as expected")
	th.Assert(t, criteria.CertificationEdition == "2015", "CertificationEdition not as expected")
	th.Assert(t, criteria.Removed == false, "Removed not as expected")
}

func Test_GetCHPLVendors(t *testing.T) {
	var err error
	var actualVendsStored int

	if viper.GetString("chplapikey") == "" {
		t.Skip("Skipping Test_GetCHPLVendors because the CHPL API key is not set.")
	}

	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 35,
	}

	// as of 5/11/20, at least 1440 entries are expected to be added to the database
	minNumExpVendsStored := 1440

	err = chplquerier.GetCHPLVendors(ctx, store, client, "")
	th.Assert(t, err == nil, err)
	rows := store.DB.QueryRow("SELECT COUNT(*) FROM vendors;")
	err = rows.Scan(&actualVendsStored)
	th.Assert(t, err == nil, err)
	th.Assert(t, actualVendsStored >= minNumExpVendsStored, fmt.Sprintf("Expected at least %d vendors stored. Actually had %d vendors stored.", minNumExpVendsStored, actualVendsStored))

	// expect to see this entry in the database:
	// {
	// "developerId": 1658,
	// "developerCode": "2657",
	// "name": "Carefluence",
	// "website": "http://www.carefluence.com",
	// "selfDeveloper": false,
	// "address": {
	// 	"addressId": 101,
	// 	"line1": "8359 Office Park Drive",
	// 	"line2": null,
	// 	"city": "Grand Blanc",
	// 	"state": "MI",
	// 	"zipcode": "48439",
	// 	"country": "US"
	// }
	vend, err := store.GetVendorUsingName(ctx, "Carefluence")
	th.Assert(t, err == nil, err)
	th.Assert(t, vend.CHPLID == 1658, "CHPLID not as expected")
	th.Assert(t, vend.DeveloperCode == "2657", "DeveloperCode not as expected")
	th.Assert(t, vend.Name == "Carefluence", "Name not as expected")
	th.Assert(t, vend.URL == "http://www.carefluence.com", "URL not as expected")
	th.Assert(t, vend.Location.ZipCode == "48439", "ZipCode not as expected")
}

func Test_GetCHPLProducts(t *testing.T) {
	var err error

	if viper.GetString("chplapikey") == "" {
		t.Skip("Skipping Test_GetCHPLProducts because the CHPL API key is not set.")
	}

	healthit_prod_row := store.DB.QueryRow("SELECT COUNT(*) FROM healthit_products;")
	expected_hitp_count := 7829
	var hitp_count int
	err = healthit_prod_row.Scan(&hitp_count)
	helpers.FailOnError("", err)
	if hitp_count != 0 {
		t.Fatalf("Healthit product database should initially be empty")
	}

	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 35,
	}
	err = chplquerier.GetCHPLProducts(ctx, store, client, "")
	helpers.FailOnError("", err)

	healthit_prod_row = store.DB.QueryRow("SELECT COUNT(*) FROM healthit_products;")
	err = healthit_prod_row.Scan(&hitp_count)
	helpers.FailOnError("", err)
	if hitp_count < expected_hitp_count {
		t.Fatalf("Database should have at least " + strconv.Itoa(expected_hitp_count) + " health it products after querying chpl Got: " + strconv.Itoa(hitp_count))
	}

	// expect this in db
	//{
	//	"id":3,
	//	"chplProductNumber":"CHP-029177",
	//	"edition":"2014",
	//	"practiceType":"Inpatient",
	//	"developer":"Intuitive Medical Documents",
	//	"product":"Intuitive Medical Document",
	//	"version":"2.0",
	//	"certificationDate":1457049600000,
	//	"certificationStatus":"Active",
	//	"criteriaMet":"100☺101☺103☺106☺107☺108☺109☺114☺115☺116☺61☺62☺63☺64☺65☺66☺67☺68☺69☺70☺71☺72☺73☺74☺75☺76☺77☺81☺82☺83☺84☺86☺87☺88☺91☺92☺93☺94☺95☺96☺97☺98☺99"
	//	}

	ctx = context.Background()
	vend, err := store.GetVendorUsingName(ctx, "Intuitive Medical Documents")
	th.Assert(t, err == nil, err)

	var testHITP endpointmanager.HealthITProduct = endpointmanager.HealthITProduct{
		Name:                  "Intuitive Medical Document",
		Version:               "2.0",
		VendorID:              vend.ID,
		CertificationStatus:   "Retired",
		CertificationDate:     time.Date(2016, 3, 4, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2014",
		CHPLID:                "CHP-029177",
		CertificationCriteria: []int{100, 101, 103, 106, 107, 108, 109, 114, 115, 116, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 81, 82, 83, 84, 86, 87, 88, 91, 92, 93, 94, 95, 96, 97, 98, 99},
	}

	hitp, err := store.GetHealthITProductUsingNameAndVersion(ctx, "Intuitive Medical Document", "2.0")
	th.Assert(t, err == nil, err)
	th.Assert(t, hitp.CHPLID == testHITP.CHPLID, "CHPL ID is not what was expected")
	th.Assert(t, hitp.CertificationEdition == testHITP.CertificationEdition, "Certification edition is not what was expected")
	th.Assert(t, hitp.VendorID == testHITP.VendorID, "Developer is not what was expected")
	th.Assert(t, hitp.CertificationDate.Equal(testHITP.CertificationDate), "Certification date is not what was expected")
	th.Assert(t, hitp.CertificationStatus == testHITP.CertificationStatus, "Certification status is not what was expected")
	th.Assert(t, reflect.DeepEqual(hitp.CertificationCriteria, testHITP.CertificationCriteria), "Certification criteria is not what was expected")

	// check that there are links in the product_criteria database
	var link_count int
	prod_crit_row := store.DB.QueryRow("SELECT COUNT(*) FROM product_criteria;")
	err = prod_crit_row.Scan(&link_count)
	helpers.FailOnError("", err)

	if link_count <= 0 {
		t.Fatalf("There should be links in the product_criteria table.")
	}
}

func Test_RetrieveCapabilityStatements(t *testing.T) {
	var err error
	queueIsEmpty(t, testQName)
	defer checkCleanQueue(t, testQName, channel)
	capQName := viper.GetString("endptinfo_capquery_qname")

	var mq lanternmq.MessageQueue
	var chID lanternmq.ChannelID
	mq, chID, err = aq.ConnectToServerAndQueue(qUser, qPassword, qHost, qPort, capQName)
	defer mq.Close()
	th.Assert(t, err == nil, err)
	th.Assert(t, mq != nil, "expected message queue to be created")
	th.Assert(t, chID != nil, "expected channel ID to be created")

	ctx := context.Background()
	sendEndpointsOverQueue(ctx, t, capQName, mq, chID)

	mq, chID, err = aq.ConnectToQueue(mq, chID, testQName)
	defer mq.Close()
	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)
	go capabilityhandler.ReceiveCapabilityStatements(ctx, store, mq, chID, testQName)
	select {
	case <-ctx.Done():
		return
	}
	query_str := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info where capability_statement is not null;")
	var capability_statement_count int
	err = query_str.Scan(&capability_statement_count)
	helpers.FailOnError("", err)
	if capability_statement_count == 0 {
		t.Fatalf("Fhir_endpoints_info db should have capability statements")
	}

	query_str = store.DB.QueryRow("SELECT COUNT(capability_statement->>'fhirVersion') FROM fhir_endpoints_info;")
	var fhir_version_count int
	expected_fhir_version_count := 30
	err = query_str.Scan(&fhir_version_count)
	helpers.FailOnError("", err)
	if fhir_version_count < expected_fhir_version_count {
		t.Fatalf("There should be at least 30 capability statement with fhir version specified, actual is " + strconv.Itoa(fhir_version_count))
	}

	epic, err := store.GetVendorUsingName(ctx, "Epic Systems Corporation")
	helpers.FailOnError("", err)
	cerner, err := store.GetVendorUsingName(ctx, "Cerner Corporation")
	helpers.FailOnError("", err)

	common_vendor_list := [2]int{epic.ID, cerner.ID}
	vendor_rows, err := store.DB.Query("SELECT DISTINCT vendor_id FROM fhir_endpoints_info where vendor_id!=0;")
	helpers.FailOnError("", err)
	var test_vendor_list []int
	defer vendor_rows.Close()
	for vendor_rows.Next() {
		var vendorID int
		err = vendor_rows.Scan(&vendorID)
		helpers.FailOnError("", err)
		test_vendor_list = append(test_vendor_list, vendorID)
	}
	th.Assert(t, len(test_vendor_list) >= len(common_vendor_list), "List of distinct vendors should at least include most common vendors")
	Assert.Contains(t, test_vendor_list, common_vendor_list[0], "List of distinct vendors should include Epic")
	Assert.Contains(t, test_vendor_list, common_vendor_list[1], "List of distinct vendors should include Cerner")

	// Test that availability is populated
	availability_ct_st := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_availability;")
	expected_availability_ct := 30
	var availability_count int
	err = availability_ct_st.Scan(&availability_count)
	helpers.FailOnError("", err)
	if availability_count != expected_availability_ct {
		t.Fatalf("There should be same number of endpoints in availability table as fhir_endpoints_info, Got: %d", availability_count)
	}

	// Test that old endpoints are removed if not in list on update
	populateTestEndpointData(shortEndptList, "Test")

	expected_endpt_ct := 26
	endpt_ct_st := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints;")
	var endpt_count int
	err = endpt_ct_st.Scan(&endpt_count)
	helpers.FailOnError("", err)
	if endpt_count != expected_endpt_ct {
		t.Fatalf("Only %d endpoints should be in fhir_endpoints after updating with file %s, Got: %d", expected_endpt_ct, shortEndptList, endpt_count)
	}

	//This will add one link to the endpoint_organization table
	endpointlinker.LinkAllOrgsAndEndpoints(ctx, store, "./testdata/fakeWhitelist.json", "./testdata/fakeBlacklist.json", false)

	// Check that links were not deleted on update in order to maintain previous mappings from endpoints
	// to organizations
	expected_link_count := 31
	var link_count int
	endpoint_orgs_row := store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;")
	err = endpoint_orgs_row.Scan(&link_count)
	helpers.FailOnError("", err)
	if link_count != expected_link_count {
		t.Fatalf("endpoint_organization should still have %d links after update", expected_link_count)
	}

	// Check that endpoints were not deleted from availability table
	err = availability_ct_st.Scan(&availability_count)
	helpers.FailOnError("", err)
	if availability_count != expected_availability_ct {
		t.Fatalf("fhir_endpoints_availability should still have %d endpoints after update, Got: %d", expected_availability_ct, availability_count)
	}

	endpt_info_ct_st := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info;")
	var endpt_info_count int
	err = endpt_info_ct_st.Scan(&endpt_info_count)
	helpers.FailOnError("", err)
	if endpt_info_count != expected_endpt_ct {
		t.Fatalf("fhir_endpoints_info should have %d endpoints after update. Got: %d", expected_endpt_ct, link_count)
	}

	// List of endpoint urls that were removed in TestEndpointSources_1.json
	fhir_urls := []string{"https://eloh-mapilive.primehealthcare.com/v1/argonaut/v1/",
		"https://epicproxy.et1094.epichosted.com/FHIRProxy/api/FHIR/DSTU2/",
		"https://webproxy.comhs.org/FHIR/api/FHIR/DSTU2/",
		"https://rwebproxy.elcaminohospital.org/FHIR/api/FHIR/DSTU2/",
		"https://lmcrcs.lexmed.com/FHIR/api/FHIR/DSTU2/",
		"https://proxy.cfmedicalcenter.com/FHIRProxyPRD/api/FHIR/DSTU2/"}

	expected_deleted_endpt := 6
	rows, err := store.DB.Query("SELECT url FROM fhir_endpoints_info_history WHERE operation='D';")
	var deleted_fhir_urls []string
	helpers.FailOnError("", err)
	defer rows.Close()
	for rows.Next() {
		var fhirURL string
		err = rows.Scan(&fhirURL)
		helpers.FailOnError("", err)
		deleted_fhir_urls = append(deleted_fhir_urls, fhirURL)
	}

	if len(deleted_fhir_urls) != expected_deleted_endpt {
		t.Fatalf("%d endpoints should have been deleted from fhir_endpoints_info update. Got: %d", expected_deleted_endpt, len(deleted_fhir_urls))
	}
	th.Assert(t, helpers.StringArraysEqual(deleted_fhir_urls, fhir_urls), fmt.Sprintf("expected %v to equal %v", deleted_fhir_urls, fhir_urls))

	for _, url := range fhir_urls {
		var endpoint_id string
		query_str := "SELECT id FROM fhir_endpoints WHERE url=$1;"
		err = store.DB.QueryRow(query_str, url).Scan(&endpoint_id)
		th.Assert(t, err == sql.ErrNoRows, fmt.Sprintf("expected %s to be deleted", url))

		// check that endpoint availability is correct
		var http_response int
		var availability float64
		get_availability_str := "SELECT http_response, availability FROM fhir_endpoints_info WHERE url=$1;"
		err = store.DB.QueryRow(get_availability_str, url).Scan(&http_response, &availability)
		helpers.FailOnError("", err)
		if http_response == 200 {
			th.Assert(t, availability == 1.0, fmt.Sprintf("expected availability for %s to be %f", url, availability))
		} else {
			th.Assert(t, availability == 0, fmt.Sprintf("expected availability for %s to be %f", url, availability))
		}
	}
}

func Test_LanternSource(t *testing.T) {
	// reset values
	_, err := store.DB.Exec("DELETE FROM fhir_endpoints;")
	th.Assert(t, err == nil, err)

	// reset values
	_, err = store.DB.Exec("DELETE FROM endpoint_organization;")
	th.Assert(t, err == nil, err)

	// reset values
	_, err = store.DB.Exec("DELETE FROM fhir_endpoints_info;")
	th.Assert(t, err == nil, err)

	populateTestEndpointData(LanternEndptList, "Lantern")

	var endpt_count int
	expected_endpt_ct := 2
	endpt_ct_st := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints;")
	err = endpt_ct_st.Scan(&endpt_count)
	helpers.FailOnError("", err)
	if endpt_count != expected_endpt_ct {
		t.Fatalf("Only %d endpoints should be in fhir_endpoints after updating with file %s, Got: %d", expected_endpt_ct, LanternEndptList, endpt_count)
	}

	ctx := context.Background()
	endpointlinker.LinkAllOrgsAndEndpoints(ctx, store, "./testdata/fakeWhitelist.json", "./testdata/fakeBlacklist.json", false)

	expected_link_count := 4
	var link_count int

	endpoint_orgs_row := store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;")
	err = endpoint_orgs_row.Scan(&link_count)
	helpers.FailOnError("", err)
	if link_count != expected_link_count {
		t.Fatalf("endpoint_organization should have %d links, had %d", expected_link_count, link_count)
	}

	expected_link_count = 2

	endpoint_orgs_row = store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization WHERE url = 'example.com/';")
	err = endpoint_orgs_row.Scan(&link_count)
	helpers.FailOnError("", err)
	if link_count != expected_link_count {
		t.Fatalf("example.com should have %d links in endpoint_organization, had %d", expected_link_count, link_count)
	}
}
