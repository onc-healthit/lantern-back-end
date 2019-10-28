package rabbitmq

import (
	"errors"
	"fmt"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/streadway/amqp"
)

// Ensure MessageQueue implements lanternmq.MessageQueue.
var _ lanternmq.MessageQueue = &MessageQueue{}

// Ensure MessageMessages implements lanternmq.Messages.
var _ lanternmq.Messages = &Messages{}

// MessageQueue is a wrapper around RabbitMQ's implementation of queues and implements the lanternmq.MessageQueue
// interface. It allows the user to:
// * connect to a queueing service
// * create a channel for that queueing service
// * state how many messages a receiver can process at one time
// * declare a durable queue, and send and receive from that queue
// * declare a durable topic, and send and receive from that topic
// * close the MessageQueue, which includes closing all channels and the connection to the underlying service.
type MessageQueue struct {
	connection *amqp.Connection
	channels   []*amqp.Channel
}

// Messages wraps the delivery channel.
type Messages struct {
	deliveryChannel <-chan amqp.Delivery
}

// addChannel adds the given channel to the MessageQueue.channels array and returns the
// index to that array casted to a lanternmq.ChannelID.
func (mq *MessageQueue) addChannel(ch *amqp.Channel) (lanternmq.ChannelID, error) {
	if mq.channels == nil {
		mq.channels = []*amqp.Channel{}
	}
	mq.channels = append(mq.channels, ch)
	index := len(mq.channels) - 1
	id := lanternmq.ChannelID(index)

	return id, nil
}

// getChannel retrieves the channel provided by `id` by casting `id` back to an integer and
// retrieving the channel at the corresponding index of MessageQueue.channels array.
func (mq *MessageQueue) getChannel(id lanternmq.ChannelID) (*amqp.Channel, error) {
	idInt, ok := id.(int)
	if !ok {
		return nil, errors.New("ChannelID not of correct type")
	}
	if idInt >= len(mq.channels) {
		return nil, errors.New("no channel with the requested ID was found")
	}
	ch := mq.channels[idInt]
	return ch, nil
}

// Connect creates a connection to a RabbitMQ service.
func (mq *MessageQueue) Connect(username string, password string, host string, port string) error {
	s := fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port)
	conn, err := amqp.Dial(s)
	if err != nil {
		err = errors.New("unable to connect to message queue")
	}
	mq.connection = conn

	return err
}

// CreateChannel creates a channel to the RabbitMQ service that has already been connected to.
// If the RabbitMQ service has not been connected to already, an error is thrown.
// The channel's ID is returned.
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

// NumConcurrentMsgs defines how many messages the user can process from the channel at one time.
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

// DeclareQueue creates a queue with the given name on the given channel using RabbitMQ's
// QueueDeclare method with the following arguments:
// * name: qName
// * durable: true
// * autoDelete: false
// * exclusive: false
// * noWait: false
// * args: nil
func (mq *MessageQueue) DeclareQueue(chID lanternmq.ChannelID, qName string) error {
	ch, err := mq.getChannel(chID)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		qName, // name
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

// PublishToQueue publishes 'message' on the queue with name 'qName' over the channenl with ID 'chID'
// by calling the RabbitMQ Publish method with the following arguments:
// exchange: ""
// key: qName
// mandatory: false
// immediate: false
// publishing:
//   DeliveryMode: amqp.Persistent
//   ContentType: "text/plain"
//   Body: []byte(message)
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

// ConsumeFromQueue opens a receive channel for amqp.Delivery objects for the queue with name 'qName'
// over the channel with ID 'chID'. ConsumeFromQueue wraps amqp.Delivery in a lanternmq.Messages object.
// ConsumeFromQueue creates the receive channel using the RabbitMQ Consume method with the following arguments:
// queue: qName
// consumer: ""
// autoAck: false
// exclusive: false
// noLocal: false
// noWait: false
// args: nil
func (mq *MessageQueue) ConsumeFromQueue(chID lanternmq.ChannelID, qName string) (lanternmq.Messages, error) {
	ch, err := mq.getChannel(chID)
	if err != nil {
		return nil, err
	}

	deliveryChannel, err := ch.Consume(
		qName, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	msgs := Messages{deliveryChannel: deliveryChannel}

	return &msgs, err
}

// ProcessMessages takes 'msgs', which wraps a receive channel for amqp.Delivery objects, and processes each Delivery
// object by retrieving the message from the Delivery object and providing that along with 'args' to the
// lanternmq.MessageHandler 'handler'. An acknowledgement is sent to the sender after each message is processed.
// If there's an error processing a message, the error is sent to the 'errs' channel.
// ProcessMessages should be called as a gorouting. Example:
//     go mq.ProcessMessages(msgs, handler, nil, errs)
func (mq *MessageQueue) ProcessMessages(msgs lanternmq.Messages, handler lanternmq.MessageHandler, args *map[string]interface{}, errs chan<- error) {
	msgsd, ok := msgs.(*Messages)
	if !ok {
		errs <- errors.New("the messages are of the wrong type")
	}

	for d := range msgsd.deliveryChannel {
		err := handler(d.Body, args)
		if err != nil {
			errs <- err
		}
		err = d.Ack(false)
		if err != nil {
			errs <- err
		}
	}
}

// DeclareTopic creates a target named 'name' over the channel with ID 'chID'. It uses RabbitMQ's
// ExchangeDeclare method with the following arguments:
// name: name
// kind: "topic"
// durable: true
// autoDelete: false
// intenral: false
// noWait: false
// args: nil
func (mq *MessageQueue) DeclareTopic(chID lanternmq.ChannelID, name string) error {
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

// PublishToTopic sends 'message' to the topic 'name' over the channel with ID 'chID' with routing key 'routingKey'. It
// uses RabbitMQ's Publish method with the following arguments:
// exchange: name
// key: routingKey
// mandatory: false
// immediate: false
// publishing:
//   ContentType: "text/plain"
//   Body: []byte(message)
func (mq *MessageQueue) PublishToTopic(chID lanternmq.ChannelID, name string, routingKey string, message string) error {
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

// DeclareTopicReceiveQueue creates a queue named 'qName' to receive messages from the topic named 'topicName'
// with routing key 'routingKey' over the channel with ID 'chID'. It uses the RabbitMQ method QueueDeclare with
// the following arguments:
// name: qName
// durable: false
// autoDelete: false
// exclusive: true
// noWait: false
// args: nil
//
// It then calls the RabbitMQ method QueueBind with the following arguments:
// name: qName
// key: routingKey
// exchange: topicName
// noWait: false
// args: nil
func (mq *MessageQueue) DeclareTopicReceiveQueue(chID lanternmq.ChannelID, topicName string, qName string, routingKey string) error {
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
		qName,      // queue name
		routingKey, // routing key
		topicName,  // exchange
		false,
		nil)
	if err != nil {
		err = fmt.Errorf("unable to bind queue %s to target %s with routing key %s", qName, topicName, routingKey)
		return err
	}

	return err
}

// Close closes each channel that's been created, and then closes the connection to the underlying RabbitMQ
// message service.
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
