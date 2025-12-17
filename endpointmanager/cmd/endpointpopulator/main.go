package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fetcher"
	endptQuerier "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fhirendpointquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"

	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func main() {
	var endpointsFile string
	var source string
	var listURL string
	var format string
	var sourceCategory string

	err := config.SetupConfig()
	helpers.FailOnError("Error setting up config", err)

	var channel *amqp.Channel

	capQName := viper.GetString("endptinfo_capquery_qname")
	qUser := viper.GetString("quser")
	qPassword := viper.GetString("qpassword")
	qHost := viper.GetString("qhost")
	qPort := viper.GetString("qport")

	// setup specific queue info so we can test what's in the queue
	s := fmt.Sprintf("amqp://%s:%s@%s:%s/", qUser, qPassword, qHost, qPort)
	conn, err := amqp.Dial(s)
	helpers.FailOnError("", err)

	channel, err = conn.Channel()
	helpers.FailOnError("", err)

	count, err := accessqueue.QueueCount(capQName, channel)
	helpers.FailOnError("", err)

	if count != 0 {
		log.Fatalf("There are %d messages in the queue. Queue must be empty to run the endpoint populator.", count)
	}

	if len(os.Args) == 5 {
		endpointsFile = os.Args[1]
		format = os.Args[2]
		source = os.Args[3]
		sourceCategory = os.Args[4]
	} else if len(os.Args) == 6 {
		endpointsFile = os.Args[1]
		format = os.Args[2]
		source = os.Args[3]
		sourceCategory = os.Args[4]
		listURL = os.Args[5]
	} else if len(os.Args) == 3 {
		log.Fatalf("ERROR: Missing endpoints list format command-line argument")
	} else if len(os.Args) == 4 {
		log.Fatalf("ERROR: Missing endpoints list source command-line argument")
	} else {
		log.Fatalf("ERROR: Endpoints list command-line arguments are not correct")
	}

	listOfEndpoints, err := fetcher.GetEndpointsFromFilepath(endpointsFile, format, source, listURL)
	if err != nil && strings.Contains(err.Error(), "incorrect reference value") {
		log.Error("Endpoint List Parsing Error: ", err)
	}

	err = config.SetupConfig()
	helpers.FailOnError("", err)

	ctx := context.Background()
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	log.Info("Successfully connected to DB!")

	if len(listOfEndpoints.Entries) != 0 {
		dbErr := endptQuerier.AddEndpointData(ctx, store, &listOfEndpoints)
		helpers.FailOnError("Saving in fhir_endpoints database error: ", dbErr)
	} else {
		var listSource string
		if listURL != "" {
			listSource = listURL
		} else {
			listSource = source
		}
		dbErr := endptQuerier.RemoveOldEndpointOrganizations(ctx, store, time.Now().Add(time.Hour*24), listSource)
		helpers.FailOnError("Deleting old endpoint organizations in fhir_endpoint_organizations database error: ", dbErr)

		dbErr = endptQuerier.RemoveOldEndpoints(ctx, store, time.Now().Add(time.Hour*24), listSource)
		helpers.FailOnError("Deleting old endpoints in fhir_endpoints database error: ", dbErr)
	}

	// UPSERT statement - inserts new list_source or updates existing one
	// Multiple developers can share the same list_source URL
	// Uses composite unique constraint (list_source, developer_name)
	addListSourceStatement := `
	INSERT INTO list_source_info (
		list_source,
		is_chpl,
		developer_name
	)
	VALUES ($1, $2, $3)
	ON CONFLICT (list_source, developer_name)
	DO UPDATE SET
		is_chpl = EXCLUDED.is_chpl;
	`

	if sourceCategory == "State Medicaid" {
		// State Medicaid: Each endpoint has ListSource = developer name from CSV
		// Collect all unique ListSource values (developer names) from endpoints
		uniqueListSources := make(map[string]bool)
		for _, endpoint := range listOfEndpoints.Entries {
			if endpoint.ListSource != "" {
				uniqueListSources[endpoint.ListSource] = true
			}
		}

		// Insert each unique ListSource (developer name) into list_source_info
		// For State Medicaid: list_source = developer_name = ListSource (same value)
		for listSourceValue := range uniqueListSources {
			_, sourceErr := store.DB.ExecContext(ctx, addListSourceStatement, listSourceValue, sourceCategory, listSourceValue)
			if sourceErr != nil {
				log.Warnf("Error adding list source '%s' to list_source_info: %v", listSourceValue, sourceErr)
			}
		}
	} else {
		// CHPL/Payer/Other: Insert single list source with developer name
		var listSource string
		if listURL != "" {
			listSource = listURL
		} else {
			listSource = source
		}

		// source = developer name from ${NAME} in populatedb.sh
		_, sourceErr := store.DB.ExecContext(ctx, addListSourceStatement, listSource, sourceCategory, source)
		if sourceErr != nil {
			log.Warnf("Error adding list source '%s' to list_source_info: %v", listSource, sourceErr)
		}
	}
}
