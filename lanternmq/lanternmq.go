package lanternmq

// MessageQueue is an interface for writing messages to either a basic queue or a topic. Below are
// some usage examples.
//
//
// Example: Publish a message to a queue
// --------
// mq := <implementation of MessageQueue
// err := mq.Connect("guest", "guest", "localhost", "5672")
// chID, err := mq.CreateChannel()
// err = mq.DeclareQueue(chID, "queueName")
// err = mq.PublishToQueue(chID, "queueName", "message")
//
//
// Example: Read a message from a queue
// --------
// mq := <implementation of MessageQueue
// err := mq.Connect("guest", "guest", "localhost", "5672")
// chID, err := mq.CreateChannel()
// err = mq.DeclareQueue(chID, "queueName")
// msgs, err := mq.ConsumeFromQueue(chID, "queueName")
// forever := make(chan bool)
// errs := make(chan error)
// go mq.ProcessMessages(
// 		msgs,
// 		func(msg []byte, _ *map[string]interface{}) error {
//			fmt.Printf("Received message: %s\n")
// 		},
// 		nil,
//      errs)
// <-forever
//
// Example: Publish a message to a topic
// --------
// mq := <implementation of MessageQueue
// err := mq.Connect("guest", "guest", "localhost", "5672")
// chID, err := mq.CreateChannel()
// err = mq.DeclareTopic(chID, "topicName")
// err = mq.PublishToTopic(chID, "topicName", "topicRoutingKey", "message")
//
//
// Example: Read a message from a topic
// --------
// mq := <implementation of MessageQueue
// err := mq.Connect("guest", "guest", "localhost", "5672")
// chID, err := mq.CreateChannel()
// err = mq.DeclareTopic(chID, "topicName")
// err = mq.DeclareTopicReceiveQueue(chID, "topicName", "queueName", "topicRoutingKey")
// msgs, err := mq.ConsumeFromQueue(chID, "queueName")
// forever := make(chan bool)
// errs := make(chan error)
// go mq.ProcessMessages(
// 		msgs,
// 		func(msg []byte, _ *map[string]interface{}) error {
//			fmt.Printf("Received message: %s\n")
// 		},
// 		nil,
//      errs)
// <-forever
type MessageQueue interface {
	// Connect opens a connection with the underlying queuing service.
	Connect(username string, password string, host string, port string) error
	// CreateChannel opens a channel associated with the connected queuing service.
	CreateChannel() (ChannelID, error)
	// NumConcurrentMsgs defines how many messages can be processed in parallel.
	NumConcurrentMsgs(chID ChannelID, num int) error
	// DeclareQueue creates a queue with the name 'qName' on the channel with ID 'chID' if one
	// does not exist.
	DeclareQueue(chID ChannelID, qName string) error
	// PublishToQueue sends 'message' to the queue with name 'qName' over the channel with ID
	// 'chID'.
	PublishToQueue(chID ChannelID, qName string, message string) error
	// ConsumeFromQueue returns an instance of Messages, which acts like the receiving channel
	// for any messages that present on queue 'qName' on the channel with ID 'chID'.
	ConsumeFromQueue(chID ChannelID, qName string) (Messages, error)
	// ProcessMessages applies the 'handler' MessageHandler with arguments 'args' to each
	// message that is received through 'msgs'. Sends any errors to the 'errs' channel.
	ProcessMessages(msgs Messages, handler MessageHandler, args *map[string]interface{}, errs chan<- error)
	// DeclareTopic creates a topic with the name 'name' on the channel with ID 'chID' if one
	// does not exist.
	DeclareTopic(chID ChannelID, name string) error
	// PublishToTopics sends 'message' to the topic 'name' on channel with ID 'chID', which will be
	// routed to receivers using 'routingKey'.
	PublishToTopic(chID ChannelID, name string, routingKey string, message string) error
	// DeclareTopicReceiveQueue creates queue with name 'qName' associated to the topic with name
	// 'topicName' on the channel with ID 'chID' to receive messages routed with the routing key 'routingKey'.
	DeclareTopicReceiveQueue(chID ChannelID, topicName string, qName string, routingKey string) error
	// Close closes the MessageQueue and any associated resources including associated channels and the
	// connection to the underlying queuing service.
	Close()
}

// Messages is the stream of messages that will be received from a queue.
type Messages interface{}

// ChannelID is the identifier for a channel.
type ChannelID interface{}

// MessageHandler is a function to process an individual message.
type MessageHandler func([]byte, *map[string]interface{}) error
