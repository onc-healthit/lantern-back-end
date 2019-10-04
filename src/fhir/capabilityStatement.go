package fhir

import "encoding/xml"

type CapabilityStatement struct {
	XMLName xml.Name `xml:"Conformance"`
	Text    string   `xml:",chardata"`
	Xmlns   string   `xml:"xmlns,attr"`
	ID      struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"id"`
	URL struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"url"`
	Version struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"version"`
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
	Copyright struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"copyright"`
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
		Version struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"version"`
		ReleaseDate struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"releaseDate"`
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
		Security struct {
			Text string `xml:",chardata"`
			Cors struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"cors"`
			Service []struct {
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
			Extension struct {
				Text      string `xml:",chardata"`
				URL       string `xml:"url,attr"`
				Extension []struct {
					Text     string `xml:",chardata"`
					URL      string `xml:"url,attr"`
					ValueUri struct {
						Text  string `xml:",chardata"`
						Value string `xml:"value,attr"`
					} `xml:"valueUri"`
				} `xml:"extension"`
			} `xml:"extension"`
		} `xml:"security"`
		Resource []struct {
			Text string `xml:",chardata"`
			Type struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"type"`
			ReadHistory struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"readHistory"`
			UpdateCreate struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"updateCreate"`
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
			Interaction []struct {
				Text string `xml:",chardata"`
				Code struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"code"`
			} `xml:"interaction"`
			SearchParam []struct {
				Text string `xml:",chardata"`
				Name struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"name"`
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
	} `xml:"rest"`
}
