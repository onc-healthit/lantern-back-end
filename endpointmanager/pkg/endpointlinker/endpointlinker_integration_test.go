package endpointlinker

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
)

var store *postgresql.Store

func TestMain(m *testing.M) {
	var err error

	err = config.SetupConfigForTests()
	if err != nil {
		panic(err)
	}

	err = setup()
	if err != nil {
		panic(err)
	}

	hap := th.HostAndPort{Host: viper.GetString("dbhost"), Port: viper.GetString("dbport")}
	err = th.CheckResources(hap)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	teardown()
	os.Exit(code)
}

func Test_matchByID(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var ep = &endpointmanager.FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"FOO FOO BAR"},
		NPIIDs:            []string{"1", "2", "3"},
		ListSource:        "https://open.epic.com/MyApps/EndpointsJson"}

	ctx := context.Background()

	// test with no orgs
	matches, confidences, err := matchByID(ctx, ep, store, false)
	expected := 0
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, "expected no matches")
	th.Assert(t, len(confidences) == expected, "expected no confidences")

	// test with non matching org
	err = store.AddNPIOrganization(ctx, nonMatchingOrg)
	th.Assert(t, err == nil, err)
	matches, confidences, err = matchByID(ctx, ep, store, false)
	expected = 0
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, "expected no matches")
	th.Assert(t, len(confidences) == expected, "expected no confidences")

	err = store.AddNPIOrganization(ctx, exactPrimaryNameOrg)
	th.Assert(t, err == nil, err)
	err = store.AddNPIOrganization(ctx, nonExactSecondaryNameOrg)
	th.Assert(t, err == nil, err)
	err = store.AddNPIOrganization(ctx, exactSecondaryNameOrg)
	th.Assert(t, err == nil, err)
	err = store.AddNPIOrganization(ctx, exactSecondaryNameOrgNoPrimaryName)
	th.Assert(t, err == nil, err)

	// test with single match
	ep.NPIIDs = []string{"1"}
	matches, confidences, err = matchByID(ctx, ep, store, false)
	expected = 1
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	org := exactPrimaryNameOrg
	expectedConf := 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s to match %v with confidence %f. got %f", org.NPI_ID, ep.NPIIDs, expectedConf, confidences[org.NPI_ID]))

	// test with multiple matches
	ep.NPIIDs = []string{"1", "2", "3", "4"}
	matches, confidences, err = matchByID(ctx, ep, store, false)
	expected = 3 // no org w id "3"
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	org = exactPrimaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s to match %v with confidence %f. got %f", org.NPI_ID, ep.NPIIDs, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s to match %v with confidence %f. got %f", org.NPI_ID, ep.NPIIDs, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s to match %v with confidence %f. got %f", org.NPI_ID, ep.NPIIDs, expectedConf, confidences[org.NPI_ID]))
}

func Test_addMatch(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	ctx := context.Background()
	ep := &endpointmanager.FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"FOO FOO BAR"},
		NPIIDs:            []string{},
		ListSource:        "https://open.epic.com/MyApps/EndpointsJson"}
	npiID := "1"
	confidence := .6

	// add new match
	err := addMatch(ctx, store, npiID, ep, confidence)
	th.Assert(t, err == nil, err)
	sNpiID, sEpURL, sConfidence, err := store.GetNPIOrganizationFHIREndpointLink(ctx, npiID, ep.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, sNpiID == npiID, fmt.Sprintf("expected stored ID '%s' to be the same as the ID that was stored '%s'.", sNpiID, npiID))
	th.Assert(t, sEpURL == ep.URL, fmt.Sprintf("expected stored url '%s' to be the same as the url that was stored '%s'.", sEpURL, ep.URL))
	th.Assert(t, sConfidence == confidence, fmt.Sprintf("expected stored confidence '%f' to be the same as the confidence that was stored '%f'.", sConfidence, confidence))

	// update match, lower confidence
	newConfidence := .5
	err = addMatch(ctx, store, npiID, ep, newConfidence)
	th.Assert(t, err == nil, err)
	sNpiID, sEpURL, sConfidence, err = store.GetNPIOrganizationFHIREndpointLink(ctx, npiID, ep.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, sNpiID == npiID, fmt.Sprintf("expected stored ID '%s' to be the same as the ID that was stored '%s'.", sNpiID, npiID))
	th.Assert(t, sEpURL == ep.URL, fmt.Sprintf("expected stored url '%s' to be the same as the url that was stored '%s'.", sEpURL, ep.URL))
	th.Assert(t, sConfidence == confidence, fmt.Sprintf("expected stored confidence '%f' to be the same as the original confidence that was stored '%f'.", sConfidence, confidence))

	// update match, higher confidence
	newConfidence = .7
	err = addMatch(ctx, store, npiID, ep, newConfidence)
	th.Assert(t, err == nil, err)
	sNpiID, sEpURL, sConfidence, err = store.GetNPIOrganizationFHIREndpointLink(ctx, npiID, ep.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, sNpiID == npiID, fmt.Sprintf("expected stored ID '%s' to be the same as the ID that was stored '%s'.", sNpiID, npiID))
	th.Assert(t, sEpURL == ep.URL, fmt.Sprintf("expected stored url '%s' to be the same as the url that was stored '%s'.", sEpURL, ep.URL))
	th.Assert(t, sConfidence == newConfidence, fmt.Sprintf("expected stored confidence '%f' to be the same as the new confidence that was stored '%f'.", sConfidence, newConfidence))
}

func setup() error {
	var err error
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	if err != nil {
		return err
	}

	return nil
}

func teardown() {
	store.Close()
}
