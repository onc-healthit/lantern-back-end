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

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func main() {
	var endpointsFile string
	var source string
	var listURL string
	var format string

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

	if len(os.Args) == 4 {
		endpointsFile = os.Args[1]
		format = os.Args[2]
		source = os.Args[3]
	} else if len(os.Args) == 5 {
		endpointsFile = os.Args[1]
		format = os.Args[2]
		source = os.Args[3]
		listURL = os.Args[4]
	} else if len(os.Args) == 2 {
		log.Fatalf("ERROR: Missing endpoints list format command-line argument")
	} else if len(os.Args) == 3 {
		log.Fatalf("ERROR: Missing endpoints list source command-line argument")
	} else {
		log.Fatalf("ERROR: Endpoints list command-line arguments are not correct")
	}

	listOfEndpoints, err := fetcher.GetEndpointsFromFilepath(endpointsFile, format, source, listURL)
	helpers.FailOnError("Endpoint List Parsing Error: ", err)

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
		dbErr := endptQuerier.RemoveOldEndpoints(ctx, store, time.Now().Add(time.Hour*24), listSource)
		helpers.FailOnError("Deleting old endpoints in fhir_endpoints database error: ", dbErr)
	}
}
