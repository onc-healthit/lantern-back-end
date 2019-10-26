package rabbitmq

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/streadway/amqp"
)

// Ensure MessageQueue implements lanternmq.MessageQueue.
var _ lanternmq.MessageQueue = &MessageQueue{}

type MessageQueue struct {
	connection *amqp.Connection
	channels   []*amqp.Channel
}

func (mq *MessageQueue) addChannel(ch *amqp.Channel) (lanternmq.ChannelID, error) {
	if mq.channels == nil {
		mq.channels = []*amqp.Channel{}
	}
	mq.channels = append(mq.channels, ch)
	index := len(mq.channels) - 1
	id := lanternmq.ChannelID(strconv.Itoa(index))

	return id, nil
}

func (mq *MessageQueue) getChannel(id lanternmq.ChannelID) (*amqp.Channel, error) {
	idStr := string(id)
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, err
	}
	if idInt >= len(mq.channels) {
		return nil, errors.New("no channel with the requested ID was found")
	}
	ch := mq.channels[idInt]
	return ch, nil
}

func (mq *MessageQueue) Connect(username string, password string, host string, port string) error {
	s := fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port)
	conn, err := amqp.Dial(s)
	if err != nil {
		err = errors.New("unable to connect to message queue")
	}
	mq.connection = conn

	return err
}

func (mq *MessageQueue) CreateChannel() (lanternmq.ChannelID, error) {
	var err error
	if mq.connection == nil {
		err = errors.New("connection must exist before creating a channel")
		return "", err
	}
	ch, err := mq.connection.Channel()
	if err != nil {
		err = errors.New("unable to create channel")
		return "", err
	}
	id, err := mq.addChannel(ch)

	return id, err
}

func (mq *MessageQueue) NumConcurrentMsgs(chID lanternmq.ChannelID, num int) error {
	ch, err := mq.getChannel(chID)
	if err != nil {
		return err
	}

	err = ch.Qos(
		num,   // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		err = errors.New("unable to set the number of concurrent messages that can be handled")
		return err
	}
	return err
}

func (mq *MessageQueue) DeclareQueue(chID lanternmq.ChannelID, name string) error {
	ch, err := mq.getChannel(chID)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		err = fmt.Errorf("unable to create queue: %s", err.Error()) //errors.New("unable to create queue")
		return err
	}
	return err
}

func (mq *MessageQueue) PublishToQueue(chID lanternmq.ChannelID, qName string, message string) error {
	ch, err := mq.getChannel(chID)
	if err != nil {
		return err
	}

	err = ch.Publish(
		"",    // exchange
		qName, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
		})
	return err
}

func (mq *MessageQueue) ConsumeFromQueue(chID lanternmq.ChannelID, qName string) (<-chan amqp.Delivery, error) {
	ch, err := mq.getChannel(chID)
	if err != nil {
		return nil, err
	}

	msgs, err := ch.Consume(
		qName, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	return msgs, err
}

func (mq *MessageQueue) ProcessMessages(msgs <-chan amqp.Delivery, handler lanternmq.MessageHandler, args *map[string]interface{}) error {
	for d := range msgs {
		err := handler(d.Body, args)
		if err != nil {
			return err
		}
		d.Ack(false)
	}
	return nil
}

func (mq *MessageQueue) DeclareTarget(chID lanternmq.ChannelID, name string) error {
	ch, err := mq.getChannel(chID)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(
		name,    // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		err = errors.New("unable to declare target")
	}

	return err
}

func (mq *MessageQueue) PublishToTarget(chID lanternmq.ChannelID, name string, routingKey string, message string) error {
	ch, err := mq.getChannel(chID)
	if err != nil {
		return err
	}

	err = ch.Publish(
		name,       // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		err = fmt.Errorf("unable to publish to target %s with routing key %s", name, routingKey)
	}

	return err
}

func (mq *MessageQueue) DeclareTargetReceiveQueue(chID lanternmq.ChannelID, targetName string, qName string, routingKey string) error {
	ch, err := mq.getChannel(chID)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		qName, // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		err = fmt.Errorf("unable to create queue: %s", err.Error()) //errors.New("unable to create queue")
		return err
	}

	err = ch.QueueBind(
		qName,        // queue name
		routingKey,   // routing key
		"logs_topic", // exchange
		false,
		nil)
	if err != nil {
		err = fmt.Errorf("unable to bind queue %s to target %s with routing key %s", qName, targetName, routingKey)
		return err
	}

	return err
}

func (mq *MessageQueue) Close() {
	if mq.channels != nil {
		for _, ch := range mq.channels {
			ch.Close()
		}
	}
	if mq.connection != nil {
		mq.connection.Close()
	}
}
