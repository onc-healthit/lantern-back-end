package chplmapper

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

var MedicaidListSourceToVendor = map[string]string{
	"1up (Gainwell)":      "1up (Gainwell)",
	"Acentra":             "Acentra",
	"CNSI Provider One":   "CNSI Provider One",
	"Conduent":            "Conduent",
	"Edifecs":             "Edifecs",
	"Safhir from Onyx":    "Safhir from Onyx",
	"Salesforce/MiHIN":    "Salesforce/MiHIN",
	"State Developed":     "State Developed",
	"Implemented":         "State Developed",
	"Not Yet Implemented": "Not Available",
	"Offline":             "Not Available",
}

// Add any State Medicaid list sources that should be mapped to Unknown vendor
var MedicaidUnknownListSources = map[string]bool{
	"State Medicaid": true,
}

var MedicaidVendorCHPLIDs = map[string]int{
	"1up (Gainwell)":    2000001001,
	"Acentra":           2000001002,
	"CNSI Provider One": 2000001003,
	"Conduent":          2000001004,
	"Edifecs":           2000001005,
	"Safhir from Onyx":  2000001006,
	"Salesforce/MiHIN":  2000001007,
	"State Developed":   2000001008,
	"Not Available":     2000001009,
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
	ChplDeveloper  string
}

// MatchEndpointToVendor assigns a VendorID to an endpoint based on
// (in priority order):
// 1. CHPL developer name (from listSourceMap)
// 2. 1up list source
// 3a. Medicaid unknown category (leave vendor unassigned)
// 3b. Medicaid list source category (known)
// 4. CapabilityStatement fallback
func MatchEndpointToVendor(
	ctx context.Context,
	ep *endpointmanager.FHIREndpointInfo,
	store *postgresql.Store,
	listSourceMap map[string]ChplMapResults,
) error {

	log.Infof("[MatchEndpointToVendor] Starting for URL=%s", ep.URL)

	fhirEndpointList, err := store.GetFHIREndpointUsingURL(ctx, ep.URL)
	if err != nil {
		return errors.Wrap(err, "error getting fhir endpoints from DB")
	}

	log.Infof("[MatchEndpointToVendor] Found %d DB FHIR endpoint records for URL=%s", len(fhirEndpointList), ep.URL)

	// Iterate over all DB rows with this URL; return as soon as we find a vendor
	for _, fhirEndpoint := range fhirEndpointList {
		listSource := fhirEndpoint.ListSource
		chplInfo := listSourceMap[listSource]
		developerName := chplInfo.ChplDeveloper

		log.Infof("[MatchEndpointToVendor] Checking listSource='%s' developer='%s'", listSource, developerName)

		// ------------------------------------------------------------
		// 1. CHPL DEVELOPER NAME (highest priority)
		// ------------------------------------------------------------
		if developerName != "" {
			vendorMatch, err := store.GetVendorUsingName(ctx, developerName)

			if err == nil {
				// No errors thrown means a vendor with developer name was found and can be set on ep
				log.Infof("[MatchEndpointToVendor] CHPL vendor matched: %s (ID=%d)", vendorMatch.Name, vendorMatch.ID)
				ep.VendorID = vendorMatch.ID
				return nil
			}

			if err != sql.ErrNoRows {
				return errors.Wrap(err, "error matching the CHPL endpoint list developer name to a vendor for endpoint")
			}

			// CHPL name exists but vendor not found -> FALL THROUGH
			log.Warnf("[MatchEndpointToVendor] No vendor for CHPL developer '%s' — trying special cases",
				developerName)
		}

		// --------------------------------------------------
		// 2. 1UP List Source
		// --------------------------------------------------
		if listSource == "https://1up.health/fhir-endpoint-directory" {
			vendorName := "1upHealth"
			log.Infof("[MatchEndpointToVendor] 1up list source detected")

			vendorMatch, err := store.GetVendorUsingName(ctx, vendorName)
			if err == sql.ErrNoRows {
				log.Warn("[MatchEndpointToVendor] 1up vendor missing — creating")

				newVendor := &endpointmanager.Vendor{
					Name:          vendorName,
					URL:           "https://1up.health",
					CHPLID:        2000000000,
					DeveloperCode: "2000000000",
				}
				if err := store.AddVendor(ctx, newVendor); err != nil {
					return errors.Wrap(err, "failed to insert vendor 1upHealth")
				}

				ep.VendorID = newVendor.ID
				log.Infof("[MatchEndpointToVendor] Created new vendor 1upHealth (ID=%d)", newVendor.ID)
				return nil
			}
			if err != nil {
				return errors.Wrap(err, "error querying vendor 1upHealth")
			}

			ep.VendorID = vendorMatch.ID
			return nil
		}

		// --------------------------------------------------
		// 3. Medicaid list-source mappings
		// --------------------------------------------------

		// 3a. Unknown Medicaid list sources -> leave vendor unset
		if MedicaidUnknownListSources[listSource] {
			log.Infof("[MatchEndpointToVendor] Medicaid unknown listSource='%s' — leaving vendor unassigned", listSource)
			// Design choice: do NOT fall back to CS for these;
			// return and keep VendorID = 0.
			return nil
		}

		// 3b. Known Medicaid vendors
		if vendorName, ok := MedicaidListSourceToVendor[listSource]; ok {
			log.Infof("[MatchEndpointToVendor] Medicaid listSource='%s' -> vendor='%s'", listSource, vendorName)

			vendorMatch, err := store.GetVendorUsingName(ctx, vendorName)
			if err == sql.ErrNoRows {
				log.Warnf("[MatchEndpointToVendor] Medicaid vendor '%s' not found — creating", vendorName)

				chplID, ok := MedicaidVendorCHPLIDs[vendorName]
				if !ok {
					return fmt.Errorf("no static CHPLID configured for Medicaid vendor '%s'", vendorName)
				}

				newVendor := &endpointmanager.Vendor{
					Name:          vendorName,
					URL:           "",
					CHPLID:        chplID,
					DeveloperCode: fmt.Sprintf("%d", chplID),
				}
				if err := store.AddVendor(ctx, newVendor); err != nil {
					return errors.Wrap(err, "failed inserting Medicaid vendor")
				}

				ep.VendorID = newVendor.ID
				log.Infof("[MatchEndpointToVendor] Created Medicaid vendor '%s' (ID=%d)", vendorName, newVendor.ID)
				return nil
			}
			if err != nil {
				return errors.Wrap(err, "error querying Medicaid vendor")
			}

			ep.VendorID = vendorMatch.ID
			log.Infof("[MatchEndpointToVendor] Matched Medicaid vendor '%s' (ID=%d)", vendorName, vendorMatch.ID)
			return nil
		}
	}

	// --------------------------------------------------
	// 4. Fallback to CapabilityStatement-based matching
	// --------------------------------------------------
	if ep.CapabilityStatement == nil {
		log.Warn("[MatchEndpointToVendor] No capability statement available — cannot match vendor from CS")
		return nil
	}

	log.Infof("[MatchEndpointToVendor] Falling back to capability statement matching")

	vendorID, err := getVendorMatch(ctx, ep.CapabilityStatement, store)
	if err != nil {
		return errors.Wrap(err, "CS vendor matching failure")
	}

	log.Infof("[MatchEndpointToVendor] Result from CS vendor matching: vendorID=%d", vendorID)
	ep.VendorID = vendorID

	return nil
}

