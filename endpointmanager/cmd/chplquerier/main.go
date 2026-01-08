package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

func setupLogging() {
	logDir := "/etc/lantern/logs"
	_ = os.MkdirAll(logDir, 0755)

	logFile := filepath.Join(logDir, "chplquerier_logs.txt")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Warn("Could not open CHPL log file, using console only:", err)
		return
	}

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	setupLogging()

	var err error

	err = config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	defer store.Close()
	log.Info("Successfully connected!")

	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 60,
	}

	// Read version file that is mounted to make user agent
	version, err := os.ReadFile("/etc/lantern/VERSION")
	if err != nil {
		log.Warnf("Cannot read VERSION file")
	}
	versionString := string(version)
	versionNum := strings.Split(versionString, "=")
	userAgent := "LANTERN/" + versionNum[1]
	userAgent = strings.TrimSuffix(userAgent, "\n")
	log.Infof("user agent is %s", userAgent)

	err = chplquerier.GetCHPLCriteria(ctx, store, client, userAgent)
	helpers.FailOnError("", err)
	err = chplquerier.GetCHPLVendors(ctx, store, client, userAgent)
	helpers.FailOnError("", err)
	err = chplquerier.GetCHPLProducts(ctx, store, client, userAgent)
	helpers.FailOnError("", err)
	err = chplquerier.GetCHPLEndpointListProducts(ctx, store)
	helpers.FailOnError("", err)
}
