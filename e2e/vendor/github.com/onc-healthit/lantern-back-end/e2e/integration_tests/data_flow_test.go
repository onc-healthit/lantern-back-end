// +build e2e

package integration_tests

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	config.SetupConfigForTests()
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

func TestMetricsAvailableInQuerier(t *testing.T) {
	var client http.Client
	resp, err := client.Get("http://endpoint_querier_1:3333/metrics")
	failOnError(err)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Error retrieving metrics from endpoint querier")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	failOnError(err)

	bodyString := string(bodyBytes)

	if !strings.Contains(bodyString, "AllEndpoints_fhir_version{orgName=\"LanternTestOrg\"} 102") {
		t.Fatalf("Endpoint querier missing or incorrect fhir_version metric for LanternTestOrg")
	}

	if !strings.Contains(bodyString, "AllEndpoints_http_request_responses{orgName=\"LanternTestOrg\"} 200") {
		t.Fatalf("Endpoint querier missing or incorrect response code metric for LanternTestOrg")
	}

	if !strings.Contains(bodyString, "AllEndpoints_http_response_time{orgName=\"LanternTestOrg\"}") {
		t.Fatalf("Endpoint querier missing response time metric for LanternTestOrg")
	}

	if !strings.Contains(bodyString, "AllEndpoints_tls_version{orgName=\"LanternTestOrg\"} 0") {
		t.Fatalf("Endpoint querier missing or incorrect tls version metric for LanternTestOrg")
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
	println(viper.GetString("dbhost"))
	println(viper.GetString("dbuser"))
	println(viper.GetString("dbpassword"))
	println(viper.GetString("dbname"))
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError(err)

	println("SUCCEEDED CONNECTING TO DB")

	defer store.Close()
	fhir_version_row := store.DB.QueryRow("SELECT * FROM metrics_labels WHERE metric_name = 'AllEndpoints_fhir_version';")
	var id, metric_name, result_label string
	err = fhir_version_row.Scan(&id, &metric_name, &result_label)
	failOnError(err)

	if result_label != "{\"job\": \"FHIRQUERY\", \"orgName\": \"LanternTestOrg\", \"instance\": \"endpoint_querier_1:3333\"}" {
		t.Fatalf("LanternTestOrg not found in AllEndpoints_fhir_version metric")
	}
	// TODO add additional queries for other metrics
}
