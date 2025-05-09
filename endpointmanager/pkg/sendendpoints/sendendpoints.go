package sendendpoints

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/historypruning"

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
		now := time.Now()
		log.Info("Current Time: ", now)

		targetTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 0, 0, 0, now.Location())

		// If the current time is after the target time, set the target time to the next day
		if now.After(targetTime) {
			targetTime = targetTime.Add(24 * time.Hour)
		}

		durationToSleep := time.Until(targetTime)

		// Set the process completion status to true to ensure that the status has not remained false in the case previous process was interrupted and terminated.
		err := store.UpdateProcessCompletionStatus(ctx, "true")
		if err != nil {
			log.Errorf("Failed to set process completion status: %v", err)
		}

		log.Infof("Waiting for %d minutes before processing endpoints", int(durationToSleep.Minutes()))
		time.Sleep(durationToSleep)

		log.Info("Starting daily querying process")

		// Set the process completion status to false to indicate that the process is in progress
		err = store.UpdateProcessCompletionStatus(ctx, "false")
		if err != nil {
			log.Errorf("Failed to set process completion status: %v", err)
		}

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

		if len(listOfEndpoints) != 0 {
			err = accessqueue.SendToQueue(ctx, "FINISHED", mq, channelID, qName)
			if err != nil {
				errs <- err
			}
		}

		// Set the process completion status to true to indicate that the process has completed
		err = store.UpdateProcessCompletionStatus(ctx, "true")
		if err != nil {
			log.Errorf("Failed to set process completion status: %v", err)
		}

		log.Info("Daily querying process complete")

		// Wait 30 minutes to ensure querier is done before starting history pruning and json export
		// time.Sleep(time.Duration(30) * time.Minute)
		//log.Info("Starting json export")
		//err = jsonexport.CreateJSONExport(ctx, store, "/etc/lantern/exportfolder/fhir_endpoints_fields.json", "30days")
		//if err != nil {
		//	log.Infof("Failed to export JSON due to %d", err)
		//If there is an error, wait for the duration to try it again, instead of overloading the memory with repeated tries
		//	log.Infof("Waiting %d minutes", qInterval)
		//	time.Sleep(time.Duration(qInterval) * time.Minute)
		//	errs <- err
		//}
	}
}

func HistoryPruning(
	ctx context.Context,
	wg *sync.WaitGroup,
	qInterval int,
	store *postgresql.Store,
	errs chan<- error) {

	defer wg.Done()

	for {
		log.Info("Starting history pruning")
		historypruning.PruneInfoHistory(ctx, store, true)

		log.Infof("History Pruning complete. Waiting %d minutes", qInterval)
		time.Sleep(time.Duration(qInterval) * time.Minute)
	}
}
