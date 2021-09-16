// +build integration

package accessqueue_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	aq "github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	"github.com/onc-healthit/lantern-back-end/lanternmq/rabbitmq"
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

	time.Sleep(time.Duration(60) * time.Second)

	// ack the message
	delivery, deliveryBool, err := channel.Get(qName, true)
	th.Assert(t, err == nil, err)

	count, err = aq.QueueCount(qName, channel)
	th.Assert(t, err == nil, err)
	th.Assert(t, count != 1, fmt.Sprintf("There are %d messages in the queue. Message count is %v. Delivery bool is %v", count, delivery.MessageCount, deliveryBool))
		
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
