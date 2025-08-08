package chplmapper

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/pkg/errors"
)

var fluffWords = []string{
	"inc.",
	"inc",
	"llc",
	"corp.",
	"corp",
	"corporation",
	"lmt",
	"lmt.",
	"limited",
	"corporation.",
}

type details struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ChplEndpointListProductInfo struct {
	ListSourceURL    string                 `json:"listSourceURL"`
	SoftwareProducts []ChplCertifiedProduct `json:"softwareProducts"`
}

type ChplCertifiedProduct struct {
	ChplProductNumber string  `json:"chplProductNumber"`
	Developer         details `json:"developer"`
}

type ChplMapResults struct {
	ChplProductIDs []string
	ChplDeveloper  []string
}

// MatchEndpointToVendor creates the database association between the endpoint and the vendor,
// and the endpoint and the healht IT product.
func MatchEndpointToVendor(ctx context.Context, ep *endpointmanager.FHIREndpointInfo, store *postgresql.Store, developerName string) error {

	if len(developerName) > 0 {
		// No errors thrown means a vendor with developer name was found and can be set on ep
		vendorMatch, err := store.GetVendorUsingName(ctx, developerName)
		if err == sql.ErrNoRows {
			log.Infof("No vendor found matching the CHPL endpoint list developer name. Ensure the vendor table is not empty.")
			return nil
		} else if err != nil {
			return errors.Wrap(err, "error matching the CHPL endpoint list developer name to a vendor for endpoint")
		}
		ep.VendorID = vendorMatch.ID
		return nil
	}

	if ep.CapabilityStatement == nil {
		return nil
	}

	vendorID, err := getVendorMatch(ctx, ep.CapabilityStatement, store)
	if err != nil {
		return errors.Wrap(err, "error matching the capability statement to a vendor for endpoint")
	}

	ep.VendorID = vendorID

	return nil
}

// MatchEndpointToProduct creates the database association between the endpoint and the HealthITProduct,
func MatchEndpointToProduct(ctx context.Context, ep *endpointmanager.FHIREndpointInfo, store *postgresql.Store, matchFile string, productIds []string) error {

	softwareName := ""
	softwareVersion := ""
	chplIDArr := []string{}

	if ep.CapabilityStatement != nil {
		chplProductNameVersion, err := openProductLinksFile(matchFile)
		if err != nil {
			return errors.Wrap(err, "error matching the capability statement to a CHPL product")
		}

		softwareName, err = ep.CapabilityStatement.GetSoftwareName()
		if err != nil {
			return errors.Wrap(err, "error matching the capability statement to a CHPL product")
		}
		softwareVersion, err = ep.CapabilityStatement.GetSoftwareVersion()
		if err != nil {
			return errors.Wrap(err, "error matching the capability statement to a CHPL product")
		}

		chplIDMatchFile := chplProductNameVersion[softwareName][softwareVersion]

		if len(chplIDMatchFile) != 0 {
			chplIDArr = append(chplIDArr, chplIDMatchFile)
		}
	}

	if len(productIds) > 0 {
		chplIDArr = append(chplIDArr, productIds...)
	}

	var healthITProductsArr []*endpointmanager.HealthITProduct
	var err error
	if len(softwareName) != 0 {
		healthITProductsArr, err = store.GetActiveHealthITProductsUsingName(ctx, softwareName)
		if err != nil {
			return err
		}
	}

	for _, healthITProduct := range healthITProductsArr {
		if len(softwareVersion) == 0 {
			if !helpers.StringArrayContains(chplIDArr, healthITProduct.CHPLID) {
				chplIDArr = append(chplIDArr, healthITProduct.CHPLID)
			}
		} else {
			if strings.EqualFold(healthITProduct.Version, softwareVersion) {
				if !helpers.StringArrayContains(chplIDArr, healthITProduct.CHPLID) {
					chplIDArr = append(chplIDArr, healthITProduct.CHPLID)
				}
			}
		}
	}

	log.Info("chplIDArr: ", chplIDArr, "\n")

	for _, chplID := range chplIDArr {
		healthITProductID, err := store.GetHealthITProductIDByCHPLID(ctx, chplID)
		// No errors thrown means a healthit product with CHPLID was found and can be set on ep
		if err == nil {
			healthITMapID, err := store.AddHealthITProductMap(ctx, ep.HealthITProductID, healthITProductID)
			if err != nil {
				return err
			}
			ep.HealthITProductID = healthITMapID
		}
	}

	return nil
}

