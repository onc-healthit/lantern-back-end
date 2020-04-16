package nppesquerier

import (
	"context"
	"encoding/csv"
	"os"
	"strings"
	"regexp"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointlinker"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

// "endpoint_pfile" .csv downloaded from http://download.cms.gov/nppes/NPI_Files.html
type NPIContactCsvLine struct {
	NPI									string
	Endpoint_Type						string
	Endpoint_Type_Description			string
	Endpoint							string
	Affiliation							string
	Endpoint_Description				string
	Affiliation_Legal_Business_Name		string
	Use_Code							string
	Use_Description						string
	Other_Use_Description				string
	Content_Type						string
	Content_Description					string
	Other_Content_Description			string
	Affiliation_Address_Line_One		string
	Affiliation_Address_Line_Two		string
	Affiliation_Address_City			string
	Affiliation_Address_State			string
	Affiliation_Address_Country			string
	Affiliation_Address_Postal_Code		string
}

func parseNPIContactdataLine(line []string) NPIContactCsvLine {
	data := NPIContactCsvLine{
		NPI:	line[0],
		Endpoint_Type:	line[1],
		Endpoint_Type_Description:	line[2],
		Endpoint:	line[3],
		Affiliation:	line[4],
		Endpoint_Description:	line[5],
		Affiliation_Legal_Business_Name:	line[6],
		Use_Code:	line[7],
		Use_Description:	line[8],
		Other_Use_Description:	line[9],
		Content_Type:	line[10],
		Content_Description:	line[11],
		Other_Content_Description:	line[12],
		Affiliation_Address_Line_One:	line[13],
		Affiliation_Address_Line_Two:	line[14],
		Affiliation_Address_City:	line[15],
		Affiliation_Address_State:	line[16],
		Affiliation_Address_Country:	line[17],
		Affiliation_Address_Postal_Code:	line[18],
	}
	return data
}

func buildNPIContactFromNPICsvLine(data NPIContactCsvLine) *endpointmanager.NPIContact {
	normalizedName := endpointlinker.NormalizeOrgName(data.Affiliation_Legal_Business_Name)
	validURL := isValidURL(data.Endpoint)
	npiContact := &endpointmanager.NPIContact{
		NPI_ID:        data.NPI,
		Endpoint_Type:	data.Endpoint_Type,
		Endpoint_Type_Description:	data.Endpoint_Type_Description,
		Endpoint:	data.Endpoint,
		Valid_URL: validURL,
		Affiliation:	data.Affiliation,
		Endpoint_Description:	data.Endpoint_Description,
		Affiliation_Legal_Business_Name:	data.Affiliation_Legal_Business_Name,
		Normalized_Affiliation_Legal_Business_Name: normalizedName,
		Use_Code:	data.Use_Code,
		Use_Description:	data.Use_Description,
		Other_Use_Description:	data.Other_Use_Description,
		Content_Type:	data.Content_Type,
		Content_Description:	data.Content_Description,
		Other_Content_Description:	data.Other_Content_Description,
		Location: &endpointmanager.Location{
			Address1: data.Affiliation_Address_Line_One,
			Address2: data.Affiliation_Address_Line_Two,
			City:     data.Affiliation_Address_City,
			State:    data.Affiliation_Address_State,
			ZipCode:  data.Affiliation_Address_Postal_Code},
	}
	return npiContact
}

func isValidURL(url string) bool {
	urlregex := `^(?:http(s)?:\/\/)?[\w.-]+(?:\.[\w\.-]+)+[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+$`
	matched, _ := regexp.MatchString(urlregex, strings.ToLower(url))
	return matched
}

// readContactCsv accepts a file and returns its content as a multi-dimentional type
// with lines and each column. Only parses to string type.
func readContactCsv(ctx context.Context, filename string) ([][]string, error) {
	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "did not read csv; context ended")
	default:
		// ok
	}

	// Open CSV file
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}
	// return lines without header line
	return lines[1:], nil
}

// ParseAndStoreNPIFile parses NPI Org data out of fname, writes it to store and returns the number of Contacts processed
func ParseAndStoreNPIContactsFile(ctx context.Context, fname string, store *postgresql.Store) (int, error) {
	// Provider Contact .csv downloaded from http://download.cms.gov/nppes/NPI_Files.html
	lines, err := readContactCsv(ctx, fname)
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

		data := parseNPIContactdataLine(line)
		// We will only parse out Contacts with endpoint_type of FHIR
		if data.Endpoint_Type == "FHIR" {
			npiContact := buildNPIContactFromNPICsvLine(data)
			err = store.AddNPIContact(ctx, npiContact)
			if err != nil {
				log.Error(err)
			} else {
				added += 1
			}
		}
	}
	return added, nil
}
