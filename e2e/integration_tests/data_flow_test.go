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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointlinker"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	endptQuerier "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fhirendpointquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/nppesquerier"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
	"github.com/spf13/viper"
)

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

	if link_count != 1 {
		t.Fatalf("Only one endpoint should have been parsed out of TestEndpointSources.json, Got: " + strconv.Itoa(link_count))
	}
}

func Test_EndpointLinksAreAvailable(t *testing.T) {
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError(err)

	defer store.Close()
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

	if link_count != 2 {
		t.Fatalf("Database should only have made one link given the fake NPPES data that was loaded. Has: " + strconv.Itoa(link_count))
	}

	var lantern_org_endpoint_id string
	err = store.DB.QueryRow("SELECT id FROM fhir_endpoints WHERE organization_name='Lantern Test Org';").Scan(&lantern_org_endpoint_id)
	failOnError(err)

	// Based on the fixture file npidata_min.csv there should be 2 matches to "Lantern Test Org" NPI ID =: 111111111 and NPI ID: 2222222222
	// The database ids for these npi orgs should exist in the link table mapped to lantern test org
	var id_npi_1 string
	err = store.DB.QueryRow("SELECT id FROM npi_organizations WHERE npi_id='1111111111';").Scan(&id_npi_1)
	failOnError(err)
	var id_npi_2 string
	err = store.DB.QueryRow("SELECT id FROM npi_organizations WHERE npi_id='2222222222';").Scan(&id_npi_2)
	failOnError(err)

	var linked_endpoint_id string
	query_str := "SELECT endpoint_id FROM endpoint_organization WHERE  organization_id=" + id_npi_1 + ";"
	err = store.DB.QueryRow(query_str).Scan(&linked_endpoint_id)
	failOnError(err)
	if linked_endpoint_id != lantern_org_endpoint_id {
		t.Fatalf("Org mapped to wrong dndpoint id")
	}
	query_str = "SELECT endpoint_id FROM endpoint_organization WHERE  organization_id=" + id_npi_2 + ";"
	err = store.DB.QueryRow(query_str).Scan(&linked_endpoint_id)
	failOnError(err)
	if linked_endpoint_id != lantern_org_endpoint_id {
		t.Fatalf("Org mapped to wrong dndpoint id")
	}

	// Assert that deletion from npi_organizations list removes the link
	query_str = "DELETE FROM npi_organizations WHERE id=" + id_npi_1 + ";"
	_, err = store.DB.Exec(query_str)
	err = store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;").Scan(&link_count)
	failOnError(err)
	if link_count != 1 {
		t.Fatalf("Database should only contain 1 link after npi_organization was deleted. Has: " + strconv.Itoa(link_count))
	}

	// Assert that deletion from fhir_endpoint list removes the link
	query_str = "DELETE FROM fhir_endpoints WHERE id=" + lantern_org_endpoint_id + ";"
	_, err = store.DB.Exec(query_str)

	err = store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization;").Scan(&link_count)
	failOnError(err)
	if link_count != 0 {
		t.Fatalf("Database should not contain any links. Has: " + strconv.Itoa(link_count))
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

func Test_GetCHPLProducts(t *testing.T) {
	var err error
	var actualProdsStored int

	if viper.GetString("chplapikey") == "" {
		t.Skip("Skipping Test_GetCHPLProducts because the CHPL API key is not set.")
	}

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))

	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 35,
	}

	// as of 12/5/19, at least 7676 entries are expected to be added to the database
	minNumExpProdsStored := 7676

	err = chplquerier.GetCHPLProducts(ctx, store, client)
	assert(t, err == nil, err)
	rows := store.DB.QueryRow("SELECT COUNT(*) FROM healthit_products;")
	err = rows.Scan(&actualProdsStored)
	assert(t, err == nil, err)
	assert(t, actualProdsStored >= minNumExpProdsStored, fmt.Sprintf("Expected at least %d products stored. Actually had %d products stored.", minNumExpProdsStored, actualProdsStored))

	// expect to see this entry in the DB:
	// {
	// 	"id": 7849,
	// 	"chplProductNumber": "15.04.04.2657.Care.01.00.0.160701",
	// 	"edition": "2015",
	// 	"developer": "Carefluence",
	// 	"product": "Carefluence Open API",
	// 	"version": "1",
	// 	"certificationDate": 1467331200000,
	// 	"certificationStatus": "Active",
	// 	"criteriaMet": "170.315 (d)(1)☺170.315 (d)(10)☺170.315 (d)(9)☺170.315 (g)(4)☺170.315 (g)(5)☺170.315 (g)(6)☺170.315 (g)(7)☺170.315 (g)(8)☺170.315 (g)(9)",
	// 	"apiDocumentation": "170.315 (g)(7)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
	// }
	hitp, err := store.GetHealthITProductUsingNameAndVersion(ctx, "Carefluence Open API", "1")
	assert(t, err == nil, err)
	assert(t, hitp.CHPLID == "15.04.04.2657.Care.01.00.0.160701", "CHPL ID is not what was expected")
	assert(t, hitp.CertificationEdition == "2015", "Certification edition is not what was expected")
	assert(t, hitp.Developer == "Carefluence", "Developer is not what was expected")
	assert(t, hitp.CertificationDate.Equal(time.Unix(1467331200, 0).UTC()), "Certification date is not what was expected")
	assert(t, hitp.CertificationStatus == "Active", "Certification status is not what was expected")
	// TODO: Can continue to assert this format after changes described in https://oncprojectracking.healthit.gov/support/browse/LANTERN-156 are addressed
	//assert(t, reflect.DeepEqual(hitp.CertificationCriteria, []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"}), "Certification criteria is not what was expected")
}
