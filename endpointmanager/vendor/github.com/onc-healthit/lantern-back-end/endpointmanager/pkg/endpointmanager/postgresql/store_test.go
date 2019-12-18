package postgresql

import (
	"flag"
	"os"
	"testing"

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
	err = viper.BindEnv("logfile")
	failOnError(err)

	viper.SetDefault("dbhost", "localhost")
	viper.SetDefault("dbport", 5432)
	viper.SetDefault("dbsslmode", "disable")
	viper.SetDefault("logfile", "endpointmanagerLog.json")
}

func TestMain(m *testing.M) {
	flag.Parse()

	setupConfig()
	os.Exit(m.Run())
}
