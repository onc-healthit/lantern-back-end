# Testing the Lantern Message Queue Interface Implemented with RabbitMQ

Each step assumes that you start in directory `lanternmq/test`. Each go executable file includes hard coded values for subscribing to the queue. These are the default location values for the queue, but may need to be changed if you've implemented the queue differently.

The tests also assume that you have RabbitMQ running. See the [instructions for starting RabbitMQ](../README.md#rabbitmq).

1. In two separate terminals, start two receivers using different argument strings.
   The argument strings represent the identifier for the queue that will receive topics it subscribes to as these must be unique.
   ```
   $ cd receive
   $ go run receive.go q1
   ```
   ```
   $ cd receive
   $ go run receive.go q2
   ```
1. Send queued messages using periods as the argument string. Each receiver waits \<number of periods\> seconds. Send several of these messages to see the receivers sharing the work load.
   ```
   $ cd sendQueue
   $ go run send.go <some number of periods>
   ```

   Example:

   ```
   $ cd sendQueue
   $ go run send.go .
   $ go run send.go ..
   $ go run send.go ...
   $ go run send.go ....
   $ go run send.go .
   $ go run send.go ....
   $ go run send.go .
   $ go run send.go .
   $ go run send.go .
   ...
   ```

   Expected result: You should see the recievers sharing the messages based on how long it takes them to process a message. You should not see the messages alternating between the two receivers. An example result using the messages above sent in quick succession:

   ```
   $ cd receive
   $ go run receive.go q1
    [*] Waiting for messages. To exit press CTRL+C
    QUEUE: Received a message: .
    Done
    QUEUE: Received a message: ...
    Done
    QUEUE: Received a message: .
    Done
    QUEUE: Received a message: ....
    Done
   ``` 

   ```
   $ cd receive
   $ go run receive.go q2
    [*] Waiting for messages. To exit press CTRL+C
    QUEUE: Received a message: ..
    Done
    QUEUE: Received a message: ....
    Done
    QUEUE: Received a message: .
    Done
    QUEUE: Received a message: .
    Done
    QUEUE: Received a message: .
    Done
   ```

1. Send topic messages using a topic string as the first argument and the message as the second argument. If no topic string is provided, `anonymous.info` is used. if no message string is provided, `hello` is used. The topic strings `error` and `warning` are subscribed to by the receivers.

   ```
   $ cd sendTopic
   $ go run send.go <topic>
   ```

   Example:

   ```
   $ cd sendTopic
   $ go run send.go foo foo
   $ go run send.go error error
   $ go run send.go warning warning
   $ go run send.go bar bar
   ```

   Expected result: You should only see the receivers receive the strings associated with the `error` and `warning` topics. An example result using the messages sent above is:


   ```
   $ cd receive
   $ go run receive.go q1
    [*] Waiting for messages. To exit press CTRL+C
    TARGET: Received message: error
    TARGET: Received message: warning
   ``` 

   ```
   $ cd receive
   $ go run receive.go q2
    [*] Waiting for messages. To exit press CTRL+C
    TARGET: Received message: error
    TARGET: Received message: warning
   ```

   You will need to observe that the messages received align with when the messages using `error` and `warning` were sent.