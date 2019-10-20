package main

// Location represents a US postal address.
type Location struct {
	Address1 string // the first line of the address. For example, "123 Governors Ln".
	Address2 string // the second line of the address, if it exists. For example, "Suite 123".
	Address3 string // the third line of the address, if it exists. For example, "MailStop 123".
	City     string
	State    string // the two-letter state or posession abbreviation as defined in https://pe.usps.com/text/pub28/28apb.htm.
	ZipCode  string // the five-digit zip code.
}

// Equal checks if the location is equal to the given location.
func (l *Location) Equal(l2 *Location) bool {
	if l == nil && l2 == nil {
		return true
	} else if l == nil {
		return false
	} else if l2 == nil {
		return false
	}

	if l.Address1 != l2.Address1 {
		return false
	}
	if l.Address2 != l2.Address2 {
		return false
	}
	if l.Address3 != l2.Address3 {
		return false
	}
	if l.City != l2.City {
		return false
	}
	if l.State != l2.State {
		return false
	}
	if l.ZipCode != l2.ZipCode {
		return false
	}

	return true
}
