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
	eps "github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/endpoints"
	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/queue"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/prometheus/common/log"
	"github.com/streadway/amqp"

	"github.com/onc-healthit/lantern-back-end/lanternmq"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/config"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
	"github.com/spf13/viper"
)

var mq *lanternmq.MessageQueue
var chID *lanternmq.ChannelID
var endpoints *fetcher.ListOfEndpoints

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
	queueName := viper.GetString("capquery_qname")
	queueIsEmpty(t, queueName)
	defer cleanQueue(t, queueName)

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
			err = capabilityquerier.GetAndSendCapabilityStatement(ctx, metadataURL, client, mq, chID, queueName)
			th.Assert(t, err == nil, err)
		}
	}
	count, err := queueCount(queueName)
	th.Assert(t, err == nil, err)
	// need to pause to ensure all messages are on the queue before we count them
	time.Sleep(10 * time.Second)
	th.Assert(t, count >= 9, fmt.Sprintf("expected there to be 9 or 10 messages in the queue (difference because dealing with real data and endpoints); saw %d", count))
}

func queueIsEmpty(t *testing.T, queueName string) {
	count, err := queueCount(queueName)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 0, "should be no messages in queue.")
}

func cleanQueue(t *testing.T, queueName string) {
	_, err := channel.QueuePurge(queueName, false)
	th.Assert(t, err == nil, err)
	count, err := queueCount(queueName)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 0, "should be no messages in queue.")
}

func queueCount(queueName string) (int, error) {
	queue, err := channel.QueueDeclarePassive(
		queueName,
		true,
		false,
		false,
		false,
		nil, // args
	)
	if err != nil {
		return -1, err
	}

	return queue.Messages, nil
}

func setup() error {
	var err error

	qUser := viper.GetString("quser")
	qPassword := viper.GetString("qpassword")
	qHost := viper.GetString("qhost")
	qPort := viper.GetString("qport")
	qName := viper.GetString("capquery_qname")

	// set up wrapped queue info
	mq_, chID_, err := queue.ConnectToQueue(qUser, qPassword, qHost, qPort, qName)
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
	endpoints, err = eps.GetEndpoints("../../../networkstatsquerier/resources/EndpointSources.json")

	return err
}

func teardown() {
	(*mq).Close()
	channel.Close()
	conn.Close()
}
