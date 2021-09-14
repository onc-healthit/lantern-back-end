// +build integration

package accessqueue_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	aq "github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	"github.com/onc-healthit/lantern-back-end/lanternmq/rabbitmq"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var qUser, qPassword, qHost, qPort, qName string

var mq *lanternmq.MessageQueue
var conn *amqp.Connection
var channel *amqp.Channel

func TestMain(m *testing.M) {
	var err error

	err = setupConfigForTests()
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

func Test_ConnectToServerAndQueue(t *testing.T) {
	queueIsEmpty(t, qName)
	defer checkCleanQueue(t, qName, channel)

	var chID lanternmq.ChannelID
	var err error

	// bad host
	_, _, err = aq.ConnectToServerAndQueue(qUser, qPassword, "asdf", qPort, qName)
	th.Assert(t, errors.Cause(err).Error() == "unable to connect to message queue", "expected error due to bad host")

	// bad port
	_, _, err = aq.ConnectToServerAndQueue(qUser, qPassword, qHost, "8080", qName)
	th.Assert(t, errors.Cause(err).Error() == "unable to connect to message queue", "expected error due to bad port")

	// bad username
	_, _, err = aq.ConnectToServerAndQueue("nottheuser", qPassword, qHost, qPort, qName)
	th.Assert(t, errors.Cause(err).Error() == "unable to connect to message queue", "expected error due to bad username")

	// bad password
	_, _, err = aq.ConnectToServerAndQueue(qUser, "notthepassword", qHost, qPort, qName)
	th.Assert(t, errors.Cause(err).Error() == "unable to connect to message queue", "expected error due to bad password")

	// bad queue name
	_, _, err = aq.ConnectToServerAndQueue(qUser, qPassword, qHost, qPort, "notthequeuename")
	th.Assert(t, errors.Cause(err).Error() == "queue notthequeuename does not exist", "expected error due to bad queue name")

	// all ok
	mq_, chID, err := aq.ConnectToServerAndQueue(qUser, qPassword, qHost, qPort, qName)
	mq = &mq_
	th.Assert(t, err == nil, err)
	th.Assert(t, mq != nil, "expected message queue to be created")
	th.Assert(t, chID != nil, "expected channel ID to be created")
}

func Test_ConnectToQueue(t *testing.T) {
	queueIsEmpty(t, qName)
	defer checkCleanQueue(t, qName, channel)

	var err error

	// set up

	mq2 := &rabbitmq.MessageQueue{}
	err = mq2.Connect(qUser, qPassword, qHost, qPort)
	th.Assert(t, err == nil, "unable to connect to message queue server")
	ch, err := mq2.CreateChannel()
	th.Assert(t, err == nil, "unable to create channel to message queue server")

	// all ok
	mq_, _, err := aq.ConnectToQueue(mq2, ch, qName)
	mq = &mq_
	th.Assert(t, err == nil, err)

	// queue does not exist
	_, _, err = aq.ConnectToQueue(mq2, ch, "nonsense")
	th.Assert(t, errors.Cause(err).Error() == "queue nonsense does not exist", err)

	// channel does not exist
	_, _, err = aq.ConnectToQueue(mq2, "ch", qName)
	th.Assert(t, errors.Cause(err).Error() == "ChannelID not of correct type", "given channel should not exist")
}

func Test_CleanQueue(t *testing.T) {
	queueIsEmpty(t, qName)
	defer checkCleanQueue(t, qName, channel)

	var err error

	// set up
	mq2 := &rabbitmq.MessageQueue{}
	err = mq2.Connect(qUser, qPassword, qHost, qPort)
	th.Assert(t, err == nil, "unable to connect to message queue server")
	ch, err := mq2.CreateChannel()
	th.Assert(t, err == nil, "unable to create channel to message queue server")

	// all ok
	mq_, _, err := aq.ConnectToQueue(mq2, ch, qName)
	mq = &mq_
	th.Assert(t, err == nil, err)

	// add message to queue then clean
	ctx := context.Background()
	err = aq.SendToQueue(ctx, "clean queue message", mq, &ch, qName)
	th.Assert(t, err == nil, err)

	// ack the message so it will get purged
	_, _, err = channel.Get(qName, true)
	th.Assert(t, err == nil, err)

	err = aq.CleanQueue(qName, channel)
	th.Assert(t, err == nil, err)
}

func Test_QueueCount(t *testing.T) {
	queueIsEmpty(t, qName)
	defer checkCleanQueue(t, qName, channel)

	var err error

	// set up
	mq2 := &rabbitmq.MessageQueue{}
	err = mq2.Connect(qUser, qPassword, qHost, qPort)
	th.Assert(t, err == nil, "unable to connect to message queue server")
	ch, err := mq2.CreateChannel()
	th.Assert(t, err == nil, "unable to create channel to message queue server")

	// all ok
	mq_, _, err := aq.ConnectToQueue(mq2, ch, qName)
	mq = &mq_
	th.Assert(t, err == nil, err)

	// base test
	count, err := aq.QueueCount(qName, channel)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 0, "there should be no messages in the queue")

	// add message to queue
	ctx := context.Background()
	err = aq.SendToQueue(ctx, "queue count message", mq, &ch, qName)
	th.Assert(t, err == nil, err)
	// ack the message
	msg, deliveryOk, err := channel.Get(qName, true)
	th.Assert(t, err == nil, err)

	count, err = aq.QueueCount(qName, channel)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, fmt.Sprintf("there should be one message in the queue, instead there are %d. The delivery bool was %v. The message count was %v, body was %v, expiration was %v, delivery tag was %v", count, deliveryOk, msg.MessageCount, msg.Body, msg.Expiration, msg.DeliveryTag))
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

	qUser = viper.GetString("quser")
	qPassword = viper.GetString("qpassword")
	qHost = viper.GetString("qhost")
	qPort = viper.GetString("qport")
	qName = viper.GetString("qname")

	fmt.Printf("amqp://%s:%s@%s:%s/", qUser, qPassword, qHost, qPort)
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

	return err
}

func setupConfigForTests() error {
	var err error

	viper.SetEnvPrefix("lantern")
	viper.AutomaticEnv()

	err = viper.BindEnv("qhost")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qport")
	if err != nil {
		return err
	}
	err = viper.BindEnv("capquery_qname")
	if err != nil {
		return err
	}

	viper.SetDefault("qhost", "localhost")
	viper.SetDefault("qport", "5672")
	viper.SetDefault("capquery_qname", "capability-statements")

	prevQName := viper.GetString("capquery_qname")

	viper.SetEnvPrefix("lantern_test")
	viper.AutomaticEnv()

	err = viper.BindEnv("quser")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qpassword")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qname")
	if err != nil {
		return err
	}

	viper.SetDefault("quser", "capabilityquerier")
	viper.SetDefault("qpassword", "capabilityquerier")
	viper.SetDefault("qname", "test-queue")

	if prevQName == viper.GetString("qname") {
		panic("Test queue and dev/prod queue must be different. Test queue: " + viper.GetString("qname") + ". Prod/Dev queue: " + prevQName)
	}

	return nil
}

func teardown() {
	(*mq).Close()
	channel.Close()
	conn.Close()
}