// MatchEndpointToProduct creates the database association between the endpoint and the HealthITProduct,
func MatchEndpointToProduct(ctx context.Context, ep *endpointmanager.FHIREndpointInfo, store *postgresql.Store, matchFile string, listSourceMap map[string]ChplMapResults) error {

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

	// If endpoint's list source found in CHPL endpoint list, match to product associated with that list source
	fhirEndpointList, err := store.GetFHIREndpointUsingURL(ctx, ep.URL)
	if err != nil {
		return errors.Wrap(err, "error getting fhir endpoints from DB")
	}

	for _, fhirEndpoint := range fhirEndpointList {
		chplIDList := listSourceMap[fhirEndpoint.ListSource].ChplProductIDs
		if len(chplIDList) > 0 {
			chplIDArr = append(chplIDArr, chplIDList...)
		}
	}

	var healthITProductsArr []*endpointmanager.HealthITProduct
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
		return 0, errors.Wrap(err, "error matching vendors in database using capability statement publisher")
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
			return 0, errors.Wrapf(err, "error retrieving vendor using name %s", match)
		} else {
			log.Infof("[getVendorMatch] Matched vendor: %s (ID=%d)", vendor.Name, vendor.ID)
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

			chplMapResult := ChplMapResults{ChplProductIDs: []string{}, ChplDeveloper: ""}

			chplID := ""

			for _, prod := range softwareProducts {
				chplID = prod.ChplProductNumber

				if chplID != "" {
					chplMapResult.ChplProductIDs = append(chplMapResult.ChplProductIDs, chplID)
				}
			}

			if listSource != "" {
				if len(softwareProducts) > 0 {
					// Developer is the same for all products, just grab first one
					chplMapResult.ChplDeveloper = softwareProducts[0].Developer.Name
				}

				softwareListMap[listSource] = chplMapResult
			}

		}
	}

	return softwareListMap, nil
}
