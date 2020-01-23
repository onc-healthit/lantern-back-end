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

// TODO: should this throw an error if there's no publisher in the capability statement?
func getVendorMatch(ctx context.Context, capStat capabilityparser.CapabilityStatement, store endpointmanager.HealthITProductStore) (string, error) {
	var vendorsNorm []string

	publisher, err := capStat.GetPublisher()
	if err != nil {
		return "", errors.Wrap(err, "unable to get vendor information from capability statement")
	}
	publisherNorm := normalizeName(publisher)

	vendorsRaw, err := store.GetHealthITProductDevelopers(ctx)
	if err != nil {
		return "", errors.Wrap(err, "error retrieving vendor list from database")
	}
	for _, vendorRaw := range vendorsRaw {
		vendorNorm := normalizeName(vendorRaw)
		vendorsNorm = append(vendorsNorm, vendorNorm)
	}

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
