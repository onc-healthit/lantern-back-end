// +build integration

package postgresql

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/spf13/viper"
)

var store *Store

func TestMain(m *testing.M) {
	var err error

	err = setup()
	if err != nil {
		panic(err)
	}

	err = checkResources()
	if err != nil {
		panic(err)
	}

	code := m.Run()

	teardown()
	os.Exit(code)
}

func setup() error {
	err := config.SetupConfigForTests()
	if err != nil {
		return err
	}

	store, err = NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	if err != nil {
		return err
	}

	return nil
}

// confirms that the network resources we need are available
func checkResources() error {
	host := viper.GetString("dbhost")
	port := viper.GetString("dbport")

	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		fmt.Println("Connecting error:", err)
	}
	if conn != nil {
		defer conn.Close()
		fmt.Println("Opened", net.JoinHostPort(host, port))
	}

	return nil
}

func teardown() {
	store.Close()
}
