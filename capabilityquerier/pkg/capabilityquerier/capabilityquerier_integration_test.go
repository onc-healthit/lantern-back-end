// +build integration

package capabilityquerier_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/capabilityquerier"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	aq "github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fetcher"
	"github.com/spf13/viper"
)

var mq *lanternmq.MessageQueue
var chID *lanternmq.ChannelID
var endpoints fetcher.ListOfEndpoints

var conn *amqp.Connection
var channel *amqp.Channel

func TestMain(m *testing.M) {
	var err error

	err = config.SetupConfigForTests()
	if err != nil {
		panic(err)
	}

	hap := th.HostAndPort{Host: viper.GetString("qhost"), Port: viper.GetString("qport")}
	err = th.CheckResources(hap)
	if err != nil {
		panic(err)
	}

	err = setup()
	if err != nil {
		panic(err)
	}

	code := m.Run()

	teardown()
	os.Exit(code)
}

func Test_Integration_GetAndSendCapabilityStatement(t *testing.T) {
	queueName := viper.GetString("qname")
	queueIsEmpty(t, queueName)
	defer checkCleanQueue(t, queueName, channel)

	var err error

	// TODO 1/7/2020: rewriting test to use real data. next step: have iterate over subset of
	// endpoints and make call using real client, context, etc.
	client := &http.Client{
		Timeout: time.Second * 35,
	}

	ctx := context.Background()

	for i, endpointEntry := range endpoints.Entries {
		if i >= 10 {
			break
		}
		var urlString = endpointEntry.FHIRPatientFacingURI
		// Specifically query the FHIR endpoint metadata
		metadataURL, err := url.Parse(urlString)
		if err != nil {
			log.Warn("endpoint URL parsing error: ", err.Error())
		} else {
			fmt.Printf("Getting and sending capability statement %d/10\n", i+1)
			metadataURL.Path = path.Join(metadataURL.Path, "metadata")
			args := make(map[string]interface{})
			querierArgs := capabilityquerier.QuerierArgs{
				FhirURL:      metadataURL.String(),
				Client:       client,
				MessageQueue: mq,
				ChannelID:    chID,
				QueueName:    queueName,
			}
			args["querierArgs"] = querierArgs
			err = capabilityquerier.GetAndSendCapabilityStatement(ctx, &args)
			th.Assert(t, err == nil, err)
		}
	}
	count, err := aq.QueueCount(queueName, channel)
	th.Assert(t, err == nil, err)
	// need to pause to ensure all messages are on the queue before we count them
	time.Sleep(10 * time.Second)
	th.Assert(t, (count == 9 || count == 10), fmt.Sprintf("expected there to be 9 or 10 messages in the queue (difference because dealing with real data and endpoints); saw %d", count))
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

func setup() error {
	var err error

	qUser := viper.GetString("quser")
	qPassword := viper.GetString("qpassword")
	qHost := viper.GetString("qhost")
	qPort := viper.GetString("qport")
	qName := viper.GetString("qname")

	// set up wrapped queue info
	mq_, chID_, err := aq.ConnectToServerAndQueue(qUser, qPassword, qHost, qPort, qName)
	mq = &mq_
	chID = &chID_
	if err != nil {
		return err
	}

	// setup specific queue info so we can test what's in the queue
	s := fmt.Sprintf("amqp://%s:%s@%s:%s/", qUser, qPassword, qHost, qPort)
	conn, err = amqp.Dial(s)
	if err != nil {
		return err
	}

	channel, err = conn.Channel()
	if err != nil {
		return err
	}

	// grab endpoints
	// TODO: eventually this method of getting endpoints will change
	endpoints, err = fetcher.GetEndpointsFromFilepath("../../../endpointmanager/resources/EndpointSources.json", "")

	return err
}

func teardown() {
	(*mq).Close()
	channel.Close()
	conn.Close()
}
