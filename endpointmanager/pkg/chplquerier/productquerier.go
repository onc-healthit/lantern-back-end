package chplquerier

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/google/go-cmp/cmp"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

var chplAPICertProdListPath string = "/search/v3"

type chplEndpointListProductInfo struct {
	ListSourceURL    string                 `json:"listSourceURL"`
	SoftwareProducts []chplCertifiedProduct `json:"softwareProducts"`
}

type chplCertifiedProductList struct {
	Results     []chplCertifiedProduct `json:"results"`
	RecordCount int                    `json:"recordCount"`
}

type details struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type criteriaMet struct {
	Id     int    `json:"id"`
	Number string `json:"number"`
	Title  string `json:"title"`
}

type apiDocumentation struct {
	Criterion criteriaMet `json:"criterion"`
	Value     string      `json:"value"`
}

type chplCertifiedProduct struct {
	ID                  int                `json:"id"`
	ChplProductNumber   string             `json:"chplProductNumber"`
	Edition             details            `json:"edition"`
	PracticeType        details            `json:"practiceType"`
	Developer           details            `json:"developer"`
	Product             details            `json:"product"`
	Version             details            `json:"version"`
	CertificationDate   string             `json:"certificationDate"`
	CertificationStatus details            `json:"certificationStatus"`
	CriteriaMet         []criteriaMet      `json:"criteriaMet"`
	APIDocumentation    []apiDocumentation `json:"apiDocumentation"`
	ACB                 string             `json:"acb"`
}

// GetCHPLProducts queries CHPL for its HealthIT products using 'cli' and stores the products in 'store'
// within the given context 'ctx'.
func GetCHPLProducts(ctx context.Context, store *postgresql.Store, cli *http.Client, userAgent string) error {
	pageSize := 100
	pageNumber := 0
	persistedProducts := 0
	for {
		log.Debug("requesting page of products from CHPL")
		prodJSON, err := getProductJSON(ctx, cli, userAgent, pageSize, pageNumber)
		if err != nil {
			return nil
		}
		log.Debug("done requesting page of products from CHPL")

		log.Debug("converting chpl json into product objects")
		prodList, err := convertProductJSONToObj(ctx, prodJSON)
		if err != nil {
			return errors.Wrap(err, "converting health IT product JSON into a 'chplCertifiedProductList' object failed")
		}
		log.Debug("done converting chpl json into product objects")
		if persistedProducts >= prodList.RecordCount {
			log.Debug("done persisting all chpl products")
			break
		}
		log.Debug("persisting chpl products")
		err = persistProducts(ctx, store, prodList)
		if err != nil {
			return errors.Wrap(err, "persisting the list of retrieved health IT products failed")
		}
		pageNumber = pageNumber + 1
		persistedProducts = persistedProducts + len(prodList.Results)
		log.Debug("done persisting chpl products")
		if persistedProducts%100 == 0 {
			log.Infof("have persisted chpl products %d/%d", persistedProducts, prodList.RecordCount)
		}
	}
	return nil
}

// GetCHPLEndpointListProducts grabs software information from the CHPLProductsInfo.json file and stores the products in 'store'
// within the given context 'ctx'.
func GetCHPLEndpointListProducts(ctx context.Context, store *postgresql.Store) error {

	var CHPLEndpointListProducts []chplEndpointListProductInfo

	log.Info("Getting chpl product information from CHPLProductsInfo.json file")
	// Get CHPL Endpoint list stored in Lantern resources folder
	CHPLFile, err := os.ReadFile("/etc/lantern/resources/CHPLProductsInfo.json")
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Converting product information into list of chplCertifiedProducts")
	err = json.Unmarshal(CHPLFile, &CHPLEndpointListProducts)
	if err != nil {
		log.Fatal(err)
	}

	for _, listSourceEntry := range CHPLEndpointListProducts {
		var prodList []chplCertifiedProduct
		var CHPLProductList chplCertifiedProductList

		prodList = listSourceEntry.SoftwareProducts

		CHPLProductList.Results = prodList
		CHPLProductList.RecordCount = len(prodList)

		log.Debug("persisting chpl products")
		err = persistProducts(ctx, store, &CHPLProductList)
		if err != nil {
			return errors.Wrap(err, "persisting the list of retrieved health IT products failed")
		}
	}

	log.Debug("done persisting chpl products")
	return nil
}

// makes the request to CHPL and returns the byte string
func getProductJSON(ctx context.Context, client *http.Client, userAgent string, pageSize int, pageNumber int) ([]byte, error) {
	chplURL, err := makeCHPLProductURL(pageSize, pageNumber)
	if err != nil {
		return nil, errors.Wrap(err, "error creating CHPL product URL")
	}

	// None of the returned errors should break the system, so print a warning instead
	jsonBody, err := getJSON(ctx, client, chplURL, userAgent)
	if err != nil {
		log.Warnf("Got error:\n%s\n\nfrom URL: %s", err.Error(), chplURL.String())
	}
	return jsonBody, nil
}

