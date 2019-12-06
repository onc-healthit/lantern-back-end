// +build integration

package chplquerier_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/mock"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
)

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
	os.Exit(code)
}

func Test_Integration_GetCHPLProducts(t *testing.T) {
	var err error
	var tc *th.TestClient
	var ctx context.Context
	var store endpointmanager.HealthITProductStore

	// as of 12/5/19, at least 7676 entries are expected to be added to the database
	minNumExpProdsStored := 7676

	tc = &http.Client{
		Timeout: time.Second * 35,
	}
	defer tc.Close()

	ctx = context.Background()

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpass"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	th.Assert(t, err == nil, err)
	defer store.Close()

	err = GetCHPLProducts(ctx, store, &(tc.Client))
	th.Assert(t, err == nil, err)
	actualProdsStored := len(store.(*mock.BasicMockStore).HealthITProductData)
	th.Assert(t, actualProdsStored == expectedProdsStored, fmt.Sprintf("Expected %d products stored. Actually had %d products stored.", expectedProdsStored, actualProdsStored))
}

func setup() error {
	err := config.SetupConfigForTests()
	if err != nil {
		return err
	}

	return nil
}

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
