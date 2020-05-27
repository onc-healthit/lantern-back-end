// +build e2e

package integration_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityhandler"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointlinker"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	endptQuerier "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fhirendpointquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/nppesquerier"
	se "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/sendendpoints"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	aq "github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
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
	populateTestEndpointData()
	go setupTestServer()
	// Give time for the querier to query the test server we just setup
	time.Sleep(30 * time.Second)

	code := m.Run()

	teardown(store.DB)
	channel.Close()
	conn.Close()

	os.Exit(code)
}

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func assert(t *testing.T, boolStatement bool, errorValue interface{}) {
	if !boolStatement {
		t.Fatalf("%s: %+v", t.Name(), errorValue)
	}
}

func populateTestNPIData() {
	var err error
	fname := "./testdata/npidata_min.csv"
	ctx := context.Background()
	err = store.DeleteAllNPIOrganizations(ctx)
	_, err = nppesquerier.ParseAndStoreNPIFile(ctx, fname, store)
	failOnError(err)
}

func populateTestEndpointData() {
	content, err := ioutil.ReadFile("./testdata/TestEndpointSources.json")
	failOnError(err)
	listOfEndpoints, err := fetcher.GetListOfEndpoints(content, "Test")
	failOnError(err)

	ctx := context.Background()

	dbErr := endptQuerier.AddEndpointData(ctx, store, &listOfEndpoints)
	failOnError(dbErr)
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
	failOnError(err)

	if link_count != 34 {
		t.Fatalf("Only 34 endpoint should have been parsed out of TestEndpointSources.json, Got: " + strconv.Itoa(link_count))
	}
}

func Test_EndpointLinksAreAvailable(t *testing.T) {
	var err error
	expected_link_count := 38
	endpoint_orgs_row := store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;")
	var link_count int
	err = endpoint_orgs_row.Scan(&link_count)
	failOnError(err)

	if link_count != 0 {
		t.Fatalf("Empty database should not have had any links made yet. Has: " + strconv.Itoa(link_count))
	}

	ctx := context.Background()
	endpointlinker.LinkAllOrgsAndEndpoints(ctx, store, false)

	endpoint_orgs_row = store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;")
	err = endpoint_orgs_row.Scan(&link_count)
	failOnError(err)

	if link_count != expected_link_count {
		t.Fatalf("Database should only have made 38 links given the fake NPPES data that was loaded. Has: " + strconv.Itoa(link_count))
	}

	// endpoint maps to one org
	ep1 := Endpoint{
		url:               "https://epicproxy.et1094.epichosted.com/FHIRProxy/api/FHIR/DSTU2/",
		organization_name: "Cape Fear Valley Health",
		mapped_npi_ids:    []string{"1588667794"},
	}

	// endpoint maps to multiple orgs
	ep2 := Endpoint{
		url:               "https://FHIR.valleymed.org/FHIR-PRD/api/FHIR/DSTU2/",
		organization_name: "Valley Medical Center",
		mapped_npi_ids:    []string{"1629071758", "1164427431", "1245230598", "1790787307", "1366444978", "1356343735"},
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
		failOnError(err)

		// Assert that endpoint id has correct url
		var endpoint_url string
		query_str = "SELECT url FROM fhir_endpoints WHERE id=$1;"
		err = store.DB.QueryRow(query_str, endpoint_id).Scan(&endpoint_url)
		failOnError(err)
		if endpoint_url != ep.url {
			t.Fatalf("Endpoint id mapped to wrong endpoint url")
		}
		// Assert that the correct endpoint has correct number of npi organizations mapped
		var num_npi_ids int
		query_str = "SELECT count(*) FROM endpoint_organization WHERE url =$1;"
		err = store.DB.QueryRow(query_str, ep.url).Scan(&num_npi_ids)
		failOnError(err)
		if num_npi_ids != len(ep.mapped_npi_ids) {
			t.Fatalf("Expected number of npi organizations mapped to endpoint is " + strconv.Itoa(len(ep.mapped_npi_ids)) + " Got: " + strconv.Itoa(num_npi_ids))
		}

		for _, npi_id := range ep.mapped_npi_ids {
			// Assert that each npi organization is mapped to correct endpoint
			var linked_endpoint_url string
			query_str = "SELECT url FROM endpoint_organization WHERE organization_npi_id =$1;"
			err = store.DB.QueryRow(query_str, npi_id).Scan(&linked_endpoint_url)
			failOnError(err)
			if linked_endpoint_url != ep.url {
				t.Fatalf("Endpoint url mapped to wrong npi organization")
			}
		}
	}
}

