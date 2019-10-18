package main

// Location represents a US postal address.
type Location struct {
	// Address1 is the first line of the address. For example, "123 Governors Ln".
	Address1 string
	// Address2 is the second line of the address, if it exists. For example, "Suite 123".
	Address2 string
	// Address3 is the third line of the address, if it exists. For example, "MailStop 123".
	Address3 string
	City     string
	// State is the two-letter state or posession abbreviation as defined in https://pe.usps.com/text/pub28/28apb.htm.
	State string
	// ZipCode is the five-digit zip code.
	ZipCode string
}
