package rabbitmq

import (
	"errors"
	"fmt"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/streadway/amqp"
)

// Ensure MessageQueue implements lanternmq.MessageQueue.
var _ lanternmq.MessageQueue = &MessageQueue{}

type MessageQueue struct{}

func (mq *MessageQueue) Connect(username string, password string, host string, port string) (*amqp.Connection, error) {
	s := fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port)
	conn, err := amqp.Dial(s)
	if err != nil {
		err = errors.New("unable to connect to message queue")
		return nil, err
	}

	return conn, err
}

func (mq *MessageQueue) CreateChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		err = errors.New("unable to create channel")
		return nil, err
	}

	return ch, err
}

func (mq *MessageQueue) NumConcurrentMsgs(ch *amqp.Channel, num int) error {
	err := ch.Qos(
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

func (mq *MessageQueue) DeclareQueue(ch *amqp.Channel, name string) error {
	_, err := ch.QueueDeclare(
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

func (mq *MessageQueue) PublishToQueue(ch *amqp.Channel, qName string, message string) error {
	err := ch.Publish(
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

func (mq *MessageQueue) ConsumeFromQueue(ch *amqp.Channel, qName string) (<-chan amqp.Delivery, error) {
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

func (mq *MessageQueue) DeclareTarget(ch *amqp.Channel, name string) error {
	err := ch.ExchangeDeclare(
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

func (mq *MessageQueue) PublishToTarget(ch *amqp.Channel, name string, routingKey string, message string) error {
	err := ch.Publish(
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

func (mq *MessageQueue) DeclareTargetReceiveQueue(ch *amqp.Channel, targetName string, qName string, routingKey string) error {
	_, err := ch.QueueDeclare(
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
