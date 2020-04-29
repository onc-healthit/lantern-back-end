package nppesquerier

import (
	"context"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"
)

func Test_ParseNPIContactdataLine(t *testing.T) {
	ctx := context.Background()

	lines, err := readContactCsv(ctx, "testdata/npi_contact_file.csv")
	if err != nil {
		t.Errorf("Error reading NPI data from fixture file")
	}
	// fixture file has 44414 lines, we throw out the header line
	if len(lines) != 44413 {
		t.Errorf("Expected %d lines to be read from CSV, got %d", 44413, len(lines))
	}

	// This is the 3rd line in the test file, empty entries have been populated with test strings for more realistic testing
	// "1427051648","OTHERS","Other URL","https://foobar.com/metadata/","N","Test_EndpointDescription","Test_AffiliationLegalBusinessName","Test_UseCode","Test_UseDescription","Test_OtherUseDescription","Test_ContentType","Test_ContentDescription","Test_OtherContentDescription","30454 E Rustic Drive","Test_Address_Line_Two","Salisbury","MD","US","21804"
	data := parseNPIContactdataLine(lines[2])
	// NPI
	if data.NPI != "1427051648" {
		t.Errorf("Expected NPI to be %s, got %s", "1427051648", data.NPI)
	}
	// EndpointType
	if data.EndpointType != "OTHERS" {
		t.Errorf("Expected EndpointType to be %s, got %s", "OTHERS", data.EndpointType)
	}
	// EndpointTypeDescription
	if data.EndpointTypeDescription != "Other URL" {
		t.Errorf("Expected EndpointTypeDescription to be %s, got %s", "Other URL", data.EndpointTypeDescription)
	}
	// Endpoint
	if data.Endpoint != "https://foobar.com/metadata/" {
		t.Errorf("Expected Endpoint to be %s, got %s", "Salisbury", data.Endpoint)
	}
	// Affiliation
	if data.Affiliation != "N" {
		t.Errorf("Expected Affiliation to be %s, got %s", "N", data.Affiliation)
	}
	// EndpointDescription
	if data.EndpointDescription != "Test_EndpointDescription" {
		t.Errorf("Expected EndpointDescription to be %s, got %s", "Test_EndpointDescription", data.EndpointDescription)
	}
	// AffiliationLegalBusinessName
	if data.AffiliationLegalBusinessName != "Test_AffiliationLegalBusinessName" {
		t.Errorf("Expected AffiliationLegalBusinessName to be %s, got %s", "Test_AffiliationLegalBusinessName", data.AffiliationLegalBusinessName)
	}
	// UseCode
	if data.UseCode != "Test_UseCode" {
		t.Errorf("Expected UseCode to be %s, got %s", "Test_UseCode", data.UseCode)
	}
	// UseDescription
	if data.UseDescription != "Test_UseDescription" {
		t.Errorf("Expected UseDescription to be %s, got %s", "Test_UseDescription", data.UseDescription)
	}
	// OtherUseDescription
	if data.OtherUseDescription != "Test_OtherUseDescription" {
		t.Errorf("Expected OtherUseDescription to be %s, got %s", "Test_OtherUseDescription", data.OtherUseDescription)
	}
	// ContentType
	if data.ContentType != "Test_ContentType" {
		t.Errorf("Expected ContentType to be %s, got %s", "Test_ContentType", data.ContentType)
	}
	// ContentDescription
	if data.ContentDescription != "Test_ContentDescription" {
		t.Errorf("Expected ContentDescription to be %s, got %s", "Test_ContentDescription", data.ContentDescription)
	}
	// OtherContentDescription
	if data.OtherContentDescription != "Test_OtherContentDescription" {
		t.Errorf("Expected OtherContentDescription to be %s, got %s", "Test_OtherContentDescription", data.OtherContentDescription)
	}
	// Affiliation_Address_Line_One
	if data.Affiliation_Address_Line_One != "30454 E Rustic Drive" {
		t.Errorf("Expected Affiliation_Address_Line_One to be %s, got %s", "30454 E Rustic Drive", data.Affiliation_Address_Line_One)
	}
	// Affiliation_Address_Line_Two
	if data.Affiliation_Address_Line_Two != "Test_Address_Line_Two" {
		t.Errorf("Expected Affiliation_Address_Line_Two to be %s, got %s", "", data.Affiliation_Address_Line_Two)
	}
	// Affiliation_Address_City
	if data.Affiliation_Address_City != "Salisbury" {
		t.Errorf("Expected Affiliation_Address_City to be %s, got %s", "Salisbury", data.Affiliation_Address_City)
	}
	// Affiliation_Address_State
	if data.Affiliation_Address_State != "MD" {
		t.Errorf("Expected Affiliation_Address_State to be %s, got %s", "MD", data.Affiliation_Address_State)
	}
	// Affiliation_Address_Country
	if data.Affiliation_Address_Country != "US" {
		t.Errorf("Expected Affiliation_Address_Country to be %s, got %s", "US", data.Affiliation_Address_Country)
	}
	// Affiliation_Address_Postal_Code
	if data.Affiliation_Address_Postal_Code != "21804" {
		t.Errorf("Expected Affiliation_Address_Postal_Code to be %s, got %s", "21804", data.Affiliation_Address_Postal_Code)
	}

}

