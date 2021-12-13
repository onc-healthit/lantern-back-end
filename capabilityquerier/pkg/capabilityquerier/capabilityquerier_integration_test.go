// +build integration

package capabilityquerier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/mock"
	aq "github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fetcher"
	"github.com/spf13/viper"
)

var store *postgresql.Store
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

	err = setup()
	if err != nil {
		panic(err)
	}

	hapDB := th.HostAndPort{Host: viper.GetString("dbhost"), Port: viper.GetString("dbport")}
	err = th.CheckResources(hapDB)
	if err != nil {
		panic(err)
	}

	hapQ := th.HostAndPort{Host: viper.GetString("qhost"), Port: viper.GetString("qport")}
	err = th.CheckResources(hapQ)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	teardown()
	os.Exit(code)
}

func Test_Integration_GetAndSendCapabilityStatement(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

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
			querierArgs := QuerierArgs{
				FhirURL:      metadataURL.String(),
				Client:       client,
				MessageQueue: mq,
				ChannelID:    chID,
				QueueName:    queueName,
				Store:        store,
			}
			args["querierArgs"] = querierArgs
			err = GetAndSendCapabilityStatement(ctx, &args)
			th.Assert(t, err == nil, err)
		}
	}
	count, err := aq.QueueCount(queueName, channel)
	th.Assert(t, err == nil, err)
	// need to pause to ensure all messages are on the queue before we count them
	time.Sleep(10 * time.Second)
	th.Assert(t, (count == 9 || count == 10), fmt.Sprintf("expected there to be 9 or 10 messages in the queue (difference because dealing with real data and endpoints); saw %d", count))
}

func Test_Integration_GetAndSendCapabilityStatement2(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var ctx context.Context
	var fhirURL *url.URL
	var tc *th.TestClient
	var message []byte
	var ch lanternmq.ChannelID
	ctx = context.Background()
	var err error

	mq := mock.NewBasicMockMessageQueue()
	ch = 1
	queueName := "queue name"

	// basic test

	fhirURL = &url.URL{}
	fhirURL, err = fhirURL.Parse(sampleURL)
	th.Assert(t, err == nil, err)
	ctx = context.Background()
	tc, err = testClientWithContentType(fhir2LessJSONMIMEType)
	th.Assert(t, err == nil, err)
	defer tc.Close()

	// create the expected result
	expectedCapStat, err := capabilityStatement()
	th.Assert(t, err == nil, err)
	expectedMimeType := []string{fhir2LessJSONMIMEType, fhir3PlusJSONMIMEType}
	expectedTLSVersion := "TLS 1.0"
	expectedMsgStruct := Message{
		URL:                  fhirURL.String(),
		MIMETypes:            expectedMimeType,
		TLSVersion:           expectedTLSVersion,
		HTTPResponse:         200,
		SMARTHTTPResponse:    200,
		ResponseTime:         0,
		RequestedFhirVersion: "None",
	}
	err = json.Unmarshal(expectedCapStat, &(expectedMsgStruct.CapabilityStatement))
	th.Assert(t, err == nil, err)
	// GetAndSendCapabilityStatement uses one client to call requestCapabilityStatementAndSmartOnFhir
	// which makes make multiple request. The tes client only returns the metadata info which is why smart_response
	// has the same value as capabilityStatement
	err = json.Unmarshal(expectedCapStat, &(expectedMsgStruct.SMARTResp))
	th.Assert(t, err == nil, err)
	expectedMsg, err := json.Marshal(expectedMsgStruct)
	th.Assert(t, err == nil, err)

	args := make(map[string]interface{})
	querierArgs := QuerierArgs{
		FhirURL:        sampleURL,
		Client:         &(tc.Client),
		RequestVersion: "None",
		MessageQueue:   &mq,
		ChannelID:      &ch,
		QueueName:      queueName,
		Store:          store,
	}
	args["querierArgs"] = querierArgs

	// execute tested function
	err = GetAndSendCapabilityStatement(ctx, &args)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(mq.(*mock.BasicMockMessageQueue).Queue) == 1, "expect one message on the queue")
	message = <-mq.(*mock.BasicMockMessageQueue).Queue

	//Change response time in message to 0 to make the response time match with the expected message
	var messageStruct Message
	err = json.Unmarshal(message, &messageStruct)
	th.Assert(t, err == nil, "expect no error to be thrown when unmarshalling message")
	messageStruct.ResponseTime = 0
	messageStruct.RequestedFhirVersion = "None"
	message, err = json.Marshal(messageStruct)
	th.Assert(t, err == nil, "expect no error to be thrown when marshalling message")

	th.Assert(t, bytes.Equal(message, expectedMsg), "expected the capability statement on the queue to be the same as the one sent")

	// context canceled error
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = GetAndSendCapabilityStatement(ctx, &args)
	th.Assert(t, err == nil, "expected GetAndSendCapabilityStatement not to error out due to context ending")
	th.Assert(t, len(mq.(*mock.BasicMockMessageQueue).Queue) == 1, "expect one messages on the queue")
	message = <-mq.(*mock.BasicMockMessageQueue).Queue
	err = json.Unmarshal(message, &messageStruct)
	th.Assert(t, err == nil, err)
	th.Assert(t, messageStruct.HTTPResponse == 0, fmt.Sprintf("expected to capture 0 response in message, got %v", messageStruct.HTTPResponse))

	// server error response
	ctx = context.Background()

	tc = th.NewTestClientWith404()
	defer tc.Close()

	querierArgs.Client = &(tc.Client)
	args["querierArgs"] = querierArgs

	err = GetAndSendCapabilityStatement(ctx, &args)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(mq.(*mock.BasicMockMessageQueue).Queue) == 1, "expect one message on the queue")
	message = <-mq.(*mock.BasicMockMessageQueue).Queue
	err = json.Unmarshal(message, &messageStruct)
	th.Assert(t, err == nil, err)
	th.Assert(t, messageStruct.HTTPResponse == 404, fmt.Sprintf("expected to capture 404 response in message, got %v", messageStruct.HTTPResponse))
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

	dbHost := viper.GetString("dbhost")
	dbPort := viper.GetInt("dbport")
	dbUser := viper.GetString("dbuser")
	dbPass := viper.GetString("dbpassword")
	dbName := viper.GetString("dbname")
	dbSSL := viper.GetString("dbsslmode")
	store, err = postgresql.NewStore(dbHost, dbPort, dbUser, dbPass, dbName, dbSSL)
	if err != nil {
		return err
	}

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
	endpoints, err = fetcher.GetEndpointsFromFilepath("../../../endpointmanager/resources/EpicEndpointSources.json", "Epic", "Epic", "https://epwebapps.acpny.com/FHIRproxy/api/FHIR/DSTU2/")

	return err
}

func teardown() {
	(*mq).Close()
	channel.Close()
	conn.Close()
	conn.Close()
	store.Close()
}
