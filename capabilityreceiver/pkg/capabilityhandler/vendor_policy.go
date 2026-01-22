package capabilityhandler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
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
	"1up (Gainwell)":    "1up (Gainwell)",
	"Acentra":           "Acentra",
	"CNSI Provider One": "CNSI Provider One",
	"Conduent":          "Conduent",
	"Edifecs":           "Edifecs",
	"Safhir from Onyx":  "Safhir from Onyx",
	"Salesforce/MiHIN":  "Salesforce/MiHIN",
	"State Developed":   "State Developed",
	"Not Available":     "Not Available",
}

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

type VendorMatchSource string

const (
	VendorMatchCHPL            VendorMatchSource = "chpl"
	VendorMatch1Up             VendorMatchSource = "1up"
	VendorMatchMedicaidKnown   VendorMatchSource = "medicaid_known"
	VendorMatchMedicaidUnknown VendorMatchSource = "medicaid_unknown"
	VendorMatchCapability      VendorMatchSource = "capability_statement"
	VendorMatchNone            VendorMatchSource = "none"
)

type VendorMatchResult struct {
	VendorID int
	Source   VendorMatchSource
	Detail   string
}

func ResolveVendor(
	ctx context.Context,
	store *postgresql.Store,
	listSource string,
	developerName string,
	capStat capabilityparser.CapabilityStatement,
) (VendorMatchResult, error) {

	log.Infof(
		"[ResolveVendor] start listSource=%q developerName=%q hasCapStat=%v",
		listSource,
		developerName,
		capStat != nil,
	)

	// --------------------------------------------------
	// 1. CHPL developer (highest priority)
	// --------------------------------------------------
	if developerName != "" {
		v, err := store.GetVendorUsingName(ctx, developerName)
		if err == nil {
			log.Infof(
				"[ResolveVendor] matched via CHPL developer name=%q vendorID=%d",
				developerName,
				v.ID,
			)

			return VendorMatchResult{
				VendorID: v.ID,
				Source:   VendorMatchCHPL,
				Detail:   developerName,
			}, nil
		}
		if err != sql.ErrNoRows {
			return VendorMatchResult{}, errors.Wrap(err, "query vendor by CHPL developer")
		}
		// fall through if not found
	}

	// --------------------------------------------------
	// 2. 1up list source
	// --------------------------------------------------
	if listSource == "https://1up.health/fhir-endpoint-directory" {
		vendorID, err := ensure1UpVendor(ctx, store)
		if err != nil {
			return VendorMatchResult{}, err
		}

		log.Infof(
			"[ResolveVendor] matched via 1up list source vendorID=%d",
			vendorID,
		)

		return VendorMatchResult{
			VendorID: vendorID,
			Source:   VendorMatch1Up,
			Detail:   "1upHealth",
		}, nil
	}

	// --------------------------------------------------
	// 3a. Medicaid unknown list sources
	// --------------------------------------------------
	if MedicaidUnknownListSources[listSource] {
		log.Warnf(
			"[ResolveVendor] Medicaid UNKNOWN list source=%q → vendorID=0",
			listSource,
		)

		return VendorMatchResult{
			VendorID: 0,
			Source:   VendorMatchMedicaidUnknown,
			Detail:   listSource,
		}, nil
	}

	// --------------------------------------------------
	// 3b. Medicaid known list sources
	// --------------------------------------------------
	if vendorName, ok := MedicaidListSourceToVendor[listSource]; ok {
		vendorID, err := ensureMedicaidVendor(ctx, store, vendorName)
		if err != nil {
			return VendorMatchResult{}, err
		}

		log.Infof(
			"[ResolveVendor] Medicaid KNOWN vendor=%q vendorID=%d",
			vendorName,
			vendorID,
		)

		return VendorMatchResult{
			VendorID: vendorID,
			Source:   VendorMatchMedicaidKnown,
			Detail:   vendorName,
		}, nil
	}

	// --------------------------------------------------
	// 4. CapabilityStatement fallback
	// --------------------------------------------------
	if capStat == nil {
		log.Warnf(
			"[ResolveVendor] no capability statement → vendorID=0 listSource=%q",
			listSource,
		)

		return VendorMatchResult{
			VendorID: 0,
			Source:   VendorMatchNone,
			Detail:   "no capability statement",
		}, nil
	}

	vendorID, err := getVendorMatch(ctx, capStat, store)
	if err != nil {
		return VendorMatchResult{}, errors.Wrap(err, "capability vendor match failed")
	}

	return VendorMatchResult{
		VendorID: vendorID,
		Source:   VendorMatchCapability,
		Detail:   "capability statement",
	}, nil
}

func ensure1UpVendor(ctx context.Context, store *postgresql.Store) (int, error) {
	const vendorName = "1upHealth"

	v, err := store.GetVendorUsingName(ctx, vendorName)
	if err == nil {
		return v.ID, nil
	}
	if err != sql.ErrNoRows {
		return 0, errors.Wrap(err, "query 1up vendor")
	}

	newVendor := &endpointmanager.Vendor{
		Name:          vendorName,
		URL:           "https://1up.health",
		CHPLID:        2000000000,
		DeveloperCode: "2000000000",
	}
	if err := store.AddVendor(ctx, newVendor); err != nil {
		return 0, errors.Wrap(err, "insert 1up vendor")
	}
	return newVendor.ID, nil
}

func ensureMedicaidVendor(ctx context.Context, store *postgresql.Store, vendorName string) (int, error) {
	v, err := store.GetVendorUsingName(ctx, vendorName)
	if err == nil {
		return v.ID, nil
	}
	if err != sql.ErrNoRows {
		return 0, errors.Wrap(err, "query Medicaid vendor")
	}

	chplID, ok := MedicaidVendorCHPLIDs[vendorName]
	if !ok {
		return 0, fmt.Errorf("no CHPLID configured for Medicaid vendor %q", vendorName)
	}

	newVendor := &endpointmanager.Vendor{
		Name:          vendorName,
		URL:           "",
		CHPLID:        chplID,
		DeveloperCode: fmt.Sprintf("%d", chplID),
	}
	if err := store.AddVendor(ctx, newVendor); err != nil {
		return 0, errors.Wrap(err, "insert Medicaid vendor")
	}
	return newVendor.ID, nil
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
