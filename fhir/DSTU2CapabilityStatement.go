package fhir

import "encoding/xml"

type DSTU2CapabilityStatement struct {
	XMLName  xml.Name `xml:"Conformance"`
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
			P     []struct {
				Text string `xml:",chardata"`
				A    struct {
					Text string `xml:",chardata"`
					Href string `xml:"href,attr"`
				} `xml:"a"`
			} `xml:"p"`
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
	Publisher struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"publisher"`
	Contact struct {
		Text string `xml:",chardata"`
		Name struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"name"`
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
	Date struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"date"`
	Description struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"description"`
	Requirements struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"requirements"`
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
	Implementation struct {
		Text        string `xml:",chardata"`
		Description struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"description"`
		URL struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"url"`
	} `xml:"implementation"`
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
				Text   string `xml:",chardata"`
				Coding struct {
					Text   string `xml:",chardata"`
					System struct {
						Text  string `xml:",chardata"`
						Value string `xml:"value,attr"`
					} `xml:"system"`
					Code struct {
						Text  string `xml:",chardata"`
						Value string `xml:"value,attr"`
					} `xml:"code"`
				} `xml:"coding"`
			} `xml:"service"`
			Description struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"description"`
			Certificate struct {
				Text string `xml:",chardata"`
				Type struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"type"`
				Blob struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"blob"`
			} `xml:"certificate"`
		} `xml:"security"`
		Resource struct {
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
			Versioning struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"versioning"`
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
			SearchInclude struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"searchInclude"`
			SearchRevInclude struct {
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
				Modifier struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"modifier"`
				Target struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"target"`
				Chain []struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"chain"`
			} `xml:"searchParam"`
		} `xml:"resource"`
		Interaction []struct {
			Text string `xml:",chardata"`
			Code struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"code"`
		} `xml:"interaction"`
		Compartment struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"compartment"`
	} `xml:"rest"`
	Messaging struct {
		Text     string `xml:",chardata"`
		Endpoint struct {
			Text     string `xml:",chardata"`
			Protocol struct {
				Text   string `xml:",chardata"`
				System struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"system"`
				Code struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"code"`
			} `xml:"protocol"`
			Address struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"address"`
		} `xml:"endpoint"`
		ReliableCache struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"reliableCache"`
		Documentation struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"documentation"`
		Event struct {
			Text string `xml:",chardata"`
			Code struct {
				Text   string `xml:",chardata"`
				System struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"system"`
				Code struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"code"`
			} `xml:"code"`
			Category struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"category"`
			Mode struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"mode"`
			Focus struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"focus"`
			Request struct {
				Text      string `xml:",chardata"`
				Reference struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"reference"`
			} `xml:"request"`
			Response struct {
				Text      string `xml:",chardata"`
				Reference struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"reference"`
			} `xml:"response"`
			Documentation struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"documentation"`
		} `xml:"event"`
	} `xml:"messaging"`
	Document struct {
		Text string `xml:",chardata"`
		Mode struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"mode"`
		Documentation struct {
			Text  string `xml:",chardata"`
			Value string `xml:"value,attr"`
		} `xml:"documentation"`
		Profile struct {
			Text      string `xml:",chardata"`
			Reference struct {
				Text  string `xml:",chardata"`
				Value string `xml:"value,attr"`
			} `xml:"reference"`
		} `xml:"profile"`
	} `xml:"document"`
}
