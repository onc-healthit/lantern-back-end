// +build integration

package sendendpoints

import (
	"context"
	"fmt"
	"os"
	"time"

	"sync"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

var store *postgresql.Store
var mq *lanternmq.MessageQueue
var chID *lanternmq.ChannelID
var conn *amqp.Connection
var channel *amqp.Channel

var endpts []*endpointmanager.FHIREndpoint = []*endpointmanager.FHIREndpoint{
	&endpointmanager.FHIREndpoint{
		URL: "https://example.com/1",
	},
	&endpointmanager.FHIREndpoint{
		URL: "https://example.com/2",
	},
	&endpointmanager.FHIREndpoint{
		URL: "https://example.com/3",
	},
}

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

func Test_GetEnptsAndSend(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	queueName := viper.GetString("qname")
	queueIsEmpty(t, queueName)
	defer cleanQueue(t, queueName)

	ctx := context.Background()
	var err error

	// populate fhir endpoints
	for _, endpt := range endpts {
		err = store.AddFHIREndpoint(ctx, endpt)
		th.Assert(t, err == nil, err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	errs := make(chan error)
	go GetEnptsAndSend(ctx, &wg, queueName, 1, store, mq, chID, errs)

	// need to pause to ensure all messages are on the queue before we count them
	time.Sleep(10 * time.Second)
	count, err := queueCount(queueName)
	th.Assert(t, err == nil, err)
	// Expect 5 messages: start meesage, 3 endpoints, and stop message
	th.Assert(t, count == 5, fmt.Sprintf("expected there to be 5 messages in the queue, instead got %d", count))
	wg.Done()
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

	mq_, chID_, err := accessqueue.ConnectToServerAndQueue(qUser, qPassword, qHost, qPort, qName)
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

	return nil
}

func teardown() {
	(*mq).Close()
	channel.Close()
	conn.Close()
	store.Close()
}
