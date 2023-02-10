package main

import (
	"database/sql"
	"fmt"
	"os"

	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {

	viper.SetEnvPrefix("lantern")
	viper.AutomaticEnv()
	host := viper.GetString("dbhost")
	port := viper.GetInt("dbport")
	user := viper.GetString("dbuser")
	password := viper.GetString("dbpassword")
	dbname := viper.GetString("dbname")
	sslmode := viper.GetString("dbsslmode")

	var forceVersion int
	var direction string
	var err error
	if len(os.Args) > 2 {
		direction = os.Args[1]
		forceVersion, err = strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("ERROR: Could not convert force version from string to int")
		}
	} else if len(os.Args) > 1 {
		direction = os.Args[1]
		forceVersion = -1
	} else {
		forceVersion = -1
		direction = "up"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Error("endpoint URL parsing error (open postgres): ", err.Error())
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Error("endpoint URL parsing error (with instance postgres): ", err.Error())
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)

	if err != nil {
		log.Fatalf("ERROR: %s", err.Error())
	}

	if forceVersion > -1 {
		m.Force(forceVersion)
	}

	stepDirection := 1
	if direction == "down" {
		stepDirection = -1
	}

	if err := m.Steps(stepDirection); err != nil {
		version, dirty, retError := m.Version()
		fmt.Printf("Version %+v with Dirty Flag %+v threw Error \n %+v", version, dirty, retError)
		log.Fatal(err)
	}
}
