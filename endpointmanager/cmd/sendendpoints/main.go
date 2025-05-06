package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof" // Import pprof handlers

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	memprofile  = flag.String("memprofile", "", "write memory profile to `file`")
	cpuprofile  = flag.String("cpuprofile", "", "write cpu profile to `file`")
	httpProfile = flag.String("httpprofile", "", "enable http profiling on specified address, e.g., ':6060'")
	profileFreq = flag.Duration("profilefreq", 60*time.Minute, "frequency to collect memory profiles")
)

func main() {
	flag.Parse()

	// Set up CPU profiling if requested
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatalf("Could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("Could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Set up HTTP profiling server if requested
	if *httpProfile != "" {
		go func() {
			log.Printf("Starting pprof HTTP server on %s", *httpProfile)
			log.Println(http.ListenAndServe(*httpProfile, nil))
		}()
	}

	// Set up periodic memory profiling if requested
	if *memprofile != "" && *profileFreq > 0 {
		go func() {
			for {
				time.Sleep(*profileFreq)
				writeMemProfile(*memprofile)
			}
		}()
	}

	// Original main functionality
	var err error

	err = config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(
		viper.GetString("dbhost"),
		viper.GetInt("dbport"),
		viper.GetString("dbuser"),
		viper.GetString("dbpassword"),
		viper.GetString("dbname"),
		viper.GetString("dbsslmode"),
	)
	helpers.FailOnError("", err)
	defer store.Close()

	// Set up signal handler for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-signalChan:
			log.Println("Received shutdown signal, stopping...")
			// Capture a memory profile on shutdown
			if *memprofile != "" {
				writeMemProfile(fmt.Sprintf("%s.shutdown", *memprofile))
			}
			cancel()
		case <-ctx.Done():
		}
	}()

	// Configure logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	for {
		// Force garbage collection before starting a new cycle
		runtime.GC()

		// Show memory stats before starting
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		logrus.WithFields(logrus.Fields{
			"alloc_mb":       m.Alloc / 1024 / 1024,
			"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
			"sys_mb":         m.Sys / 1024 / 1024,
			"num_gc":         m.NumGC,
		}).Info("Memory stats before endpoint processing")

		// Run the endpoint manager
		err = sendendpoints.SendFHIREndpointsToQueue(ctx, store)
		if err != nil {
			if ctx.Err() != nil {
				// Context cancelled, shutting down
				break
			}
			logrus.WithError(err).Error("Error sending FHIR endpoints to queue")
		}

		// Show memory stats after completion
		runtime.ReadMemStats(&m)
		logrus.WithFields(logrus.Fields{
			"alloc_mb":       m.Alloc / 1024 / 1024,
			"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
			"sys_mb":         m.Sys / 1024 / 1024,
			"num_gc":         m.NumGC,
		}).Info("Memory stats after endpoint processing")

		// Capture memory profile after processing
		if *memprofile != "" {
			writeMemProfile(fmt.Sprintf("%s.%d", *memprofile, time.Now().Unix()))
		}

		// Get the query interval from configuration
		queryInterval := viper.GetInt("capquery_qryintvl")
		interval := time.Duration(queryInterval) * time.Minute
		logrus.WithField("interval_minutes", queryInterval).Info("Waiting for next cycle")

		// Wait for interval or shutdown signal
		select {
		case <-time.After(interval):
			// Continue to next cycle
		case <-ctx.Done():
			// Shutdown requested
			logrus.Info("Shutdown requested, exiting...")
			return
		}
	}
}

func writeMemProfile(filename string) {
	// Force garbage collection before profiling
	runtime.GC()

	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Could not create memory profile: %v", err)
		return
	}
	defer f.Close()

	log.Printf("Writing memory profile to %s", filename)
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Printf("Could not write memory profile: %v", err)
	}
}
