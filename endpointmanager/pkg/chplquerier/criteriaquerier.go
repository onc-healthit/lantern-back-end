package chplquerier

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

var chplAPICertCriteriaPath string = "/data/certification-criteria"

// var delimiter1 string = "☺"
// var delimiter2 string = "☹"

// var fields [11]string = [11]string{
// 	"id",
// 	"edition",
// 	"developer",
// 	"product",
// 	"version",
// 	"chplProductNumber",
// 	"certificationStatus",
// 	"criteriaMet",
// 	"apiDocumentation",
// 	"certificationDate",
// 	"practiceType"}

type chplCertifiedCriteriaList struct {
	Results []chplCertCriteria `json:"criteria"`
}

type chplCertCriteria struct {
	ID                     int    `json:"id"`
	Number                 string `json:"number"`
	Title                  string `json:"title"`
	CertificationEditionID int    `json:"certificationEditionId"`
	CertificationEdition   string `json:"certificationEdition"`
	Description            string `json:"description"`
	Removed                bool   `json:"removed"`
}

// @TODO Change all of the comments
func GetCHPLCriteria(ctx context.Context, store *postgresql.Store, cli *http.Client) error {
	log.Debug("requesting certification criteria from CHPL")
	critJSON, err := getCriteriaJSON(ctx, cli)
	if err != nil {
		return err
	}
	log.Debug("done requesting certification criteria from CHPL")

	log.Debug("converting chpl json into certification criteria objects")
	critList, err := convertCriteriaJSONToObj(ctx, critJSON)
	if err != nil {
		return errors.Wrap(err, "converting health IT product JSON into a 'chplCriteriaList' object failed")
	}
	log.Debug("done converting chpl json into product objects")

	log.Debug("persisting chpl products")
	err = persistCriterias(ctx, store, critList)
	log.Debug("done persisting chpl products")
	return errors.Wrap(err, "persisting the list of retrieved health IT products failed")
}

// makes the request to CHPL and returns the byte string
func getCriteriaJSON(ctx context.Context, client *http.Client) ([]byte, error) {
	chplURL, err := makeCHPLCriteriaURL()
	if err != nil {
		return nil, errors.Wrap(err, "error creating CHPL product URL")
	}

	// None of the returned errors should break the system, so print a warning instead
	jsonBody, err := getJSON(ctx, client, chplURL)
	if err != nil {
		log.Warnf("Got error:\n%s\n\nfrom URL: %s", err.Error(), chplURL.String())
	}
	return jsonBody, nil
}

func makeCHPLCriteriaURL() (*url.URL, error) {
	queryArgs := make(map[string]string)
	// fieldStr := strings.Join(fields[:], ",")
	// queryArgs["fields"] = fieldStr

	chplURL, err := makeCHPLURL(chplAPICertCriteriaPath, queryArgs)
	if err != nil {
		return nil, errors.Wrap(err, "creating the URL to query CHPL failed")
	}

	return chplURL, nil
}

// takes the json byte string and converts it into the associated JSON model
func convertCriteriaJSONToObj(ctx context.Context, critJSON []byte) (*chplCertifiedCriteriaList, error) {
	var critList chplCertifiedCriteriaList

	// don't unmarshal the JSON if the context has ended
	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "Unable to convert certified criteria JSON to objects - context ended")
	default:
		// ok
	}

	fmt.Printf("%s", string(critJSON))

	err := json.Unmarshal(critJSON, &critList)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling the JSON into a chplCertifiedCriteriaList object failed.")
	}

	return &critList, nil
}

// takes the JSON model and converts it into an endpointmanager.HealthITProduct
func parseHITCriteria(ctx context.Context, criteria *chplCertCriteria, store *postgresql.Store) (*endpointmanager.CertificationCriteria, error) {

	dbCrit := endpointmanager.CertificationCriteria{
		CertificationID:        criteria.ID,
		CertificationNumber:    criteria.Number,
		Title:                  criteria.Title,
		CertificationEditionID: criteria.CertificationEditionID,
		CertificationEdition:   criteria.CertificationEdition,
		Description:            criteria.Description,
		Removed:                criteria.Removed,
	}

	return &dbCrit, nil
}

// persists the products parsed from CHPL. Of note, CHPL includes many entries for a single product. The entry
// associated with the most recent certifition edition, most recent certification date, or most criteria is the
// one that is stored.
func persistCriterias(ctx context.Context, store *postgresql.Store, critList *chplCertifiedCriteriaList) error {
	for i, criteria := range critList.Results {

		select {
		case <-ctx.Done():
			return errors.Wrapf(ctx.Err(), "persisted %d out of %d certification criteria before context ended", i, len(critList.Results))
		default:
			// ok
		}

		if i%100 == 0 {
			log.Infof("persisting chpl certification criteria %d/%d", i, len(critList.Results))
		}

		err := persistCriteria(ctx, store, &criteria)
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
func persistCriteria(ctx context.Context,
	store *postgresql.Store,
	criteria *chplCertCriteria) error {

	newDbCrit, err := parseHITCriteria(ctx, criteria, store)
	if err != nil {
		return err
	}
	existingDbCrit, err := store.GetCriteriaByCertificationID(ctx, criteria.ID)

	if err == sql.ErrNoRows { // need to add new entry
		err = store.AddCriteria(ctx, newDbCrit)
		if err != nil {
			return errors.Wrap(err, "adding certification criteria to store failed")
		}
	} else if err != nil {
		return errors.Wrap(err, "getting certification criteria from store failed")
	} else {
		// needsUpdate, err := prodNeedsUpdate(existingDbProd, newDbProd)
		// if err != nil {
		// 	return errors.Wrap(err, "determining if a health IT product needs updating within the store failed")
		// }

		// if needsUpdate {
		err = existingDbCrit.Update(newDbCrit)
		if err != nil {
			return errors.Wrap(err, "updating certification criteria object failed")
		}
		err = store.UpdateCriteria(ctx, existingDbCrit)
		if err != nil {
			return errors.Wrap(err, "updating certification criteria to store failed")
		}
		// }
	}
	return nil
}
