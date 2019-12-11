// +build integration

package chplquerier_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
)

var store *postgresql.Store

func TestMain(m *testing.M) {
	var err error

	err := config.SetupConfigForTests()
	if err != nil {
		return err
	}

	err = setup()
	if err != nil {
		panic(err)
	}

	hap := th.HostAndPort{Host: viper.GetString("dbhost"), Port: viper.GetString("dbport")}
	th.CheckResources(hap)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	teardown()
	os.Exit(code)
}

func Test_Integration_GetCHPLProducts(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	var actualProdsStored int

	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 35,
	}

	// as of 12/5/19, at least 7676 entries are expected to be added to the database
	minNumExpProdsStored := 7676

	err = chplquerier.GetCHPLProducts(ctx, store, client)
	th.Assert(t, err == nil, err)
	rows := store.DB.QueryRow("SELECT COUNT(*) FROM healthit_products;")
	err = rows.Scan(&actualProdsStored)
	th.Assert(t, err == nil, err)
	th.Assert(t, actualProdsStored >= minNumExpProdsStored, fmt.Sprintf("Expected at least %d products stored. Actually had %d products stored.", minNumExpProdsStored, actualProdsStored))

	// expect to see this entry in the DB:
	// {
	// 	"id": 7849,
	// 	"chplProductNumber": "15.04.04.2657.Care.01.00.0.160701",
	// 	"edition": "2015",
	// 	"developer": "Carefluence",
	// 	"product": "Carefluence Open API",
	// 	"version": "1",
	// 	"certificationDate": 1467331200000,
	// 	"certificationStatus": "Active",
	// 	"criteriaMet": "170.315 (d)(1)☺170.315 (d)(10)☺170.315 (d)(9)☺170.315 (g)(4)☺170.315 (g)(5)☺170.315 (g)(6)☺170.315 (g)(7)☺170.315 (g)(8)☺170.315 (g)(9)",
	// 	"apiDocumentation": "170.315 (g)(7)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
	// }
	hitp, err := store.GetHealthITProductUsingNameAndVersion(ctx, "Carefluence Open API", "1")
	th.Assert(t, err == nil, err)
	th.Assert(t, hitp.CHPLID == "15.04.04.2657.Care.01.00.0.160701", "CHPL ID is not what was expected")
	th.Assert(t, hitp.CertificationEdition == "2015", "Certification edition is not what was expected")
	th.Assert(t, hitp.Developer == "Carefluence", "Developer is not what was expected")
	th.Assert(t, hitp.CertificationDate.Equal(time.Unix(1467331200, 0).UTC()), "Certification date is not what was expected")
	th.Assert(t, hitp.CertificationStatus == "Active", "Certification status is not what was expected")
	th.Assert(t, reflect.DeepEqual(hitp.CertificationCriteria, []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"}), "Certification criteria is not what was expected")
	th.Assert(t, hitp.APIURL == "http://carefluence.com/Carefluence-OpenAPI-Documentation.html", "API documentation is not what was expected")
}

func setup() error {
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	if err != nil {
		return err
	}

	return nil
}

func teardown() {
	store.Close()
}
