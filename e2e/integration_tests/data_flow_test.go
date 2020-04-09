// +build e2e

package integration_tests

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointlinker"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	endptQuerier "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fhirendpointquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/nppesquerier"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
	"github.com/spf13/viper"
)

type Endpoint struct {
	url               string
	organization_name string
	mapped_npi_ids    []string
}

func TestMain(m *testing.M) {
	config.SetupConfigForTests()
	populateTestNPIData()
	populateTestEndpointData()
	go setupTestServer()
	// Give time for the querier to query the test server we just setup
	time.Sleep(30 * time.Second)
	os.Exit(m.Run())
}

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func populateTestNPIData() {
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError(err)
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
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError(err)

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
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError(err)

	defer store.Close()
	response_time_row := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints;")
	var link_count int
	err = response_time_row.Scan(&link_count)
	failOnError(err)

	if link_count != 35 {
		t.Fatalf("Only 35 endpoint should have been parsed out of TestEndpointSources.json, Got: " + strconv.Itoa(link_count))
	}
}

func Test_EndpointLinksAreAvailable(t *testing.T) {
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError(err)

	defer store.Close()
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
		query_str := "SELECT id FROM fhir_endpoints WHERE organization_name='" + ep.organization_name + "';"
		err = store.DB.QueryRow(query_str).Scan(&endpoint_id)
		if err != nil {
			t.Fatalf("failed org name is " + ep.organization_name)
		}
		failOnError(err)

		// Assert that endpoint id has correct url
		var endpoint_url string
		query_str = "SELECT url FROM fhir_endpoints WHERE id=" + endpoint_id + ";"
		err = store.DB.QueryRow(query_str).Scan(&endpoint_url)
		failOnError(err)
		if endpoint_url != ep.url {
			t.Fatalf("Endpoint id mapped to wrong endpoint url")
		}
		// Assert that the correct endpoint has correct number of npi organizations mapped
		var num_npi_ids int
		query_str = "SELECT count(*) FROM endpoint_organization WHERE endpoint_id =" + endpoint_id + ";"
		err = store.DB.QueryRow(query_str).Scan(&num_npi_ids)
		failOnError(err)
		if num_npi_ids != len(ep.mapped_npi_ids) {
			t.Fatalf("Expected number of npi organizations mapped to endpoint is " + strconv.Itoa(len(ep.mapped_npi_ids)) + " Got: " + strconv.Itoa(num_npi_ids))
		}

		for _, npi_id := range ep.mapped_npi_ids {
			// Get organization id for each npi id
			var org_id string
			query_str = "SELECT id FROM npi_organizations WHERE npi_id='" + npi_id + "';"
			err = store.DB.QueryRow(query_str).Scan(&org_id)
			failOnError(err)
			// Assert that each npi organization is mapped to correct endpoint
			var linked_endpoint_id string
			query_str = "SELECT endpoint_id FROM endpoint_organization WHERE organization_id =" + org_id + ";"
			err = store.DB.QueryRow(query_str).Scan(&linked_endpoint_id)
			failOnError(err)
			if linked_endpoint_id != endpoint_id {
				t.Fatalf("Endpoint id mapped to wrong npi organization")
			}
		}

		// Assert that deletion from npi_organizations list removes the link
		// Assert that deletion from fhir_endpoints list removes the link
		if len(ep.mapped_npi_ids) == 1 {
			query_str = "DELETE FROM npi_organizations WHERE npi_id='" + ep.mapped_npi_ids[0] + "';"
			_, err = store.DB.Exec(query_str)
			err = store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;").Scan(&link_count)
			failOnError(err)
			if link_count != expected_link_count-1 {
				t.Fatalf("Database should only contain " + strconv.Itoa(expected_link_count-1) + " links after npi_organization was deleted. Has: " + strconv.Itoa(link_count))
			}
			expected_link_count = link_count
		} else {
			query_str = "DELETE FROM fhir_endpoints WHERE id=" + endpoint_id + ";"
			_, err = store.DB.Exec(query_str)
			err = store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;").Scan(&link_count)
			failOnError(err)
			if link_count != expected_link_count-len(ep.mapped_npi_ids) {
				t.Fatalf("Database should contain " + strconv.Itoa(expected_link_count) + " links. Has: " + strconv.Itoa(link_count))
			}
			expected_link_count = link_count
		}
	}
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
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError(err)

	defer store.Close()
	response_time_row := store.DB.QueryRow("SELECT * FROM metrics_labels WHERE metric_name = 'AllEndpoints_http_response_time';")
	var id, metric_name, result_label string
	err = response_time_row.Scan(&id, &metric_name, &result_label)
	failOnError(err)

	if result_label != "{\"job\": \"FHIRQUERY\", \"orgName\": \"LanternTestOrg\", \"instance\": \"endpoint_querier:3333\"}" {
		t.Fatalf("LanternTestOrg not found in AllEndpoints_http_response_time metric")
	}
	// TODO add additional queries for other metrics
}
