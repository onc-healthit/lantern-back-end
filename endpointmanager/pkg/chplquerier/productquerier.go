package chplquerier

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

var chplAPICertProdListPath string = "/collections/certified_products"
var delimiter1 string = "☺"
var delimiter2 string = "☹"

var fields [11]string = [11]string{
	"id",
	"edition",
	"developer",
	"product",
	"version",
	"chplProductNumber",
	"certificationStatus",
	"criteriaMet",
	"apiDocumentation",
	"certificationDate",
	"practiceType"}

type chplCertifiedProductList struct {
	Results []chplCertifiedProduct `json:"results"`
}

type chplCertifiedProduct struct {
	ID                  int    `json:"id"`
	ChplProductNumber   string `json:"chplProductNumber"`
	Edition             string `json:"edition"`
	PracticeType        string `json:"practiceType"`
	Developer           string `json:"developer"`
	Product             string `json:"product"`
	Version             string `json:"version"`
	CertificationDate   int64  `json:"certificationDate"`
	CertificationStatus string `json:"certificationStatus"`
	CriteriaMet         string `json:"criteriaMet"`
	APIDocumentation    string `json:"apiDocumentation"`
}

// GetCHPLProducts queries CHPL for its HealthIT products using 'cli' and stores the products in 'store'
// within the given context 'ctx'.
func GetCHPLProducts(ctx context.Context, store *postgresql.Store, cli *http.Client) error {
	log.Debug("requesting products from CHPL")
	prodJSON, err := getProductJSON(ctx, cli)
	if err != nil {
		return errors.Wrap(err, "getting health IT product JSON failed")
	}
	log.Debug("done requesting products from CHPL")

	log.Debug("converting chpl json into product objects")
	prodList, err := convertProductJSONToObj(ctx, prodJSON)
	if err != nil {
		return errors.Wrap(err, "converting health IT product JSON into a 'chplCertifiedProductList' object failed")
	}
	log.Debug("done converting chpl json into product objects")

	log.Debug("persisting chpl products")
	err = persistProducts(ctx, store, prodList)
	log.Debug("done persisting chpl products")
	return errors.Wrap(err, "persisting the list of retrieved health IT products failed")
}

// makes the request to CHPL and returns the byte string
func getProductJSON(ctx context.Context, client *http.Client) ([]byte, error) {
	chplURL, err := makeCHPLProductURL()
	if err != nil {
		return nil, errors.Wrap(err, "creating the URL to query CHPL failed")
	}

	// request ceritified products list
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

	return body, nil
}

// creates the URL used to query CHPL including the data fields
func makeCHPLProductURL() (*url.URL, error) {
	queryArgs := make(map[string]string)
	fieldStr := strings.Join(fields[:], ",")
	queryArgs["fields"] = fieldStr

	chplURL, err := makeCHPLURL(chplAPICertProdListPath, queryArgs)
	if err != nil {
		return nil, errors.Wrap(err, "creating the URL to query CHPL failed")
	}

	return chplURL, nil
}

// takes the json byte string and converts it into the associated JSON model
func convertProductJSONToObj(ctx context.Context, prodJSON []byte) (*chplCertifiedProductList, error) {
	var prodList chplCertifiedProductList

	// don't unmarshal the JSON if the context has ended
	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "Unable to convert product JSON to objects - context ended")
	default:
		// ok
	}

	fmt.Println("### PRINTING JSON ###\n")
	fmt.Println("%v\n", prodJSON)
	fmt.Println("### DONE PRINTING JSON ###\n")
	err := json.Unmarshal(prodJSON, &prodList)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling the JSON into a chplCertifiedProductList object failed.")
	}

	return &prodList, nil
}

// takes the JSON model and converts it into an endpointmanager.HealthITProduct
func parseHITProd(prod *chplCertifiedProduct) (*endpointmanager.HealthITProduct, error) {
	dbProd := endpointmanager.HealthITProduct{
		Name:                  prod.Product,
		Version:               prod.Version,
		Developer:             prod.Developer,
		CertificationStatus:   prod.CertificationStatus,
		CertificationDate:     time.Unix(prod.CertificationDate/1000, 0).UTC(),
		CertificationEdition:  prod.Edition,
		CHPLID:                prod.ChplProductNumber,
		CertificationCriteria: strings.Split(prod.CriteriaMet, delimiter1),
	}

	apiURL, err := getAPIURL(prod.APIDocumentation)
	if err != nil {
		return nil, errors.Wrap(err, "retreiving the API URL from the health IT product API documentation list failed")
	}
	dbProd.APIURL = apiURL

	return &dbProd, nil
}

