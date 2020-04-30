package chplmapper

import (
	"context"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
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

// MatchEndpointToVendorAndProduct creates the database association between the endpoint and the vendor,
// and the endpoint and the healht IT product.
// It returns a boolean specifying if the match was possible or not.
//
// NOTE: at this time, only vendor matching is supported.
// An endpoint is matched to a vendor by adding the vendor to the endpoint entry in the database.
// In this future, this may be changed to using a vendor table and linking the endpoint entry to
// the vendor entry.
func MatchEndpointToVendorAndProduct(ctx context.Context, ep *endpointmanager.FHIREndpointInfo, store *postgresql.Store) error {
	if ep.CapabilityStatement == nil {
		return nil
	}

	vendor, err := getVendorMatch(ctx, ep.CapabilityStatement, store)
	if err != nil {
		return errors.Wrap(err, "error matching the capability statement to a vendor for endpoint")
	}

	ep.Vendor = vendor

	return nil
}

func getVendorMatch(ctx context.Context, capStat capabilityparser.CapabilityStatement, store *postgresql.Store) (string, error) {
	vendorsRaw, err := store.GetHealthITProductDevelopers(ctx)
	if err != nil {
		return "", errors.Wrap(err, "error retrieving vendor list from database")
	}
	vendorsNorm := normalizeList(vendorsRaw)

	match, err := publisherMatch(capStat, vendorsNorm, vendorsRaw)
	if err != nil {
		return "", errors.Wrap(err, "error matching health it developers in database using capability statement publisher")
	}

	if match == "" {
		match, err = hackMatch(capStat, vendorsNorm, vendorsRaw)
		if err != nil {
			return "", errors.Wrap(err, "error matching health it developers in database using method other than capability statement publisher")
		}
	}

	return match, nil
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
	// TODO: update this once we matched into a vendor table.
	if len(matches) >= 1 {
		vendors := ""
		for _, i := range matches {
			vendors += vendorsRaw[i]
			vendors += "\t"
		}
		vendors = vendors[:len(vendors)-1]
		return vendors
	}

	// fuzzy match
	// TODO

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
