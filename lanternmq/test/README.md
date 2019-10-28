# Testing the Lantern Message Queue Interface Implemented with RabbitMQ

1. Each step assumes that you start in directory `lanternmq/test`

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

1. Send topic messages using a topic string as the argument. The topic strings `error` and `warning` are subscribed to by the receivers. The message that is sent is `hello`.

   ```
   $ cd sendTopic
   $ go run send.go <topic>
   ```

   Example:

   ```
   $ cd sendTopic
   $ go run send.go foo
   $ go run send.go error
   $ go run send.go warning
   $ go run send.go bar
   ```

   Expected result: You should only see the receivers receive the string `hello` when the topics `error` and `warning` are used. An example result using the messages sent above is:


   ```
   $ cd receive
   $ go run receive.go q1
    [*] Waiting for messages. To exit press CTRL+C
    TARGET: Received message: hello
    TARGET: Received message: hello
   ``` 

   ```
   $ cd receive
   $ go run receive.go q2
    [*] Waiting for messages. To exit press CTRL+C
    TARGET: Received message: hello
    TARGET: Received message: hello
   ```

   You will need to observe that the messages received align with when the messages using `error` and `warning` were sent.