// parses 'apiDocStr' to extract the associated URL. Returns only the first URL. There may be many URLs but observationally,
// all listed URLs are the same.
// assumes that criteria/url chunks are delimited by delimiter1 and that criteria and url are separated by delimiter2.
func getAPIURL(apiDocStr string) (string, error) {
	if len(apiDocStr) == 0 {
		return "", nil
	}

	apiDocStrs := strings.Split(apiDocStr, delimiter1)
	apiCritAndURL := strings.Split(apiDocStrs[0], delimiter2)
	if len(apiCritAndURL) == 2 {
		apiURL := apiCritAndURL[1]
		// check that it's a valid URL
		_, err := url.ParseRequestURI(apiURL)
		if err != nil {
			return "", errors.Wrap(err, "the URL in the health IT product API documentation string is not valid")
		}
		return apiURL, nil
	}

	return "", errors.New("unexpected format for api doc string")
}

// persists the products parsed from CHPL. Of note, CHPL includes many entries for a single product. The entry
// associated with the most recent certifition edition, most recent certification date, or most criteria is the
// one that is stored.
func persistProducts(ctx context.Context, store *postgresql.Store, prodList *chplCertifiedProductList) error {
	for i, prod := range prodList.Results {

		select {
		case <-ctx.Done():
			return errors.Wrapf(ctx.Err(), "persisted %d out of %d products before context ended", i, len(prodList.Results))
		default:
			// ok
		}

		if i%100 == 0 {
			log.Infof("persisting chpl product %d/%d", i, len(prodList.Results))
		}

		err := persistProduct(ctx, store, &prod)
		if err != nil {
			log.Warn(err)
			continue
		}
	}
	return nil
}

// adds a product to the store if that product's name/version don't exist already. If the name/version do
// exist, determine if it makes sense to update the product (certified to more recent edition, certified at a
// later date, has more certification criteria), or not.
func persistProduct(ctx context.Context,
	store *postgresql.Store,
	prod *chplCertifiedProduct) error {

	newDbProd, err := parseHITProd(prod)
	if err != nil {
		return err
	}
	existingDbProd, err := store.GetHealthITProductUsingNameAndVersion(ctx, prod.Product, prod.Version)

	if err == sql.ErrNoRows { // need to add new entry
		err = store.AddHealthITProduct(ctx, newDbProd)
		if err != nil {
			return errors.Wrap(err, "adding health IT product to store failed")
		}
	} else if err != nil {
		return errors.Wrap(err, "getting health IT product from store failed")
	} else {
		needsUpdate, err := prodNeedsUpdate(existingDbProd, newDbProd)
		if err != nil {
			return errors.Wrap(err, "determining if a health IT product needs updating within the store failed")
		}

		if needsUpdate {
			err = existingDbProd.Update(newDbProd)
			if err != nil {
				return errors.Wrap(err, "updating health IT product object failed")
			}
			err = store.UpdateHealthITProduct(ctx, existingDbProd)
			if err != nil {
				return errors.Wrap(err, "updating health IT product to store failed")
			}
		}
	}
	return nil
}

// determines if a product needs to be udpated.
//
// if the two products are equal, do not update.
// else if the new product has a more recent certification edition than the exisitng product, update.
// else if the new product has a more recent certification date than the exisitng product, update.
// else if the new product has more certification criteria than the existing product, update.
//
// throws errors if
// - the certification edition is not a year
// - the certification criteria list is the same length but not equal
// - the two products are not equal but their differences don't fall into the categories noted above.
func prodNeedsUpdate(existingDbProd *endpointmanager.HealthITProduct, newDbProd *endpointmanager.HealthITProduct) (bool, error) {
	// check if the two are equal.
	if existingDbProd.Equal(newDbProd) {
		return false, nil
	}

	// begin by comparing certification editions.
	// Assumes certification editions are years, which is the case as of 11/20/19.
	existingCertEdition, err := strconv.Atoi(existingDbProd.CertificationEdition)
	if err != nil {
		return false, errors.Wrap(err, "unable to make certification edition into an integer - expect certification edition to be a year")
	}
	newCertEdition, err := strconv.Atoi(newDbProd.CertificationEdition)
	if err != nil {
		return false, errors.Wrap(err, "unable to make certification edition into an integer - expect certification edition to be a year")
	}

	// if new prod has more recent cert edition, should update.
	if newCertEdition > existingCertEdition {
		return true, nil
	} else if newCertEdition < existingCertEdition {
		return false, nil
	}

	// cert editions are the same. if new prod has more recent cert date, should update.
	if existingDbProd.CertificationDate.Before(newDbProd.CertificationDate) {
		return true, nil
	} else if existingDbProd.CertificationDate.After(newDbProd.CertificationDate) {
		return false, nil
	}

	// cert dates are the same. unknown update precedence. throw error and don't perform update.
	return false, fmt.Errorf("HealthITProducts certification edition and date are equal; unknown precendence for updates; not performing update: %s:%s to %s:%s", existingDbProd.Name, existingDbProd.CHPLID, newDbProd.Name, newDbProd.CHPLID)
}
