package fhir

import "encoding/xml"

type DSTU3CapabilityStatement struct {
	XMLName  xml.Name `xml:"CapabilityStatement"`
	Chardata string   `xml:",chardata"`
	Xmlns    string   `xml:"xmlns,attr"`
	ID       struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"id"`
	Text struct {
		Text   string `xml:",chardata"`
		Status struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"status"`
		Div struct {
			Text  string `xml:",chardata"`
			Xmlns string `xml:"xmlns,attr"`
			H2    string `xml:"h2"`
			Div   struct {
				Text string `xml:",chardata"`
				P    string `xml:"p"`
			} `xml:"div"`
			Table []struct {
				Text string `xml:",chardata"`
				Tr   []struct {
					Text string `xml:",chardata"`
					Td   []struct {
						Text string `xml:",chardata"`
						A    struct {
							Text string `xml:",chardata"`
							Href string `xml:"href,attr"`
						} `xml:"a"`
					} `xml:"td"`
					Th []struct {
						Text string `xml:",chardata"`
						B    string `xml:"b"`
					} `xml:"th"`
				} `xml:"tr"`
			} `xml:"table"`
		} `xml:"div"`
	} `xml:"text"`
	URL struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"url"`
	Version struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"version"`
	Name struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"name"`
	Status struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"status"`
	Experimental struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"experimental"`
	Date struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"date"`
	Publisher struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"publisher"`
	Contact struct {
		Text    string `xml:",chardata"`
		Telecom struct {
			Text   string `xml:",chardata"`
			System struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"system"`
			Value struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"value"`
		} `xml:"telecom"`
	} `xml:"contact"`
	Description struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"description"`
	Kind struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"kind"`
	Software struct {
		Text string `xml:",chardata"`
		Name struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"name"`
	} `xml:"software"`
	FhirVersion struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"fhirVersion"`
	AcceptUnknown struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"acceptUnknown"`
	Format []struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"format"`
	Rest struct {
		Text string `xml:",chardata"`
		Mode struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"mode"`
		Documentation struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"documentation"`
		Security struct {
			Text string `xml:",chardata"`
			Cors struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"cors"`
			Service struct {
				Chardata string `xml:",chardata"`
				Coding   struct {
					Text   string `xml:",chardata"`
					System struct {
						Text  string `xml:",chardata"`
						Value string `xml:"value,attr"`
					} `xml:"system"`
					Code struct {
						Text  string `xml:",chardata"`
						Value string `xml:"value,attr"`
					} `xml:"code"`
					Display struct {
						Text  string `xml:",chardata"`
						Value string `xml:"value,attr"`
					} `xml:"display"`
				} `xml:"coding"`
				Text struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"text"`
			} `xml:"service"`
			Description struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"description"`
		} `xml:"security"`
		Resource []struct {
			Text string `xml:",chardata"`
			Type struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"type"`
			Profile struct {
				Text      string `xml:",chardata"`
				Reference struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"reference"`
			} `xml:"profile"`
			Interaction []struct {
				Text string `xml:",chardata"`
				Code struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"code"`
				Documentation struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"documentation"`
			} `xml:"interaction"`
			ConditionalCreate struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"conditionalCreate"`
			ConditionalUpdate struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"conditionalUpdate"`
			ConditionalDelete struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"conditionalDelete"`
			ReferencePolicy []struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"referencePolicy"`
			SearchInclude []struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"searchInclude"`
			SearchRevInclude []struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"searchRevInclude"`
			SearchParam []struct {
				Text string `xml:",chardata"`
				Name struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"name"`
				Definition struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"definition"`
				Type struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"type"`
				Documentation struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"documentation"`
			} `xml:"searchParam"`
		} `xml:"resource"`
		Interaction []struct {
			Text string `xml:",chardata"`
			Code struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"code"`
			Documentation struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"documentation"`
		} `xml:"interaction"`
		SearchParam []struct {
			Text string `xml:",chardata"`
			Name struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"name"`
			Definition struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"definition"`
			Type struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"type"`
			Documentation struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"documentation"`
		} `xml:"searchParam"`
		Operation []struct {
			Text string `xml:",chardata"`
			Name struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"name"`
			Definition struct {
				Text      string `xml:",chardata"`
				Reference struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"reference"`
			} `xml:"definition"`
		} `xml:"operation"`
	} `xml:"rest"`
}
