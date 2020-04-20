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
	// "1427051648","OTHERS","Other URL","Salisbury","N","Test_Endpoint_Description","Test_Affiliation_Legal_Business_Name","Test_Use_Code","Test_Use_Description","Test_Other_Use_Description","Test_Content_Type","Test_Content_Description","Test_Other_Content_Description","30454 E Rustic Drive","Test_Address_Line_Two","Salisbury","MD","US","21804"
	data := parseNPIContactdataLine(lines[2])
	// NPI
	if data.NPI != "1427051648" {
		t.Errorf("Expected NPI to be %s, got %s", "1427051648", data.NPI)
	}
	// Endpoint_Type
	if data.Endpoint_Type != "OTHERS" {
		t.Errorf("Expected Endpoint_Type to be %s, got %s", "OTHERS", data.Endpoint_Type)
	}
	// Endpoint_Type_Description
	if data.Endpoint_Type_Description != "Other URL" {
		t.Errorf("Expected Endpoint_Type_Description to be %s, got %s", "Other URL", data.Endpoint_Type_Description)
	}
	// Endpoint
	if data.Endpoint != "Salisbury" {
		t.Errorf("Expected Endpoint to be %s, got %s", "Salisbury", data.Endpoint)
	}
	// Affiliation
	if data.Affiliation != "N" {
		t.Errorf("Expected Affiliation to be %s, got %s", "N", data.Affiliation)
	}
	// Endpoint_Description
	if data.Endpoint_Description != "Test_Endpoint_Description" {
		t.Errorf("Expected Endpoint_Description to be %s, got %s", "Test_Endpoint_Description", data.Endpoint_Description)
	}
	// Affiliation_Legal_Business_Name
	if data.Affiliation_Legal_Business_Name != "Test_Affiliation_Legal_Business_Name" {
		t.Errorf("Expected Affiliation_Legal_Business_Name to be %s, got %s", "Test_Affiliation_Legal_Business_Name", data.Affiliation_Legal_Business_Name)
	}
	// Use_Code
	if data.Use_Code != "Test_Use_Code" {
		t.Errorf("Expected Use_Code to be %s, got %s", "Test_Use_Code", data.Use_Code)
	}
	// Use_Description
	if data.Use_Description != "Test_Use_Description" {
		t.Errorf("Expected Use_Description to be %s, got %s", "Test_Use_Description", data. Use_Description)
	}
	// Other_Use_Description
	if data.Other_Use_Description != "Test_Other_Use_Description" {
		t.Errorf("Expected Other_Use_Description to be %s, got %s", "Test_Other_Use_Description", data.Other_Use_Description)
	}
	// Content_Type
	if data.Content_Type != "Test_Content_Type" {
		t.Errorf("Expected Content_Type to be %s, got %s", "Test_Content_Type", data.Content_Type)
	}
	// Content_Description
	if data.Content_Description != "Test_Content_Description" {
		t.Errorf("Expected Content_Description to be %s, got %s", "Test_Content_Description", data.Content_Description)
	}
	// Other_Content_Description
	if data.Other_Content_Description != "Test_Other_Content_Description" {
		t.Errorf("Expected Other_Content_Description to be %s, got %s", "Test_Other_Content_Description", data.Other_Content_Description)
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
	// Endpoint_Type
	if npi_contact.Endpoint_Type != "OTHERS" {
		t.Errorf("Expected Endpoint_Type to be %s, got %s", "OTHERS", npi_contact.Endpoint_Type)
	}
	// Endpoint_Type_Description
	if npi_contact.Endpoint_Type_Description != "Other URL" {
		t.Errorf("Expected Endpoint_Type_Description to be %s, got %s", "Other URL", npi_contact.Endpoint_Type_Description)
	}
	// Endpoint
	if npi_contact.Endpoint != "Salisbury" {
		t.Errorf("Expected Endpoint to be %s, got %s", "Salisbury", npi_contact.Endpoint)
	}
	// Affiliation
	if npi_contact.Affiliation != "N" {
		t.Errorf("Expected Affiliation to be %s, got %s", "N", npi_contact.Affiliation)
	}
	// Endpoint_Description
	if npi_contact.Endpoint_Description != "Test_Endpoint_Description" {
		t.Errorf("Expected Endpoint_Description to be %s, got %s", "Test_Endpoint_Description", npi_contact.Endpoint_Description)
	}
	// Affiliation_Legal_Business_Name
	if npi_contact.Affiliation_Legal_Business_Name != "Test_Affiliation_Legal_Business_Name" {
		t.Errorf("Expected Affiliation_Legal_Business_Name to be %s, got %s", "Test_Affiliation_Legal_Business_Name", npi_contact.Affiliation_Legal_Business_Name)
	}
	// Use_Code
	if npi_contact.Use_Code != "Test_Use_Code" {
		t.Errorf("Expected Use_Code to be %s, got %s", "Test_Use_Code", npi_contact.Use_Code)
	}
	// Use_Description
	if npi_contact.Use_Description != "Test_Use_Description" {
		t.Errorf("Expected Use_Description to be %s, got %s", "Test_Use_Description", npi_contact. Use_Description)
	}
	// Other_Use_Description
	if npi_contact.Other_Use_Description != "Test_Other_Use_Description" {
		t.Errorf("Expected Other_Use_Description to be %s, got %s", "Test_Other_Use_Description", npi_contact.Other_Use_Description)
	}
	// Content_Type
	if npi_contact.Content_Type != "Test_Content_Type" {
		t.Errorf("Expected Content_Type to be %s, got %s", "Test_Content_Type", npi_contact.Content_Type)
	}
	// Content_Description
	if npi_contact.Content_Description != "Test_Content_Description" {
		t.Errorf("Expected Content_Description to be %s, got %s", "Test_Content_Description", npi_contact.Content_Description)
	}
	// Other_Content_Description
	if npi_contact.Other_Content_Description != "Test_Other_Content_Description" {
		t.Errorf("Expected Other_Content_Description to be %s, got %s", "Test_Other_Content_Description", npi_contact.Other_Content_Description)
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
