package sendendpoints

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/historypruning"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/jsonexport"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	log "github.com/sirupsen/logrus"
)

// GetEnptsAndSend gets the current list of endpoints from the database and sends each one to the given queue
// it continues to repeat this action every time the given interval period has passed
func GetEnptsAndSend(
	ctx context.Context,
	wg *sync.WaitGroup,
	qName string,
	qInterval int,
	store *postgresql.Store,
	mq *lanternmq.MessageQueue,
	channelID *lanternmq.ChannelID,
	errs chan<- error) {

	defer wg.Done()

	for {
		listOfEndpoints, err := store.GetAllDistinctFHIREndpoints(ctx)
		if err != nil {
			errs <- err
		}

		// Shuffle Endpoints So that We Are Not Querying As Rapidly
		rand.Shuffle(len(listOfEndpoints), func(i, j int) {
			listOfEndpoints[i], listOfEndpoints[j] = listOfEndpoints[j], listOfEndpoints[i]
		})

		for i, endpt := range listOfEndpoints {
			if i%10 == 0 {
				log.Infof("Processed %d/%d messages", i, len(listOfEndpoints))
			}
			// Add a short time buffer as we enqueue items
			time.Sleep(time.Duration(500 * time.Millisecond))
			err = accessqueue.SendToQueue(ctx, endpt.URL, mq, channelID, qName)
			if err != nil {
				errs <- err
			}
		}

		log.Infof("Waiting %d minutes", qInterval)
		if len(listOfEndpoints) != 0 {
			err = accessqueue.SendToQueue(ctx, "FINISHED", mq, channelID, qName)
			if err != nil {
				errs <- err
			}
		}
		// Wait 1 second for every endpoint to ensure querier is done before starting history pruning and json export
		time.Sleep(time.Duration(time.Duration(len(listOfEndpoints)) * time.Second))
		historypruning.PruneInfoHistory(ctx, store, true)
		err = jsonexport.CreateJSONExport(ctx, store, "/etc/lantern/exportfolder/fhir_endpoints_fields.json")
		if err != nil {
			errs <- err
		}

		time.Sleep(time.Duration(qInterval) * time.Minute)
	}
}
