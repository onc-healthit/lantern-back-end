package chplmapper

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/mock"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_getVendorMatch(t *testing.T) {
	var path string
	var err error
	var expected string

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement
	var vendor string

	ctx := context.Background()
	store, err := getMockStore()
	th.Assert(t, err == nil, err)

	// cerner
	expected = "Cerner Corporation"

	path = filepath.Join("testdata", "cerner_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// epic
	expected = "" // the capability statement is missing the publisher

	path = filepath.Join("testdata", "epic_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// allscripts
	expected = "Allscripts" // the capability statement is missing the publisher

	path = filepath.Join("testdata", "allscripts_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// meditech
	expected = "Medical Information Technology, Inc. (MEDITECH)" // the capability statement is missing the publisher

	path = filepath.Join("testdata", "meditech_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))
}

func getMockStore() (endpointmanager.HealthITProductStore, error) {
	hitp, err := mock.NewStore()
	if err != nil {
		return nil, err
	}

	hitp.GetHealthITProductDevelopersFn = func(ctx context.Context) ([]string, error) {
		devList := []string{
			"Epic Systems Corporation",
			"Cerner Corporation",
			"Cerner Health Services, Inc.",
			"Medical Information Technology, Inc. (MEDITECH)",
			"Allscripts",
		}

		return devList, nil
	}

	return hitp, nil
}
