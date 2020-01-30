package chplmapper

import (
	"context"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
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
	"corporation",
}

// MatchEndpointToVendorAndProduct creates the database association between the endpoint and the vendor,
// and the endpoint and the healht IT product.
// It returns a boolean specifying if the match was possible or not.
//
// NOTE: at this time, only vendor matching is supported.
// An endpoint is matched to a vendor by adding the vendor to the endpoint entry in the database.
// In this future, this may be changed to using a vendor table and linking the endpoint entry to
// the vendor entry.
func MatchEndpointToVendorAndProduct(ctx context.Context, ep endpointmanager.FHIREndpoint, store endpointmanager.HealthITProductStore) (bool, error) {
	if ep.CapabilityStatement == nil {
		return false, nil
	}

	_, err := getVendorMatch(ctx, ep.CapabilityStatement, store)
	if err != nil {
		return false, errors.Wrapf(err, "error matching the capability statement to a vendor for endpoint %s", ep.URL)
	}

	return false, nil
}

// TODO: should this throw an error if there's no publisher in the capability statement?
func getVendorMatch(ctx context.Context, capStat capabilityparser.CapabilityStatement, store endpointmanager.HealthITProductStore) (string, error) {
	var vendorsNorm []string

	vendorsRaw, err := store.GetHealthITProductDevelopers(ctx)
	if err != nil {
		return "", errors.Wrap(err, "error retrieving vendor list from database")
	}
	for _, vendorRaw := range vendorsRaw {
		vendorNorm := normalizeName(vendorRaw)
		vendorsNorm = append(vendorsNorm, vendorNorm)
	}

	match, err := publisherMatch(capStat, vendorsNorm, vendorsRaw)
	if err != nil {
		return "", errors.Wrap(err, "error matching health it developers in database using method other than capability statement publisher")
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

	match, err := matchName(publisherNorm, vendorsNorm, vendorsRaw)
	if err != nil {
		return "", errors.Wrap(err, "error matching capability statement publisher to health it developers in database")
	}

	return match, nil
}

func matchName(name string, vendorsNorm []string, vendorsRaw []string) (string, error) {
	// exact match
	for i, vendor := range vendorsNorm {
		if name == vendor {
			return vendorsRaw[i], nil
		}
	}

	// substring match
	var matches []int
	for i, vendor := range vendorsNorm {
		if strings.Contains(vendor, name) {
			matches = append(matches, i)
		}
	}
	for i, vendor := range vendorsNorm {
		if strings.Contains(name, vendor) {
			matches = append(matches, i)
		}
	}
	if len(matches) == 1 {
		return vendorsRaw[matches[0]], nil
	}

	// fuzzy match
	// TODO

	return "", nil
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
