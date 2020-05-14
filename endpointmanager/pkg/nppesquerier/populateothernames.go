package nppesquerier

import (
	"context"
	"database/sql"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointlinker"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// OthernamesCsvLine is a struct for provider organization other names .csv downloaded from http://download.cms.gov/nppes/NPI_Files.html
type OthernamesCsvLine struct {
	NPI                              string
	Provider_Other_Organization_Name string
}

func parseOthernamesLine(line []string) OthernamesCsvLine {
	data := OthernamesCsvLine{
		NPI:                              line[0],
		Provider_Other_Organization_Name: line[1],
	}
	return data
}

func addToNPIOrgFromOthernamesCsvLine(ctx context.Context, store *postgresql.Store, data OthernamesCsvLine) (bool, error) {
	npiOrg, err := store.GetNPIOrganizationByNPIID(ctx, data.NPI)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, errors.Wrap(err, "error getting NPI org with NPI ID "+data.NPI)
	}

	normalizedName, err := endpointlinker.NormalizeOrgName(data.Provider_Other_Organization_Name)
	if err != nil {
		return false, errors.Wrap(err, "error normalizing name "+data.Provider_Other_Organization_Name)
	}

	npiOrg.AddName(data.Provider_Other_Organization_Name)
	npiOrg.AddNormalizedName(normalizedName)

	err = store.UpdateNPIOrganization(ctx, npiOrg)
	if err != nil {
		return false, errors.Wrap(err, "error updating NPI org with NPI ID "+npiOrg.NPI_ID)
	}

	return true, nil
}

// ParseAndStoreOthernamesFile parses NPI Org othernames file data out of fname, adds the names to the appropriate NPI org, and returns the number of organizations processed
func ParseAndStoreOthernamesFile(ctx context.Context, fname string, store *postgresql.Store) (int, error) {
	// Provider organization .csv downloaded from http://download.cms.gov/nppes/NPI_Files.html
	lines, err := readCsv(ctx, fname)
	if err != nil {
		return -1, err
	}
	added := 0
	// Loop through lines & turn into object
	for i, line := range lines {
		// break out of loop and return error if context has ended
		select {
		case <-ctx.Done():
			return added, errors.Wrapf(ctx.Err(), "read %d lines of the csv file before the context ended", i)
		default:
			// ok
		}

		if i%10000 == 0 {
			log.Infof("Processed %d/%d NPI 'othernames'. Added %d.", i, len(lines), added)
		}

		data := parseOthernamesLine(line)
		didAdd, err := addToNPIOrgFromOthernamesCsvLine(ctx, store, data)
		if err != nil {
			log.Warn(err)
			continue
		}
		if didAdd {
			added += 1
		}
	}
	return added, nil
}
