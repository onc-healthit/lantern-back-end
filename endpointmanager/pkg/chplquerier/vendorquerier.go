package chplquerier

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

var chplAPIVendorListPath string = "/developers"

type chplVendorList struct {
	Developers []chplVendor `json:"developers"`
}

type chplVendor struct {
	DeveloperID   int    `json:"developerId"`
	DeveloperCode string `json:"developerCode"`
	Name          string `json:"name"`
	Website       string `json:"website"`
	Address       struct {
		AddressID int         `json:"addressId"`
		Line1     interface{} `json:"line1"`
		Line2     interface{} `json:"line2"`
		City      interface{} `json:"city"`
		State     interface{} `json:"state"`
		Zipcode   interface{} `json:"zipcode"`
		Country   interface{} `json:"country"`
	} `json:"address"`
	LastModifiedDate string `json:"lastModifiedDate"`
	Status           struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	} `json:"status"`
}

func GetVendors(ctx context.Context, store *postgresql.Store, cli *http.Client) error {
	log.Debug("requesting vendors from CHPL")
	prodJSON, err := getVendorJSON(ctx, cli)
	if err != nil {
		return errors.Wrap(err, "getting vendor JSON failed")
	}
	log.Debug("done requesting vendors from CHPL")

	log.Debug("converting chpl json into vendor objects")
	prodList, err := convertVendorJSONToObj(ctx, prodJSON)
	if err != nil {
		return errors.Wrap(err, "converting vendor JSON into a 'chplCertifiedProductList' object failed")
	}
	log.Debug("done converting chpl json into evndor objects")

	log.Debug("persisting vendors")
	err = persistVendors(ctx, store, prodList)
	log.Debug("done persisting vendors")
	return errors.Wrap(err, "persisting the list of retrieved health IT products failed")
}

// makes the request to CHPL and returns the byte string
func getVendorJSON(ctx context.Context, client *http.Client) ([]byte, error) {
	chplURL, err := makeCHPLVendorURL()
	if err != nil {
		return nil, errors.Wrap(err, "creating the URL to query CHPL failed")
	}

	println(chplURL.String())

	// request vendor list
	req, err := http.NewRequest("GET", chplURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating http request failed")
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "making the GET request to the CHPL server failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("CHPL certified products request responded with status: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading the CHPL response body failed")
	}

	println(string(body))

	return body, nil
}

// creates the URL used to query CHPL including the data fields
func makeCHPLVendorURL() (*url.URL, error) {

	chplURL, err := makeCHPLURL(chplAPIVendorListPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating the URL to query CHPL failed")
	}

	return chplURL, nil
}

// takes the json byte string and converts it into the associated JSON model
func convertVendorJSONToObj(ctx context.Context, vendorJSON []byte) (*chplVendorList, error) {
	var vendorList chplVendorList

	// don't unmarshal the JSON if the context has ended
	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "Unable to convert product JSON to objects - context ended")
	default:
		// ok
	}

	err := json.Unmarshal(vendorJSON, &vendorList)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling the JSON into a chplCertifiedProductList object failed.")
	}

	return &vendorList, nil
}

// takes the JSON model and converts it into an endpointmanager.HealthITProduct
func parseVendor(vendor *chplVendor) (*endpointmanager.Vendor, error) {
	var loc endpointmanager.Location

	loc.Address1 = interfaceToString(vendor.Address.Line1)
	loc.Address2 = interfaceToString(vendor.Address.Line2)
	loc.City = interfaceToString(vendor.Address.City)
	loc.State = interfaceToString(vendor.Address.State)
	loc.ZipCode = interfaceToString(vendor.Address.Zipcode)

	dbVendor := endpointmanager.Vendor{
		Name:               vendor.Name,
		DeveloperCode:      vendor.DeveloperCode,
		URL:                vendor.Website,
		Location:           &loc,
		Status:             vendor.Status.Status,
		LastModifiedInCHPL: stringToDate(vendor.LastModifiedDate),
		CHPLID:             vendor.DeveloperID,
	}

	return &dbVendor, nil
}

// persists the vendors parsed from CHPL.
func persistVendors(ctx context.Context, store *postgresql.Store, vendorList *chplVendorList) error {
	for i, vendor := range vendorList.Developers {

		select {
		case <-ctx.Done():
			return errors.Wrapf(ctx.Err(), "persisted %d out of %d vendors before context ended", i, len(vendorList.Developers))
		default:
			// ok
		}

		if i%100 == 0 {
			log.Infof("persisting chpl vendor %d/%d", i, len(vendorList.Developers))
		}

		err := persistVendor(ctx, store, &vendor)
		if err != nil {
			log.Warn(err)
			continue
		}
	}
	return nil
}

func persistVendor(ctx context.Context,
	store *postgresql.Store,
	vendor *chplVendor) error {

	newDbVendor, err := parseVendor(vendor)
	if err != nil {
		return err
	}
	existingDbVendor, err := store.GetVendorUsingName(ctx, vendor.Name)

	if err == sql.ErrNoRows { // need to add new entry
		err = store.AddVendor(ctx, newDbVendor)
		if err != nil {
			return errors.Wrap(err, "adding vendor to store failed")
		}
	} else if err != nil {
		return errors.Wrap(err, "getting vendor from store failed")
	} else {
		newDbVendor.ID = existingDbVendor.ID
		err = store.UpdateVendor(ctx, newDbVendor)
		if err != nil {
			return errors.Wrap(err, "updating vendor to store failed")
		}
	}
	return nil
}

func interfaceToString(interStr interface{}) string {
	str, ok := interStr.(string)
	if ok {
		return str
	}

	return ""
}

func stringToDate(dateStr string) time.Time {
	dateInt, err := strconv.ParseInt(dateStr, 10, 64)

	if err == nil {
		return time.Unix(dateInt/1000, 0).UTC()
	}

	return time.Unix(0, 0)
}
