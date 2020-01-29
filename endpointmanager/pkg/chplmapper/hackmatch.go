package chplmapper

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
		return matchName("epic", vendorsNorm, vendorsRaw)
	}
	return "", nil
}