func Test_GetCHPLVendors(t *testing.T) {
	var err error
	var actualVendsStored int

	if viper.GetString("chplapikey") == "" {
		t.Skip("Skipping Test_GetCHPLProducts because the CHPL API key is not set.")
	}

	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 35,
	}

	// as of 5/11/20, at least 1440 entries are expected to be added to the database
	minNumExpVendsStored := 1440

	err = chplquerier.GetCHPLVendors(ctx, store, client)
	assert(t, err == nil, err)
	rows := store.DB.QueryRow("SELECT COUNT(*) FROM vendors;")
	err = rows.Scan(&actualVendsStored)
	assert(t, err == nil, err)
	assert(t, actualVendsStored >= minNumExpVendsStored, fmt.Sprintf("Expected at least %d vendors stored. Actually had %d vendors stored.", minNumExpVendsStored, actualVendsStored))

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
	assert(t, err == nil, err)
	assert(t, vend.CHPLID == 1658, "CHPLID not as expected")
	assert(t, vend.DeveloperCode == "2657", "DeveloperCode not as expected")
	assert(t, vend.Name == "Carefluence", "Name not as expected")
	assert(t, vend.URL == "http://www.carefluence.com", "URL not as expected")
	assert(t, vend.Location.ZipCode == "48439", "ZipCode not as expected")
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
	failOnError(err)
	if hitp_count != 0 {
		t.Fatalf("Healthit product database should initially be empty")
	}

	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 35,
	}
	err = chplquerier.GetCHPLProducts(ctx, store, client)
	failOnError(err)

	healthit_prod_row = store.DB.QueryRow("SELECT COUNT(*) FROM healthit_products;")
	err = healthit_prod_row.Scan(&hitp_count)
	failOnError(err)
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

	vend, err := store.GetVendorUsingName(ctx, "Intuitive Medical Documents")
	assert(t, err == nil, err)

	var testHITP endpointmanager.HealthITProduct = endpointmanager.HealthITProduct{
		Name:                  "Intuitive Medical Document",
		Version:               "2.0",
		VendorID:              vend.ID,
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2016, 3, 4, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2014",
		CHPLID:                "CHP-029177",
		CertificationCriteria: []string{"100", "101", "103", "106", "107", "108", "109", "114", "115", "116", "61", "62", "63", "64", "65", "66", "67", "68", "69", "70", "71", "72", "73", "74", "75", "76", "77", "81", "82", "83", "84", "86", "87", "88", "91", "92", "93", "94", "95", "96", "97", "98", "99"},
	}

	hitp, err := store.GetHealthITProductUsingNameAndVersion(ctx, "Intuitive Medical Document", "2.0")
	th.Assert(t, err == nil, err)
	th.Assert(t, hitp.CHPLID == testHITP.CHPLID, "CHPL ID is not what was expected")
	th.Assert(t, hitp.CertificationEdition == testHITP.CertificationEdition, "Certification edition is not what was expected")
	th.Assert(t, hitp.VendorID == testHITP.VendorID, "Developer is not what was expected")
	th.Assert(t, hitp.CertificationDate.Equal(testHITP.CertificationDate), "Certification date is not what was expected")
	th.Assert(t, hitp.CertificationStatus == testHITP.CertificationStatus, "Certification status is not what was expected")
	th.Assert(t, reflect.DeepEqual(hitp.CertificationCriteria, testHITP.CertificationCriteria), "Certification criteria is not what was expected")
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
	failOnError(err)
	if capability_statement_count == 0 {
		t.Fatalf("Fhir_endpoints_info db should have capability statements")
	}

	query_str = store.DB.QueryRow("SELECT COUNT(capability_statement->>'fhirVersion') FROM fhir_endpoints_info;")
	var fhir_version_count int
	expected_fhir_version_count := 25
	err = query_str.Scan(&fhir_version_count)
	failOnError(err)
	if fhir_version_count < expected_fhir_version_count {
		t.Fatalf("There should be at least 25 capability statement with fhir version specified, actual is " + strconv.Itoa(fhir_version_count))
	}
}

