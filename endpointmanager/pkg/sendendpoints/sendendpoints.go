package sendendpoints

package main

import (
	"context"
	"sync"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
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
		listOfEndpoints, err := store.GetAllFHIREndpoints(ctx)
		if err != nil {
			errs <- err
		}

		err = accessqueue.SendToQueue(ctx, "start", mq, channelID, qName)
		if err != nil {
			errs <- err
		}

		for i, endpt := range listOfEndpoints {
			if i%10 == 0 {
				log.Infof("Processed %d/%d messages", i, len(listOfEndpoints))
			}

			err = accessqueue.SendToQueue(ctx, endpt.URL, mq, channelID, qName)
			if err != nil {
				errs <- err
			}
		}

		err = accessqueue.SendToQueue(ctx, "stop", mq, channelID, qName)
		if err != nil {
			errs <- err
		}

		log.Infof("Waiting %d minutes", qInterval)
		time.Sleep(time.Duration(qInterval) * time.Minute)
	}
}