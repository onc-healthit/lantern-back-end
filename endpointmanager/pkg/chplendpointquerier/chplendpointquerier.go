package chplendpointquerier

type EndpointList struct {
	Endpoints []LanternEntry `json:"Endpoints"`
}

type LanternEntry struct {
	URL              string `json:"URL"`
	OrganizationName string `json:"OrganizationName"`
	NPIID            string `json:"NPIID"`
}

var MedHostURL = "https://api.mhdi10xasayd.com/medhost-developer-composition/v1/fhir-base-urls.json"
var NextGenURL = "https://nextgen.com/api/practice-search"
var CanvasURL = "https://docs.canvasmedical.com/reference/service-base-urls"
var AllScriptsURL = "https://open.allscripts.com/fhirendpoints"
var EpicURL = "https://open.epic.com/MyApps/Endpoints"
var MeditechURL = "https://home.meditech.com/en/d/restapiresources/pages/apidoc.htm"

func QueryCHPLEndpointList(chplURL string, fileToWriteTo string) {

	if chplURL == MedHostURL {
		MedHostQuerier(chplURL, fileToWriteTo)
	} else if chplURL == NextGenURL {
		NextGenwebscraper(chplURL, fileToWriteTo)
	} else if chplURL == CanvasURL {
		Canvaswebscraper(chplURL, fileToWriteTo)
	} else if chplURL == AllScriptsURL {
		AllScriptsQuerier(chplURL, fileToWriteTo)
	} else if chplURL == EpicURL {
		EpicQuerier(chplURL, fileToWriteTo)
	} else if chplURL == MeditechURL {
		Meditechwebscraper("https://fhir.meditech.com/explorer/endpoints", fileToWriteTo)
	}
}
