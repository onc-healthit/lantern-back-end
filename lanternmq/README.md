# LanternMQ

LanternMQ is a go package that facilitates the messaging infrastructure for the Lantern microservices. LanternMQ provides a simple interface for sending and receiving messages in a queue, and for sending and receiving topic messages.

The package includes a RabbitMQ implementation for the LanternMQ interface.

The package also includes a mock implementation for the LanternMQ interface to support testing.

To test the package, see the [testing instructions](test/README.md).

## Updating Users for RabbitMQ

The default users, their password hashes, and each user's permissions can be found in `lantern/definitions.json`.

**To update users from the RabbitMQ browser interface:**
1. Log-in as an admin at `localhost:15672`
2. Go to the *Admin* tab
3. Click the specific user to update
4. Near the bottom of the page click *Update this User*
5. Set a new password

A new user can be added by clicking *Add a user* in the *Admin* tab.
  
**To update users from the command line:** <br>
Run `docker exec -it lantern-back-end_lantern-mq_1 rabbitmqctl change_password <username> <new password>`

A new user can be added by replacing `change_password` in the above line to `add_user`.

**The definitions.json file must be updated to persist these changes. To update the definitions.json file:**
1. Get the updated JSON object by using the RabbitMQ API <br>
  (e.g. `curl -H "Accept:application/json" -u <management_username>:<management_password> "localhost:15672/api/definitions"`)
2. Replace the current definitions file with the response from Step 1

The two steps can also be combined into one command: `curl -H "Accept:application/json" -u <management_username>:<management_password> "localhost:15672/api/definitions" > lanternmq/definitions.json`

## Scaling

When scaling out the number of capability queurier services, lanternmq will create a new queue per capability querier to receive the start/stop broadcast message from the endpoint manager. 

To scale out the capability querier service edit the docker-compose.yml file 
to include another capabilityQuerier service. Under the environment define another name for the LANTERN_BROADCAST_QUEUE variable
```
capability_querier_2:
    environment:
        - LANTERN_BROADCAST_QUEUE=broadcast_queue_2
``` 

The value for LANTERN_BROADCAST_QUEUE can either be defined in directly in the docker-compose of separately in your .env file. Each capabilityQuerier must have a unique value for the LANTERN_BROADCAST_QUEUE variable.