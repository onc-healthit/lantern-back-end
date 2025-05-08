package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"

	_ "net/http/pprof" // This registers pprof handlers with http.DefaultServeMux

	log "github.com/sirupsen/logrus"
)

// StartProfilingServer starts an HTTP server for profiling
func StartProfilingServer(addr string) {
	// Create a simple handler for the root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Profiling server is running. Visit /debug/pprof/ for the profiling index.")
	})

	// Add a custom memory profiling handler
	http.HandleFunc("/debug/memory", func(w http.ResponseWriter, r *http.Request) {
		log.Info("Capturing memory profile...")

		// Run garbage collection to get more accurate memory usage
		runtime.GC()

		// Create a memory profile file
		f, err := os.Create("/tmp/memory-profile.pprof")
		if err != nil {
			log.Errorf("Could not create memory profile: %v", err)
			http.Error(w, "Could not create memory profile", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		// Write memory profile
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Errorf("Could not write memory profile: %v", err)
			http.Error(w, "Could not write memory profile", http.StatusInternalServerError)
			return
		}

		// Also capture a basic memory statistics summary
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		fmt.Fprintf(w, "Memory profile written to /tmp/memory-profile.pprof\n")
		fmt.Fprintf(w, "Memory stats:\n")
		fmt.Fprintf(w, "Alloc: %v MB\n", memStats.Alloc/1024/1024)
		fmt.Fprintf(w, "TotalAlloc: %v MB\n", memStats.TotalAlloc/1024/1024)
		fmt.Fprintf(w, "Sys: %v MB\n", memStats.Sys/1024/1024)
		fmt.Fprintf(w, "NumGC: %v\n", memStats.NumGC)
	})

	// Start the HTTP server
	log.Infof("Starting profiling server on %s", addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Errorf("Profiling server error: %v", err)
		}
	}()
}
