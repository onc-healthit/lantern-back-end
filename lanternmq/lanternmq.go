package lanternmq

import (
	"github.com/streadway/amqp"
)

type MessageHandler func([]byte, *map[string]interface{}) error

type MessageQueue interface {
	Connect(username string, password string, host string, port string) (*amqp.Connection, error)
	CreateChannel(conn *amqp.Connection) (*amqp.Channel, error)
	NumConcurrentMsgs(ch *amqp.Channel, num int) error
	DeclareQueue(ch *amqp.Channel, name string) error
	PublishToQueue(ch *amqp.Channel, qName string, message string) error
	ConsumeFromQueue(ch *amqp.Channel, qName string) (<-chan amqp.Delivery, error)
	ProcessMessages(msgs <-chan amqp.Delivery, handler MessageHandler, args *map[string]interface{}) error
	DeclareTarget(ch *amqp.Channel, name string) error
	PublishToTarget(ch *amqp.Channel, name string, routingKey string, message string) error
	DeclareTargetReceiveQueue(ch *amqp.Channel, targetName string, qName string, routingKey string) error
}
