package main

import (
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/nppesquerier"
)

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpass"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	panicOnErr(err)
	fname := "npidata_pfile_20050523-20191110.csv"
	store.DeleteAllNPIOrganizations()
	nppesquerier.ParseAndStoreNPIFile(fname, store)
}
