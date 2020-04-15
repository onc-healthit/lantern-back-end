package endpointmanager

import (
	"context"
)

// NPIContact represents the digitial contact information for an NPI Contact provided by the NPPES database
type NPIContact struct {
	ID                      int
	NPI_ID								string
	Endpoint_Type						string
	Endpoint_Type_Description			string
	Endpoint							string
	Valid_URL							bool
	Affiliation							string
	Endpoint_Description				string
	Affiliation_Legal_Business_Name		string
	Normalized_Affiliation_Legal_Business_Name string
	Use_Code							string
	Use_Description						string
	Other_Use_Description				string
	Content_Type						string
	Content_Description					string
	Other_Content_Description			string
	Location                			*Location
}

// NPIContactStore is the interface for interacting with the storage layer that holds
// NPIContact objects.
type NPIContactStore interface {
	GetNPIContact(context.Context, int) (*NPIContact, error)
	GetNPIContactByNPIID(context.Context, string) (*NPIContact, error)
	DeleteAllNPIContacts(context.Context) error
	AddNPIContact(context.Context, *NPIContact) error
	UpdateNPIContact(context.Context, *NPIContact) error
	UpdateNPIContactByNPIID(context.Context, *NPIContact) error
	DeleteNPIContact(context.Context, *NPIContact) error
	Close()
}

