package nppesquerier

import (
	"bufio"
	"context"
	"encoding/csv"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

// "endpoint_pfile" .csv downloaded from http://download.cms.gov/nppes/NPI_Files.html
type NPIContactCsvLine struct {
	NPI                             string
	EndpointType                    string
	EndpointTypeDescription         string
	Endpoint                        string
	Affiliation                     string
	EndpointDescription             string
	AffiliationLegalBusinessName    string
	UseCode                         string
	UseDescription                  string
	OtherUseDescription             string
	ContentType                     string
	ContentDescription              string
	OtherContentDescription         string
	Affiliation_Address_Line_One    string
	Affiliation_Address_Line_Two    string
	Affiliation_Address_City        string
	Affiliation_Address_State       string
	Affiliation_Address_Country     string
	Affiliation_Address_Postal_Code string
}

func parseNPIContactdataLine(line []string) NPIContactCsvLine {
	data := NPIContactCsvLine{
		NPI:                             line[0],
		EndpointType:                    line[1],
		EndpointTypeDescription:         line[2],
		Endpoint:                        line[3],
		Affiliation:                     line[4],
		EndpointDescription:             line[5],
		AffiliationLegalBusinessName:    line[6],
		UseCode:                         line[7],
		UseDescription:                  line[8],
		OtherUseDescription:             line[9],
		ContentType:                     line[10],
		ContentDescription:              line[11],
		OtherContentDescription:         line[12],
		Affiliation_Address_Line_One:    line[13],
		Affiliation_Address_Line_Two:    line[14],
		Affiliation_Address_City:        line[15],
		Affiliation_Address_State:       line[16],
		Affiliation_Address_Country:     line[17],
		Affiliation_Address_Postal_Code: line[18],
	}
	return data
}

func buildNPIContactFromNPICsvLine(data NPIContactCsvLine) *endpointmanager.NPIContact {
	validURL := isValidURL(data.Endpoint)
	data.Endpoint = strings.Replace(data.Endpoint, "/metadata", "", 1)

	// Add trailing "/" to URIs that do not have it for consistency
	if len(data.Endpoint) > 0 && data.Endpoint[len(data.Endpoint)-1:] != "/" {
		data.Endpoint = data.Endpoint + "/"
	}

	splitEndpoint := strings.Split(data.Endpoint, "://")
	data.Endpoint = "https://" + splitEndpoint[len(splitEndpoint)-1]

	npiContact := &endpointmanager.NPIContact{
		NPI_ID:                       data.NPI,
		EndpointType:                 data.EndpointType,
		EndpointTypeDescription:      data.EndpointTypeDescription,
		Endpoint:                     data.Endpoint,
		ValidURL:                     validURL,
		Affiliation:                  data.Affiliation,
		EndpointDescription:          data.EndpointDescription,
		AffiliationLegalBusinessName: data.AffiliationLegalBusinessName,
		UseCode:                      data.UseCode,
		UseDescription:               data.UseDescription,
		OtherUseDescription:          data.OtherUseDescription,
		ContentType:                  data.ContentType,
		ContentDescription:           data.ContentDescription,
		OtherContentDescription:      data.OtherContentDescription,
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
	urlregex := regexp.MustCompile(`^(?:http(s)?:\/\/)?[\w.-]+(?:\.[\w\.-]+)+[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+$`)
	urlmatched := urlregex.MatchString(strings.ToLower(url))
	// Filter out emails from valid URLS
	emailregex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	emailmatched := emailregex.MatchString(strings.ToLower(url))

	return urlmatched && !emailmatched
}

func removeNestedDoubleQuotesFromCSV(filename string) (error, string) {
	file, err := os.Open(filename)

	if err != nil {
		return err, ""
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var processedlines []string

	for scanner.Scan() {
		line := scanner.Text()
		// These 2 regexes (pattern1 and pattern2) capture every other value enclosed in quotes in the csv
		// combined they are able to examine all values. The matches resulting from the regexes need to be
		// iterated over independently as the values in the regexes are captured in different group indexes
		pattern1 := regexp.MustCompile(`(,|",")(.*?)","`)
		matches1 := pattern1.FindAllStringSubmatch(line, -1)
		for _, match := range matches1 {
			// If detected nested quote, remove nested quotes
			if strings.Contains(match[2], "\"") {
				sanitized := strings.Replace(match[2], "\"", "", -1)
				// replace entry with quotes removed
				line = strings.Replace(line, match[2], sanitized, 1)
			}
		}
		pattern2 := regexp.MustCompile(`"(.*?)(,|",")`)
		matches2 := pattern2.FindAllStringSubmatch(line, -1)
		for _, match := range matches2 {
			// If detected nested quote, remove nested quotes
			if strings.Contains(match[1], "\"") {
				sanitized := strings.Replace(match[1], "\"", "", -1)
				// replace entry with quotes removed
				line = strings.Replace(line, match[1], sanitized, 1)
			}
		}

		processedlines = append(processedlines, line)
	}

	file.Close()

	newfilename := strings.Replace(filename, ".csv", "", 1)
	newfile, err := os.Create(newfilename)
	if err != nil {
		return err, ""
	}
	for _, processedline := range processedlines {
		_, err := newfile.WriteString(processedline + "\n")
		if err != nil {
			return err, ""
		}
	}

	defer newfile.Close()

	return nil, newfilename
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

	err, newfilename := removeNestedDoubleQuotesFromCSV(filename)
	if err != nil {
		return [][]string{}, err
	}

	// Open CSV file
	f, err := os.Open(newfilename)
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

// ParseAndStoreNPIContactsFile parses NPI Org data out of fname, writes it to store and returns the number of Contacts processed
func ParseAndStoreNPIContactsFile(ctx context.Context, fname string, store *postgresql.Store) (int, error) {
	// Provider Contact .csv downloaded from http://download.cms.gov/nppes/NPI_Files.html
	lines, err := readContactCsv(ctx, fname)
	if err != nil {
		return -1, err
	}
	log.Infof("Adding NPI contacts and fhir endpoints.")
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
		if data.EndpointType == "FHIR" {
			npiContact := buildNPIContactFromNPICsvLine(data)
			err = store.AddNPIContact(ctx, npiContact)
			if err != nil {
				log.Error(err)
			} else {
				added += 1
			}
			// If contact has a valid URL, add to our fhir endpoints table, source list is NPPES
			if npiContact.ValidURL {
				var fhirEndpoint = &endpointmanager.FHIREndpoint{
					URL:        npiContact.Endpoint,
					ListSource: "NPPES"}
				if npiContact.AffiliationLegalBusinessName != "" {
					fhirEndpoint.OrganizationNames = []string{npiContact.AffiliationLegalBusinessName}
				}
				if npiContact.NPI_ID != "" {
					fhirEndpoint.NPIIDs = []string{npiContact.NPI_ID}
				}
				err = store.AddOrUpdateFHIREndpoint(ctx, fhirEndpoint)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}
	log.Infof("Added %d NPI contacts\n", added)
	return added, nil
}
