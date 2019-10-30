package main

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"

	_ "github.com/lib/pq"
)

var db *sql.DB

func setupConfig() {
	var err error

	viper.SetEnvPrefix("lantern_endptmgr")
	viper.AutomaticEnv()

	err = viper.BindEnv("dbhost")
	if err != nil {
		panic(err.Error())
	}
	err = viper.BindEnv("dbport")
	if err != nil {
		panic(err.Error())
	}
	err = viper.BindEnv("dbuser")
	if err != nil {
		panic(err.Error())
	}
	err = viper.BindEnv("dbpass")
	if err != nil {
		panic(err.Error())
	}
	err = viper.BindEnv("dbname")
	if err != nil {
		panic(err.Error())
	}
	err = viper.BindEnv("dbsslmode")
	if err != nil {
		panic(err.Error())
	}

	viper.SetDefault("dbhost", "localhost")
	viper.SetDefault("dbport", 5432)
	viper.SetDefault("dbuser", "postgres")
	viper.SetDefault("dbpass", "")
	viper.SetDefault("dbname", "postgres")
	viper.SetDefault("dbsslmode", "disable")
}

func main() {
	//var endpoint models.FHIREndpoint
	var err error

	setupConfig()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		viper.GetString("dbhost"),
		viper.GetInt("dbport"),
		viper.GetString("dbuser"),
		viper.GetString("dbpass"),
		viper.GetString("dbname"),
		viper.GetString("dbsslmode"))

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// calling db.Ping to create a connection to the database.
	// db.Open only validates the arguments, it does not create the connection.
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}