func makeCHPLProductURL(pageSize int, pageNumber int) (*url.URL, error) {
	queryArgs := make(map[string]string)

	chplURL, err := makeCHPLURL(chplAPICertProdListPath, queryArgs, pageSize, pageNumber)
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

	err := json.Unmarshal(prodJSON, &prodList)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling the JSON into a chplCertifiedProductList object failed.")
	}

	return &prodList, nil
}

// takes the JSON model and converts it into an endpointmanager.HealthITProduct
func parseHITProd(ctx context.Context, prod *chplCertifiedProduct, store *postgresql.Store) (*endpointmanager.HealthITProduct, error) {
	id, err := getProductVendorID(ctx, prod, store)
	if err != nil {
		return nil, errors.Wrap(err, "getting the product's vendor id failed")
	}

	var criteriaMetArr []int
	for _, criteriaEntry := range prod.CriteriaMet {
		criteriaMetArr = append(criteriaMetArr, criteriaEntry.Id)
	}

	dbProd := endpointmanager.HealthITProduct{
		Name:                  strings.TrimSpace(prod.Product.Name),
		Version:               strings.TrimSpace(prod.Version.Name),
		VendorID:              id,
		CertificationStatus:   strings.TrimSpace(prod.CertificationStatus.Name),
		CertificationEdition:  strings.TrimSpace(prod.Edition.Name),
		CHPLID:                prod.ChplProductNumber,
		CertificationCriteria: criteriaMetArr,
		PracticeType:          strings.TrimSpace(prod.PracticeType.Name),
		ACB:                   strings.TrimSpace(prod.ACB),
	}

	certificationDateTime, err := time.Parse("2006-01-02", prod.CertificationDate)
	if err != nil {
		return nil, errors.Wrap(err, "converting certification date to time failed")
	}
	dbProd.CertificationDate = certificationDateTime.UTC()

	apiDocURL, err := getAPIURL(prod.APIDocumentation)
	if err != nil {
		return nil, errors.Wrap(err, "retreiving the API URL from the health IT product API documentation list failed")
	}
	dbProd.APIURL = apiDocURL

	return &dbProd, nil
}

// returns 0 if no match found.
func getProductVendorID(ctx context.Context, prod *chplCertifiedProduct, store *postgresql.Store) (int, error) {
	vendor, err := store.GetVendorUsingName(ctx, prod.Developer.Name)
	if err == sql.ErrNoRows {
		log.Warnf("no vendor match for product %s with vendor %s", prod.Product.Name, prod.Developer.Name)
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrapf(err, "getting vendor for product %s %s using vendor name %s", prod.Product.Name, prod.Version.Name, prod.Developer.Name)
	}

	return vendor.ID, nil
}

