package chplmapper

import (
	"fmt"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_matchName(t *testing.T) {
	var expected string
	var actual string
	var dev string

	devList := []string{
		"Epic Systems Corporation",
		"Cerner Group", // changed for sake of test
		"Cerner Health Services, Inc.",
		"Medical Information Technology, Inc. (MEDITECH)",
		"Allscripts",
	}
	devListNorm := normalizeList(devList)

	// allscripts
	expected = "Allscripts"
	dev = normalizeName("Allscripts")
	actual = matchName(dev, devListNorm, devList)
	th.Assert(t, expected == actual, fmt.Sprintf("Expected %s. Got %s.", expected, actual))

	// meditech
	expected = "Medical Information Technology, Inc. (MEDITECH)"
	dev = normalizeName("Medical Information Technology, Inc")
	actual = matchName(dev, devListNorm, devList)
	th.Assert(t, expected == actual, fmt.Sprintf("Expected %s. Got %s.", expected, actual))

	// cerner
	expected = "Cerner Group\tCerner Health Services, Inc."
	dev = normalizeName("Cerner")
	actual = matchName(dev, devListNorm, devList)
	th.Assert(t, expected == actual, fmt.Sprintf("Expected %s. Got %s.", expected, actual))
}