func Test_VendorList(t *testing.T) {
	var err error

	if viper.GetString("chplapikey") == "" {
		t.Skip("Skipping Test_VendorList because the CHPL API key is not set.")
	}

	ctx := context.Background()

	epic, err := store.GetVendorUsingName(ctx, "Epic Systems Corporation")
	failOnError(err)
	cerner, err := store.GetVendorUsingName(ctx, "Cerner Corporation")
	failOnError(err)

	common_vendor_list := [2]int{epic.ID, cerner.ID}
	rows, err := store.DB.Query("SELECT DISTINCT vendor_id FROM fhir_endpoints_info where vendor_id!=0;")
	failOnError(err)
	var test_vendor_list []int
	defer rows.Close()
	for rows.Next() {
		var vendorID int
		err = rows.Scan(&vendorID)
		failOnError(err)
		test_vendor_list = append(test_vendor_list, vendorID)
	}
	th.Assert(t, len(test_vendor_list) >= len(common_vendor_list), "List of distinct vendors should at least include most common vendors")
	Assert.Contains(t, test_vendor_list, common_vendor_list[0], "List of distinct vendors should include Epic")
	Assert.Contains(t, test_vendor_list, common_vendor_list[1], "List of distinct vendors should include Cerner")
}

func Test_MetricsAvailableInQuerier(t *testing.T) {
	var err error
	queueIsEmpty(t, testQName)
	defer checkCleanQueue(t, testQName, channel)

	// Set-up the test queue
	var mq lanternmq.MessageQueue
	var chID lanternmq.ChannelID
	mq, chID, err = aq.ConnectToServerAndQueue(qUser, qPassword, qHost, qPort, testQName)
	defer mq.Close()
	th.Assert(t, err == nil, err)
	th.Assert(t, mq != nil, "expected message queue to be created")
	th.Assert(t, chID != nil, "expected channel ID to be created")

	ctx := context.Background()
	sendEndpointsOverQueue(ctx, t, testQName, mq, chID)

	var client http.Client
	resp, err := client.Get("http://endpoint_querier:3333/metrics")
	failOnError(err)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Error retrieving metrics from endpoint querier")
	}

	// Random set of URLs in the TestEndpointSources, unlikely that all 5 of them will have failed
	// during a run
	possibleUrls := [5]string{
		"https://interconnect.lcmchealth.org/FHIR/api/FHIR/DSTU2/",
		"https://lmcrcs.lexmed.com/FHIR/api/FHIR/DSTU2/",
		"https://fhir.healow.com/FHIRServer/fhir/IGCGAD/",
		"https://eprescribe.mercy.net/PRDFHIRSTL/rvh/api/FHIR/DSTU2/",
		"https://webproxy.comhs.org/FHIR/api/FHIR/DSTU2/",
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	failOnError(err)

	bodyString := string(bodyBytes)

	requestCheck := false
	httpRespCheck := false
	uptimeCheck := false

	for _, url := range possibleUrls {
		reqFormat := fmt.Sprintf("AllEndpoints_http_request_responses{url=\"%s\"} 200", url)
		if strings.Contains(bodyString, reqFormat) {
			requestCheck = true
		}

		respFormat := fmt.Sprintf("AllEndpoints_http_response_time{url=\"%s\"}", url)
		if strings.Contains(bodyString, respFormat) {
			httpRespCheck = true
		}

		uptimeFormat := fmt.Sprintf("AllEndpoints_total_uptime_checks{url=\"%s\"}", url)
		if strings.Contains(bodyString, uptimeFormat) {
			uptimeCheck = true
		}
	}

	th.Assert(t, requestCheck == true, "Endpoint querier missing or incorrect response code metric for all tested URLs")
	th.Assert(t, httpRespCheck == true, "Endpoint querier missing response time metric for all tested URLs")
	th.Assert(t, uptimeCheck == true, "Endpoint querier missing uptime checks metric for all tested URLs")
}
func Test_QuerierAvailableToPrometheus(t *testing.T) {
	type PrometheusTargets struct {
		Status string `json:"status"`
		Data   struct {
			ActiveTargets []struct {
				DiscoveredLabels struct {
					Address     string `json:"__address__"`
					MetricsPath string `json:"__metrics_path__"`
					Scheme      string `json:"__scheme__"`
					Job         string `json:"job"`
				} `json:"discoveredLabels"`
				Labels struct {
					Instance string `json:"instance"`
					Job      string `json:"job"`
				} `json:"labels"`
				ScrapeURL  string    `json:"scrapeUrl"`
				LastError  string    `json:"lastError"`
				LastScrape time.Time `json:"lastScrape"`
				Health     string    `json:"health"`
			} `json:"activeTargets"`
			DroppedTargets []interface{} `json:"droppedTargets"`
		} `json:"data"`
	}
	var client http.Client

	resp, err := client.Get("http://prometheus:9090/api/v1/targets")
	failOnError(err)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Error retrieving active targets from prometheus")
	}

	var targets = new(PrometheusTargets)
	err = json.NewDecoder(resp.Body).Decode(targets)
	failOnError(err)

	if targets.Data.ActiveTargets[0].Health != "up" {
		t.Fatalf("Prometheus is not reporting the endpoint_querier as being up")
	}
}

