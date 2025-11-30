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

	log.Infof("[MatchEndpointToVendor] Starting match for URL=%s developerName=%s", ep.URL, developerName)

	if len(developerName) > 0 {
		// --- 1up special handling ---
		if developerName == "https://1up.health/fhir-endpoint-directory" {
			log.Infof("[MatchEndpointToVendor] Special-case match: 1upHealth")

			vendorName := "1upHealth"
			vendorMatch, err := store.GetVendorUsingName(ctx, vendorName)
			if err == sql.ErrNoRows {
				log.Warn("[MatchEndpointToVendor] 1upHealth vendor not found — creating new vendor record")

				newVendor := &endpointmanager.Vendor{
					Name:          vendorName,
					URL:           "https://1up.health",
					CHPLID:        2000000000,
					DeveloperCode: "2000000000",
				}
				err = store.AddVendor(ctx, newVendor)
				if err != nil {
					return errors.Wrap(err, "failed to insert 1upHealth as new vendor")
				}
				log.Infof("[MatchEndpointToVendor] Created new vendor 1upHealth (ID=%d)", newVendor.ID)
				ep.VendorID = newVendor.ID
				return nil
			} else if err != nil {
				return errors.Wrap(err, "error checking for existing 1upHealth vendor")
			}

			log.Infof("[MatchEndpointToVendor] Matched vendor 1upHealth (ID=%d)", vendorMatch.ID)
			ep.VendorID = vendorMatch.ID
			return nil
		}

		// --- State Medicaid special handling ---
		if developerName == "StateMedicaid" {
			log.Infof("[MatchEndpointToVendor] Special-case match: StateMedicaid")

			vendorName := "State Medicaid"
			vendorMatch, err := store.GetVendorUsingName(ctx, vendorName)
			if err == sql.ErrNoRows {
				log.Warn("[MatchEndpointToVendor] Medicaid vendor not found — creating new vendor record")

				newVendor := &endpointmanager.Vendor{
					Name:          vendorName,
					URL:           "",
					CHPLID:        2000000001,
					DeveloperCode: "2000000001",
				}
				err = store.AddVendor(ctx, newVendor)
				if err != nil {
					return errors.Wrap(err, "failed to insert State Medicaid as new vendor")
				}
				log.Infof("[MatchEndpointToVendor] Created new vendor State Medicaid (ID=%d)", newVendor.ID)
				ep.VendorID = newVendor.ID
				return nil
			} else if err != nil {
				return errors.Wrap(err, "error checking for existing StateMedicaid vendor")
			}

			log.Infof("[MatchEndpointToVendor] Matched vendor State Medicaid (ID=%d)", vendorMatch.ID)
			ep.VendorID = vendorMatch.ID
			return nil
		}

		log.Infof("[MatchEndpointToVendor] CHPL developer name provided directly: %s", developerName)

		vendorMatch, err := store.GetVendorUsingName(ctx, developerName)
		if err == sql.ErrNoRows {
			log.Warnf("[MatchEndpointToVendor] No vendor found for CHPL developerName=%s", developerName)
			return nil
		} else if err != nil {
			return errors.Wrap(err, "error matching CHPL developer name to vendor")
		}

		log.Infof("[MatchEndpointToVendor] Matched vendor %s (ID=%d)", vendorMatch.Name, vendorMatch.ID)
		ep.VendorID = vendorMatch.ID
		return nil
	}

	if ep.CapabilityStatement == nil {
		log.Warn("[MatchEndpointToVendor] No capability statement available — cannot match vendor from CS")
		return nil
	}

	log.Infof("[MatchEndpointToVendor] Falling back to capability statement matching")
	vendorID, err := getVendorMatch(ctx, ep.CapabilityStatement, store)
	if err != nil {
		return errors.Wrap(err, "error matching capability statement to vendor")
	}

	log.Infof("[MatchEndpointToVendor] Result from CS vendor matching: vendorID=%d", vendorID)
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
	log.Infof("[getVendorMatch] Attempting vendor match from capability statement")

	var vendorID int
	vendorsRaw, err := store.GetVendorNames(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "error retrieving vendor list from database")
	}

	vendorsNorm := normalizeList(vendorsRaw)

	match, err := publisherMatch(capStat, vendorsNorm, vendorsRaw)
	log.Infof("[getVendorMatch] publisherMatch result: %s", match)

	if err != nil {
		return 0, errors.Wrap(err, "error matching via capability statement publisher")
	}

	if match == "" {
		match, err = hackMatch(capStat, vendorsNorm, vendorsRaw)
		log.Infof("[getVendorMatch] hackMatch result: %s", match)

		if err != nil {
			return 0, errors.Wrap(err, "error matching via hackMatch")
		}
	}

	if match == "" {
		log.Warn("[getVendorMatch] No vendor match found — returning vendorID=0")
		vendorID = 0
	} else {
		vendor, err := store.GetVendorUsingName(ctx, match)
		if err != nil {
			return 0, errors.Wrapf(err, "error retrieving vendor name %s", match)
		}
		log.Infof("[getVendorMatch] Matched vendor: %s (ID=%d)", vendor.Name, vendor.ID)
		vendorID = vendor.ID
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
	log.Infof("[publisherMatch] Attempting publisher-based match")

	publisher, err := capStat.GetPublisher()
	if err != nil {
		return "", errors.Wrap(err, "unable to get vendor information from capability statement")
	}
	publisherNorm := normalizeName(publisher)

	log.Infof("[publisherMatch] publisher=%s normalized=%s", publisher, publisherNorm)

	match := matchName(publisherNorm, vendorsNorm, vendorsRaw)

	log.Infof("[publisherMatch] match result=%s", match)

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
