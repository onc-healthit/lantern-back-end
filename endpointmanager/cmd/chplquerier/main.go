package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	log "github.com/sirupsen/logrus"
	"github.com/onc-healthit/lantern-back-end/sharedfunctions"

	"github.com/spf13/viper"
)

func main() {
	var err error

	err = config.SetupConfig()
	sharedfunctions.failOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	sharedfunctions.failOnError("", err)
	defer store.Close()
	log.Info("Successfully connected!")

	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 35,
	}

	// Read version file that is mounted to make user agent
	version, err := ioutil.ReadFile("/etc/lantern/VERSION")
	if err != nil {
		log.Warnf("Cannot read VERSION file")
	}
	versionString := string(version)
	versionNum := strings.Split(versionString, "=")
	userAgent := "LANTERN/" + versionNum[1]
	log.Infof("user agent is %s", userAgent)

	err = chplquerier.GetCHPLCriteria(ctx, store, client, userAgent)
	sharedfunctions.failOnError("", err)
	err = chplquerier.GetCHPLVendors(ctx, store, client, userAgent)
	sharedfunctions.failOnError("", err)
	err = chplquerier.GetCHPLProducts(ctx, store, client, userAgent)
	sharedfunctions.failOnError("", err)
}