func Test_ReadContactCSVContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	lines, err := readContactCsv(ctx, "testdata/npi_contact_file.csv")
	th.Assert(t, errors.Cause(err) == context.Canceled, "Expected canceled context error")
	th.Assert(t, lines == nil, "expected lines returned after error to be nil")
}

func Test_IsValidURL(t *testing.T) {
	th.Assert(t, isValidURL("https://www.foobar.com"), "expected https://www.foobar.com to be a valid URL")
	th.Assert(t, isValidURL("http://www.foobar.com"), "expected http://www.foobar.com to be a valid URL")
	th.Assert(t, isValidURL("HTTPS://www.foobar.com"), "expected HTTPS://www.foobar.com to be a valid URL")
	th.Assert(t, isValidURL("HTTP://www.foobar.com"), "expected HTTP://www.foobar.com to be a valid URL")
	th.Assert(t, isValidURL("www.foobar.com"), "expected www.foobar.com to be a valid URL")
	th.Assert(t, isValidURL("foobar.com"), "expected foobar.com to be a valid URL")
	th.Assert(t, isValidURL("www.foobar.org"), "expected www.foobar.orgto be a valid URL")
	th.Assert(t, isValidURL("foobar.org"), "expected foobar.org to be a valid URL")
	th.Assert(t, !isValidURL("foobar"), "expected foobar to not be a valid URL")
	th.Assert(t, !isValidURL("test.test@foobar.org"), "expected test.test@foobar.org to not be a valid URL")
	th.Assert(t, !isValidURL("test.test@foobar.com"), "expected test.test@foobar.com to not be a valid URL")
}

