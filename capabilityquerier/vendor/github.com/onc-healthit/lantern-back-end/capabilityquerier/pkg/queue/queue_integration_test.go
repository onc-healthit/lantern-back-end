// +build integration

package queue_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/queue"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/pkg/errors"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/config"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var qUser, qPassword, qHost, qPort, qName string

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

func Test_ConnectToQueue(t *testing.T) {
	queueIsEmpty(t, qName)
	defer cleanQueue(t, qName)

	var mq lanternmq.MessageQueue
	var chID lanternmq.ChannelID
	var err error

	// bad host
	_, _, err = queue.ConnectToQueue(qUser, qPassword, "asdf", qPort, qName)
	th.Assert(t, errors.Cause(err).Error() == "unable to connect to message queue", "expected error due to bad host")

	// bad port
	_, _, err = queue.ConnectToQueue(qUser, qPassword, qHost, "8080", qName)
	th.Assert(t, errors.Cause(err).Error() == "unable to connect to message queue", "expected error due to bad port")

	// bad username
	_, _, err = queue.ConnectToQueue("nottheuser", qPassword, qHost, qPort, qName)
	th.Assert(t, errors.Cause(err).Error() == "unable to connect to message queue", "expected error due to bad username")

	// bad password
	_, _, err = queue.ConnectToQueue(qUser, "notthepassword", qHost, qPort, qName)
	th.Assert(t, errors.Cause(err).Error() == "unable to connect to message queue", "expected error due to bad password")

	// bad queue name
	_, _, err = queue.ConnectToQueue(qUser, qPassword, qHost, qPort, "notthequeuename")
	th.Assert(t, errors.Cause(err).Error() == "queue notthequeuename does not exist", "expected error due to bad queue name")

	// all ok
	mq, chID, err = queue.ConnectToQueue(qUser, qPassword, qHost, qPort, qName)
	th.Assert(t, err == nil, err)
	th.Assert(t, mq != nil, "expected message queue to be created")
	th.Assert(t, chID != nil, "expected channel ID to be created")
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

	qUser = viper.GetString("quser")
	qPassword = viper.GetString("qpassword")
	qHost = viper.GetString("qhost")
	qPort = viper.GetString("qport")
	qName = viper.GetString("capquery_qname")

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

func teardown() {
}
