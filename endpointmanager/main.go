package main

import (
	"database/sql"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	_ "github.com/lib/pq"
)

var db *sql.DB

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
	viper.SetDefault("dbuser", "lantern")
	viper.SetDefault("dbpass", "postgrespassword")
	viper.SetDefault("dbname", "lantern")
	viper.SetDefault("dbsslmode", "disable")
	viper.SetDefault("logfile", "endpointmanagerLog.json")
}

func initializeLogger() {
	log.SetFormatter(&log.JSONFormatter{})
	f, err := os.OpenFile(viper.GetString("logfile"), os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal("LogFile creation error: ", err.Error())
	}
	log.SetOutput(f)
}

func main() {
	var err error

	setupConfig()
	initializeLogger()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		viper.GetString("dbhost"),
		viper.GetInt("dbport"),
		viper.GetString("dbuser"),
		viper.GetString("dbpass"),
		viper.GetString("dbname"),
		viper.GetString("dbsslmode"))

	db, err = sql.Open("postgres", psqlInfo)
	failOnError(err)
	defer db.Close()

	// calling db.Ping to create a connection to the database.
	// db.Open only validates the arguments, it does not create the connection.
	err = db.Ping()
	failOnError(err)

	fmt.Println("Successfully connected!")
}