func Test_BuildNPIContactFromNPICsvLine(t *testing.T) {
	ctx := context.Background()

	// fixture file generated by running `head -20` on the provider organization .csv downloaded from http://download.cms.gov/nppes/NPI_Files.html
	// data was added to all columns in line 2, some columns have nested quotation marks to replicate the errors that exist in the NPI Download
	lines, err := readContactCsv(ctx, "testdata/npi_contact_file.csv")
	if err != nil {
		t.Errorf("Error reading NPI data from fixture file")
	}

	data := parseNPIContactdataLine(lines[2])
	npi_contact := buildNPIContactFromNPICsvLine(data)
	// NPI Field
	if npi_contact.NPI_ID != "1427051648" {
		t.Errorf("Expected NPI_ID to be %s, got %s", "1427051648", npi_contact.NPI_ID)
	}
	// EndpointType
	if npi_contact.EndpointType != "OTHERS" {
		t.Errorf("Expected EndpointType to be %s, got %s", "OTHERS", npi_contact.EndpointType)
	}
	// EndpointTypeDescription
	if npi_contact.EndpointTypeDescription != "Other URL" {
		t.Errorf("Expected EndpointTypeDescription to be %s, got %s", "Other URL", npi_contact.EndpointTypeDescription)
	}
	// Endpoint, also tests that the /metadata was stripped off url
	if npi_contact.Endpoint != "https://foobar.com/" {
		t.Errorf("Expected Endpoint to be %s, got %s", "https://foobar.com/", npi_contact.Endpoint)
	}
	// Affiliation
	if npi_contact.Affiliation != "N" {
		t.Errorf("Expected Affiliation to be %s, got %s", "N", npi_contact.Affiliation)
	}
	// EndpointDescription
	if npi_contact.EndpointDescription != "Test_EndpointDescription" {
		t.Errorf("Expected EndpointDescription to be %s, got %s", "Test_EndpointDescription", npi_contact.EndpointDescription)
	}
	// AffiliationLegalBusinessName
	if npi_contact.AffiliationLegalBusinessName != "Test_AffiliationLegalBusinessName" {
		t.Errorf("Expected AffiliationLegalBusinessName to be %s, got %s", "Test_AffiliationLegalBusinessName", npi_contact.AffiliationLegalBusinessName)
	}
	// UseCode
	if npi_contact.UseCode != "Test_UseCode" {
		t.Errorf("Expected UseCode to be %s, got %s", "Test_UseCode", npi_contact.UseCode)
	}
	// UseDescription
	if npi_contact.UseDescription != "Test_UseDescription" {
		t.Errorf("Expected UseDescription to be %s, got %s", "Test_UseDescription", npi_contact.UseDescription)
	}
	// OtherUseDescription
	if npi_contact.OtherUseDescription != "Test_OtherUseDescription" {
		t.Errorf("Expected OtherUseDescription to be %s, got %s", "Test_OtherUseDescription", npi_contact.OtherUseDescription)
	}
	// ContentType
	if npi_contact.ContentType != "Test_ContentType" {
		t.Errorf("Expected ContentType to be %s, got %s", "Test_ContentType", npi_contact.ContentType)
	}
	// ContentDescription
	if npi_contact.ContentDescription != "Test_ContentDescription" {
		t.Errorf("Expected ContentDescription to be %s, got %s", "Test_ContentDescription", npi_contact.ContentDescription)
	}
	// OtherContentDescription
	if npi_contact.OtherContentDescription != "Test_OtherContentDescription" {
		t.Errorf("Expected OtherContentDescription to be %s, got %s", "Test_OtherContentDescription", npi_contact.OtherContentDescription)
	}
	// Location.Address1
	if npi_contact.Location.Address1 != "30454 E Rustic Drive" {
		t.Errorf("Expected Affiliation_Address_Line_One to be %s, got %s", "30454 E Rustic Drive", npi_contact.Location.Address1)
	}
	// Location.Address2
	if npi_contact.Location.Address2 != "Test_Address_Line_Two" {
		t.Errorf("Expected Affiliation_Address_Line_Two to be %s, got %s", "", npi_contact.Location.Address2)
	}
	// Location.City
	if npi_contact.Location.City != "Salisbury" {
		t.Errorf("Expected Affiliation_Address_City to be %s, got %s", "Salisbury", npi_contact.Location.City)
	}
	// Location.State
	if npi_contact.Location.State != "MD" {
		t.Errorf("Expected Affiliation_Address_State to be %s, got %s", "MD", npi_contact.Location.State)
	}
	// Location.ZipCode
	if npi_contact.Location.ZipCode != "21804" {
		t.Errorf("Expected Affiliation_Address_Postal_Code to be %s, got %s", "21804", npi_contact.Location.ZipCode)
	}
}
