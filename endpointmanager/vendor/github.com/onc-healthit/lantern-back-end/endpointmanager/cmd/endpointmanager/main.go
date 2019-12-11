package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func setupConfig() {
	var err error

	viper.SetEnvPrefix("lantern_endptmgr")
	viper.AutomaticEnv()

	err = viper.BindEnv("dbhost")
	failOnError(err)
	err = viper.BindEnv("dbport")
	failOnError(err)
	err = viper.BindEnv("dbuser")
	failOnError(err)
	err = viper.BindEnv("dbpass")
	failOnError(err)
	err = viper.BindEnv("dbname")
	failOnError(err)
	err = viper.BindEnv("dbsslmode")
	failOnError(err)
	err = viper.BindEnv("chplapikey")
	failOnError(err)

	viper.SetDefault("dbhost", "localhost")
	viper.SetDefault("dbport", 5432)
	viper.SetDefault("dbuser", "lantern")
	viper.SetDefault("dbpass", "postgrespassword")
	viper.SetDefault("dbname", "lantern")
	viper.SetDefault("dbsslmode", "disable")
}

func main() {
	var err error

	setupConfig()

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpass"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	if err != nil {
		panic(err.Error())
	}
	defer store.Close()
	fmt.Println("Successfully connected!")

	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 35,
	}
	err = chplquerier.GetCHPLProducts(ctx, store, client)
	if err != nil {
		panic(err)
	}
}