// parses 'apiDocArr' to extract the associated URL. Returns only the first URL. There may be many URLs but observationally,
// all listed URLs are the same.
func getAPIURL(apiDocArr []apiDocumentation) (string, error) {
	if len(apiDocArr) == 0 {
		return "", nil
	}
	apiURL := apiDocArr[0].Value

	// check that it's a valid URL
	_, err := url.ParseRequestURI(apiURL)
	if err != nil {
		return "", errors.Wrap(err, "the URL in the health IT product API documentation string is not valid")
	}
	return apiURL, nil
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

	newDbProd, err := parseHITProd(ctx, prod, store)
	if err != nil {
		return err
	}
	existingDbProd, err := store.GetHealthITProductUsingNameAndVersion(ctx, newDbProd.Name, newDbProd.Version)

	newElement := true
	if err == sql.ErrNoRows { // need to add new entry
		err = store.AddHealthITProduct(ctx, newDbProd)
		if err != nil {
			return errors.Wrap(err, "adding health IT product to store failed")
		}
	} else if err != nil {
		return errors.Wrap(err, "getting health IT product from store failed")
	} else {
		newElement = false
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

	if newElement {
		for _, critID := range newDbProd.CertificationCriteria {
			err = linkProductToCriteria(ctx, store, critID, newDbProd.ID)
			if err != nil {
				return err
			}
		}
	} else {
		for _, critID := range existingDbProd.CertificationCriteria {
			err = store.DeleteLinksByProduct(ctx, existingDbProd.ID)
			if err != nil {
				return errors.Wrap(err, "removing old product from links store failed")
			}
			err = linkProductToCriteria(ctx, store, critID, existingDbProd.ID)
			if err != nil {
				return err
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
// else if the new product has a populated acb, update.
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

	// If the new product has a new ACB field, update it unless the field is not populated
	if existingDbProd.ACB != newDbProd.ACB {

		if newDbProd.ACB == "" {
			return false, nil
		} else {
			return true, nil
		}
	}

	// Compare certification editions
	// Assumes certification editions are years, which is the case as of 11/20/19.
	existingCertEdition, err := strconv.Atoi(existingDbProd.CertificationEdition)
	if err != nil {
		log.Warnf("Edition parse failed (existing): product=%s chplid=%s edition=%q error=%v",
			existingDbProd.Name,
			existingDbProd.CHPLID,
			existingDbProd.CertificationEdition,
			err,
		)
		return false, errors.Wrap(err, "unable to make certification edition into an integer - expect certification edition to be a year")
	}
	newCertEdition, err := strconv.Atoi(newDbProd.CertificationEdition)
	if err != nil {
		log.Warnf("Edition parse failed (new): product=%s chplid=%s edition=%q error=%v",
			newDbProd.Name,
			newDbProd.CHPLID,
			newDbProd.CertificationEdition,
			err,
		)
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

	existingCriteriaLength := len(existingDbProd.CertificationCriteria)
	newCriteriaLength := len(newDbProd.CertificationCriteria)

	// If new criteria list is bigger than old update it, if it is smaller do not update it
	if newCriteriaLength > existingCriteriaLength {
		return true, nil
	} else if newCriteriaLength < existingCriteriaLength {
		return false, nil
	}

	// Do not update or throw error if the practice types are not the same
	if existingDbProd.PracticeType != newDbProd.PracticeType {
		return false, nil
	}

	// If the criteria lists are the same length but they are not equal, throw an error
	if !certificationCriteriaMatch(existingDbProd.CertificationCriteria, newDbProd.CertificationCriteria) {
		sortInts := func(in []int) []int {
			out := append([]int(nil), in...)
			sort.Ints(out)
			return out
		}

		log.Warnf(`criteria differ with equal length (debug dump)
		product=%s
		chplid=%s
		old_raw=%v
		new_raw=%v
		old_criteria=%v
		new_criteria=%v
		old_len=%d
		new_len=%d`,
			newDbProd.Name,
			newDbProd.CHPLID,
			existingDbProd.CertificationCriteria,
			newDbProd.CertificationCriteria,
			sortInts(existingDbProd.CertificationCriteria),
			sortInts(newDbProd.CertificationCriteria),
			len(existingDbProd.CertificationCriteria),
			len(newDbProd.CertificationCriteria),
		)

		return false, fmt.Errorf("HealthITProducts certification criteria have the same length but are not equal; not performing update: %s:%s to %s:%s", existingDbProd.Name, existingDbProd.CHPLID, newDbProd.Name, newDbProd.CHPLID)
	}

	// If the new product has a different vendor ID, update it
	if existingDbProd.VendorID != newDbProd.VendorID {
		return true, nil
	}

	// If the new product has a different certification status, update it
	if existingDbProd.CertificationStatus != newDbProd.CertificationStatus {
		return true, nil
	}

	// If the new product has a different API url, update it
	if existingDbProd.APIURL != newDbProd.APIURL {
		return true, nil
	}

	return false, fmt.Errorf("Unknown difference between HealthITProducts; not performing update: %v to %v", existingDbProd, newDbProd)

}

// linkProductToCriteria checks whether the product and certification have been linked before, and if not
// links them
func linkProductToCriteria(ctx context.Context,
	store *postgresql.Store,
	critID int,
	prodID int) error {
	_, _, _, err := store.GetProductCriteriaLink(ctx, critID, prodID)
	// Only care about whether it's not there, if it's already saved it shouldn't need
	// to be updated
	if err == sql.ErrNoRows {
		certCrit, err := store.GetCriteriaByCertificationID(ctx, critID)
		if err != nil {
			return errors.Wrap(err, "Error linking criteria to FHIR endpoint")
		}
		err = store.LinkProductToCriteria(ctx, critID, prodID, certCrit.CertificationNumber)
		if err != nil {
			return errors.Wrap(err, "Error linking criteria to FHIR endpoint")
		}
	} else if err != nil {
		return err
	}
	return nil
}

// certificationCriteriaMatch checks if the two certification criteria lists have the same contents regardless of order.
func certificationCriteriaMatch(l1 []int, l2 []int) bool {
	// This Transformer sorts a []int.
	trans := cmp.Transformer("Sort", func(in []int) []int {
		out := append([]int(nil), in...) // Copy input to avoid mutating it
		sort.Ints(out)
		return out
	})
	return cmp.Equal(l1, l2, trans)
}