func Test_MetricsWrittenToPostgresDB(t *testing.T) {
	var err error
	ctx := context.Background()
	response_time_rows, err := store.DB.QueryContext(ctx, "SELECT * FROM metrics_labels WHERE metric_name = 'AllEndpoints_http_response_time';")
	failOnError(err)

	// Random set of URLs in the TestEndpointSources, unlikely that all 5 of them will have failed
	// during a run
	possibleUrls := [5]string{
		"https://interconnect.lcmchealth.org/FHIR/api/FHIR/DSTU2/",
		"https://lmcrcs.lexmed.com/FHIR/api/FHIR/DSTU2/",
		"https://fhir.healow.com/FHIRServer/fhir/IGCGAD/",
		"https://eprescribe.mercy.net/PRDFHIRSTL/rvh/api/FHIR/DSTU2/",
		"https://webproxy.comhs.org/FHIR/api/FHIR/DSTU2/",
	}

	isInDB := false
	defer response_time_rows.Close()
	for response_time_rows.Next() {
		var id, metric_name, result_label string
		err = response_time_rows.Scan(&id, &metric_name, &result_label)
		for _, url := range possibleUrls {
			if strings.Contains(result_label, url) {
				expectedResultLabel := fmt.Sprintf("{\"job\": \"FHIRQUERY\", \"url\": \"%s\", \"instance\": \"endpoint_querier:3333\"}", url)
				if result_label == expectedResultLabel {
					isInDB = true
					break
				}
			}
		}
		if isInDB {
			break
		}
	}
	if !isInDB {
		t.Fatalf("None of the tested URLs were found in AllEndpoints_http_response_time metric")
	}
	// TODO add additional queries for other metrics
}
