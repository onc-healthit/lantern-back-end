// +build e2e

package integration_tests

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
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
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	aq "github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
	"github.com/spf13/viper"
	Assert "github.com/stretchr/testify/assert"
)

type Endpoint struct {
	url               string
	organization_name string
	mapped_npi_ids    []string
}

var store *postgresql.Store

func TestMain(m *testing.M) {
	config.SetupConfigForTests()

	var err error
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	if err != nil {
		panic(err)
	}

	teardown, err := th.IntegrationDBTestSetupMain(store.DB)

	populateTestNPIData()
	populateTestEndpointData()
	go setupTestServer()
	// Give time for the querier to query the test server we just setup
	time.Sleep(30 * time.Second)

	code := m.Run()

	teardown(store.DB)

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

func Test_EndpointDataIsAvailable(t *testing.T) {
	var err error
	response_time_row := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints;")
	var link_count int
	err = response_time_row.Scan(&link_count)
	failOnError(err)

	if link_count != 35 {
		t.Fatalf("Only 35 endpoint should have been parsed out of TestEndpointSources.json, Got: " + strconv.Itoa(link_count))
	}
}

func Test_EndpointLinksAreAvailable(t *testing.T) {
	var err error
	expected_link_count := 40
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
		t.Fatalf("Database should only have made 40 links given the fake NPPES data that was loaded. Has: " + strconv.Itoa(link_count))
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
		query_str := "SELECT id FROM fhir_endpoints WHERE organization_name=$1;"
		err = store.DB.QueryRow(query_str, ep.organization_name).Scan(&endpoint_id)
		if err != nil {
			t.Fatalf("failed org name is " + ep.organization_name)
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
		query_str = "SELECT count(*) FROM endpoint_organization WHERE endpoint_id =$1;"
		err = store.DB.QueryRow(query_str, endpoint_id).Scan(&num_npi_ids)
		failOnError(err)
		if num_npi_ids != len(ep.mapped_npi_ids) {
			t.Fatalf("Expected number of npi organizations mapped to endpoint is " + strconv.Itoa(len(ep.mapped_npi_ids)) + " Got: " + strconv.Itoa(num_npi_ids))
		}

		for _, npi_id := range ep.mapped_npi_ids {
			// Get organization id for each npi id
			var org_id string
			query_str = "SELECT id FROM npi_organizations WHERE npi_id=$1;"
			err = store.DB.QueryRow(query_str, npi_id).Scan(&org_id)
			failOnError(err)
			// Assert that each npi organization is mapped to correct endpoint
			var linked_endpoint_id string
			query_str = "SELECT endpoint_id FROM endpoint_organization WHERE organization_id =$1;"
			err = store.DB.QueryRow(query_str, org_id).Scan(&linked_endpoint_id)
			failOnError(err)
			if linked_endpoint_id != endpoint_id {
				t.Fatalf("Endpoint id mapped to wrong npi organization")
			}
		}

		// Assert that deletion from npi_organizations list removes the link
		// Assert that deletion from fhir_endpoints list removes the link
		if len(ep.mapped_npi_ids) == 1 {
			query_str = "DELETE FROM npi_organizations WHERE npi_id=$1;"
			_, err = store.DB.Exec(query_str, ep.mapped_npi_ids[0])
			err = store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;").Scan(&link_count)
			failOnError(err)
			if link_count != expected_link_count-1 {
				t.Fatalf("Database should only contain " + strconv.Itoa(expected_link_count-1) + " links after npi_organization was deleted. Has: " + strconv.Itoa(link_count))
			}
			expected_link_count = link_count
		} else {
			query_str = "DELETE FROM fhir_endpoints WHERE id=$1;"
			_, err = store.DB.Exec(query_str, endpoint_id)
			err = store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;").Scan(&link_count)
			failOnError(err)
			if link_count != expected_link_count-len(ep.mapped_npi_ids) {
				t.Fatalf("Database should contain " + strconv.Itoa(expected_link_count) + " links. Has: " + strconv.Itoa(link_count))
			}
			expected_link_count = link_count
		}
	}
}

func Test_GetCHPLProducts(t *testing.T) {
	var err error
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

	var testHITP endpointmanager.HealthITProduct = endpointmanager.HealthITProduct{
		Name:                  "Intuitive Medical Document",
		Version:               "2.0",
		Developer:             "Intuitive Medical Documents",
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
	th.Assert(t, hitp.Developer == testHITP.Developer, "Developer is not what was expected")
	th.Assert(t, hitp.CertificationDate.Equal(testHITP.CertificationDate), "Certification date is not what was expected")
	th.Assert(t, hitp.CertificationStatus == testHITP.CertificationStatus, "Certification status is not what was expected")
	th.Assert(t, reflect.DeepEqual(hitp.CertificationCriteria, testHITP.CertificationCriteria), "Certification criteria is not what was expected")
}

func Test_RetrieveCapabilityStatements(t *testing.T) {
	var err error
	qUser := viper.GetString("quser")
	qPassword := viper.GetString("qpassword")
	qHost := viper.GetString("qhost")
	qPort := viper.GetString("qport")
	qName := viper.GetString("qname")

	hap := th.HostAndPort{Host: qHost, Port: qPort}
	err = th.CheckResources(hap)
	if err != nil {
		panic(err)
	}

	var mq lanternmq.MessageQueue
	var chID lanternmq.ChannelID
	mq, chID, err = aq.ConnectToServerAndQueue(qUser, qPassword, qHost, qPort, qName)
	defer mq.Close()
	th.Assert(t, err == nil, err)
	th.Assert(t, mq != nil, "expected message queue to be created")
	th.Assert(t, chID != nil, "expected channel ID to be created")

	ctx, cancel := context.WithCancel(context.Background())
	go capabilityhandler.ReceiveCapabilityStatements(ctx, store, mq, chID, qName)
	time.Sleep(30 * time.Second)
	cancel()

	query_str := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info where capability_statement is not null;")
	var capability_statement_count int
	err = query_str.Scan(&capability_statement_count)
	failOnError(err)
	if capability_statement_count == 0 {
		t.Fatalf("Fhir_endpoints_info db should have capability statements")
	}

	query_str = store.DB.QueryRow("SELECT COUNT(capability_statement->>'fhirVersion') FROM fhir_endpoints_info;")
	var fhir_version_count int
	err = query_str.Scan(&fhir_version_count)
	failOnError(err)
	if fhir_version_count < 300 {
		t.Fatalf("There should be at least 300 capability statement with fhir version specified")
	}

	common_vendor_list := [2]string{"Epic Systems Corporation", "Cerner Corporation"}
	rows, err := store.DB.Query("SELECT DISTINCT vendor FROM fhir_endpoints_info where vendor!='';")
	failOnError(err)
	var test_vendor_list []string
	defer rows.Close()
	for rows.Next() {
		var vendor string
		err = rows.Scan(&vendor)
		failOnError(err)
		test_vendor_list = append(test_vendor_list, vendor)
	}
	th.Assert(t, len(test_vendor_list)>= len(common_vendor_list), "List of distinct vendors should at least include most common vendors")
	Assert.Contains(t, test_vendor_list, common_vendor_list, "List of distinct vendors should include Epic and Cerner")
}

func Test_MetricsAvailableInQuerier(t *testing.T) {
	var client http.Client
	resp, err := client.Get("http://endpoint_querier:3333/metrics")
	failOnError(err)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Error retrieving metrics from endpoint querier")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	failOnError(err)

	bodyString := string(bodyBytes)

	if !strings.Contains(bodyString, "AllEndpoints_http_request_responses{orgName=\"LanternTestOrg\"} 200") {
		t.Fatalf("Endpoint querier missing or incorrect response code metric for LanternTestOrg")
	}

	if !strings.Contains(bodyString, "AllEndpoints_http_response_time{orgName=\"LanternTestOrg\"}") {
		t.Fatalf("Endpoint querier missing response time metric for LanternTestOrg")
	}

	if !strings.Contains(bodyString, "AllEndpoints_total_uptime_checks{orgName=\"LanternTestOrg\"}") {
		t.Fatalf("Endpoint querier missing uptime checks metric for LanternTestOrg")
	}
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
	response_time_row := store.DB.QueryRow("SELECT * FROM metrics_labels WHERE metric_name = 'AllEndpoints_http_response_time';")
	var id, metric_name, result_label string
	err = response_time_row.Scan(&id, &metric_name, &result_label)
	failOnError(err)

	if result_label != "{\"job\": \"FHIRQUERY\", \"orgName\": \"LanternTestOrg\", \"instance\": \"endpoint_querier:3333\"}" {
		t.Fatalf("LanternTestOrg not found in AllEndpoints_http_response_time metric")
	}
	// TODO add additional queries for other metrics
}