func getVendorMatch(ctx context.Context, capStat capabilityparser.CapabilityStatement, store *postgresql.Store) (int, error) {
	var vendorID int
	vendorsRaw, err := store.GetVendorNames(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "error retrieving vendor list from database")
	}
	vendorsNorm := normalizeList(vendorsRaw)

	match, err := publisherMatch(capStat, vendorsNorm, vendorsRaw)
	if err != nil {
		return 0, errors.Wrap(err, "error matching vendors in database using capability statement publisher")
	}

	if match == "" {
		match, err = hackMatch(capStat, vendorsNorm, vendorsRaw)
		if err != nil {
			return 0, errors.Wrap(err, "error matching vendors in database using method other than capability statement publisher")
		}
	}

	if match == "" {
		vendorID = 0
	} else {
		vendor, err := store.GetVendorUsingName(ctx, match)
		if err != nil {
			return 0, errors.Wrapf(err, "error retrieving vendor using name %s", match)
		} else {
			vendorID = vendor.ID
		}
	}

	return vendorID, nil
}

func openProductLinksFile(filepath string) (map[string]map[string]string, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	var softwareNameVersion []map[string]string
	byteValueFile, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var chplMap = make(map[string]map[string]string)
	if len(byteValueFile) != 0 {
		err = json.Unmarshal(byteValueFile, &softwareNameVersion)
		if err != nil {
			return nil, err
		}
		for _, obj := range softwareNameVersion {
			var name = obj["name"]
			var version = obj["version"]
			var chplID = obj["CHPLID"]
			if name != "" && version != "" && chplID != "" {
				if chplMap[name] == nil {
					chplMap[name] = make(map[string]string)
				}
				chplMap[name][version] = chplID
			}
		}
	}

	return chplMap, nil
}

func publisherMatch(capStat capabilityparser.CapabilityStatement, vendorsNorm []string, vendorsRaw []string) (string, error) {
	publisher, err := capStat.GetPublisher()
	if err != nil {
		return "", errors.Wrap(err, "unable to get vendor information from capability statement")
	}
	publisherNorm := normalizeName(publisher)

	match := matchName(publisherNorm, vendorsNorm, vendorsRaw)

	return match, nil
}

func matchName(name string, vendorsNorm []string, vendorsRaw []string) string {
	if name == "" {
		return ""
	}

	// substring match
	var matches []int
	for i, vendor := range vendorsNorm {
		// prioritize exact match, so return if we have an exact match
		if name == vendor {
			return vendorsRaw[i]
		}

		// collect substring matches
		if strings.Contains(vendor, name) {
			matches = append(matches, i)
		}
		if strings.Contains(name, vendor) {
			matches = append(matches, i)
		}
	}

	// if we have more than 1 match, do a tab delimited return of the matched
	// vendors. This does not provide a direct "in" to the product table like the
	// single matches do, but does alert users that there is an ambiguous match.
	if len(matches) >= 1 {
		vendors := ""
		for _, i := range matches {
			vendors += vendorsRaw[i]
			vendors += "\t"
		}
		vendors = vendors[:len(vendors)-1]
		return vendors
	}

	return ""
}

func normalizeList(names []string) []string {
	var namesNorm []string

	for _, name := range names {
		nameNorm := normalizeName(name)
		namesNorm = append(namesNorm, nameNorm)
	}

	return namesNorm
}

func normalizeName(name string) string {
	name = strings.ToLower(name)

	for _, fluff := range fluffWords {
		if strings.HasSuffix(name, fluff) {
			index := strings.LastIndex(name, fluff)
			name = name[:index]
			break
		}
	}

	name = strings.TrimRight(name, ",. ")
	return name
}

func OpenCHPLEndpointListInfoFile(filepath string) (map[string]ChplMapResults, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	var softwareListMap = make(map[string]ChplMapResults)

	byteValueFile, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var chplMap []ChplEndpointListProductInfo
	if len(byteValueFile) != 0 {
		err = json.Unmarshal(byteValueFile, &chplMap)
		if err != nil {
			return nil, err
		}
		for _, obj := range chplMap {
			var listSource = obj.ListSourceURL
			var softwareProducts = obj.SoftwareProducts

			chplMapResult := ChplMapResults{ChplProductIDs: []string{}, ChplDeveloper: []string{}}

			chplID := ""

			for _, prod := range softwareProducts {
				chplID = prod.ChplProductNumber

				if chplID != "" {
					chplMapResult.ChplProductIDs = append(chplMapResult.ChplProductIDs, chplID)
				}
			}

			if listSource != "" {
				for _, softwareProduct := range softwareProducts {
					chplMapResult.ChplDeveloper = append(chplMapResult.ChplDeveloper, softwareProduct.Developer.Name)
				}

				softwareListMap[listSource] = chplMapResult
			}

		}
	}

	return softwareListMap, nil
}
