package chplmapper

// this file is for any methods of matching an endpoint to a vendor that is not through the
// publisher field on the capability statement. These methods should only be used if an endpoint
// cannot be matched using publisher field on the capability statement.

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/pkg/errors"
)

func hackMatch(capStat capabilityparser.CapabilityStatement, vendorsNorm []string, vendorsRaw []string) (string, error) {
	return hackMatchEpic(capStat, vendorsNorm, vendorsRaw)
}

func hackMatchEpic(capStat capabilityparser.CapabilityStatement, vendorsNorm []string, vendorsRaw []string) (string, error) {
	copyright, err := capStat.GetCopyright()
	if err != nil {
		return "", errors.Wrap(err, "error getting copyright from capability statement")
	}

	if copyright == "" {
		return "", nil
	}

	copyright = strings.ToLower(copyright)
	hasEpic := strings.Contains(copyright, "epic")

	if hasEpic {
		return matchName("epic", vendorsNorm, vendorsRaw), nil
	}
	return "", nil
}
