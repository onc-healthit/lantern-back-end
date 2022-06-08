package main

import (
	"encoding/json"
	"io/ioutil"
	http "net/http"
	"os"
	"strings"
	"time"
	log "github.com/sirupsen/logrus"
)

type softwareInfo struct {
	ListSourceURL    string                             `json:"listSourceURL"`
	SoftwareProducts []chplCertifiedProductEntry        `json:"softwareProducts"`
}

type endpointEntry struct {
	FormatType   string `json:"FormatType"`
	URL          string `json:"URL"`
	EndpointName string `json:"EndpointName"`
	FileName     string `json:"FileName"`
}

type details struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type certCriteria struct {
	ID     int    `json:"id"`
	Number string `json:"number"`
	Title  string `json:"title"`
}

type serviceBaseURL struct {
	Criterion certCriteria `json:"criterion"`
	Value     string       `json:"value"`
}

type CHPLEndpointList struct {
	Results []CHPLEndpointEntry `json:"results"`
}

type CHPLEndpointEntry struct {
	Developer           details `json:"developer"`
	Product             details `json:"product"`
	Version             details `json:"version"`
	CertificationStatus details `json:"certificationStatus"`
	CertificationDate   string
	Edition             details          `json:"edition"`
	CHPLProductNumber   string           `json:"chplProductNumber"`
	CriteriaMet         []certCriteria   `json:"criteriaMet"`
	ServiceBaseUrlList  serviceBaseURL   `json:"serviceBaseUrlList"`
	APIDocumentation    []serviceBaseURL `json:"apiDocumentation"`
}

type chplCertifiedProductEntry struct {
	ID                  int      `json:"id"`
	ChplProductNumber   string   `json:"chplProductNumber"`
	Edition             string   `json:"edition"`
	PracticeType        string   `json:"practiceType"`
	Developer           string   `json:"developer"`
	Product             string   `json:"product"`
	Version             string   `json:"version"`
	CertificationDate   int64    `json:"certificationDate"`
	CertificationStatus string   `json:"certificationStatus"`
	CriteriaMet         []int    `json:"criteriaMet"`
	APIDocumentation    []string `json:"apiDocumentation"`
}

func main() {

	var chplURL string
	var fileToWriteToCHPLList string
	fileToWriteToSoftwareInfo := "CHPLProductsInfo.json"

	if len(os.Args) >= 1 {
		chplURL = os.Args[1]
		fileToWriteToCHPLList = os.Args[2]
	} else {
		log.Fatalf("ERROR: Missing command-line arguments")
	}

	var endpointEntryList []endpointEntry
	var softwareInfoList []softwareInfo

	client := &http.Client{}
	req, err := http.NewRequest("GET", chplURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var chplJSON CHPLEndpointList
	err = json.Unmarshal(respBody, &chplJSON)
	if err != nil {
		log.Fatal(err)
	}

	chplResultsList := chplJSON.Results
	if chplResultsList == nil {
		log.Fatal("CHPL endpoint list is empty")
	}

	for _, chplEntry := range chplResultsList {

		developerName := chplEntry.Developer.Name
		developerName = strings.TrimSpace(developerName)

		productNumber := chplEntry.CHPLProductNumber
		productNumber = strings.TrimSpace(productNumber)

		productName := chplEntry.Product.Name
		productName = strings.TrimSpace(productName)

		productVersion := chplEntry.Version.Name
		productVersion = strings.TrimSpace(productVersion)

		productCertStatus := chplEntry.CertificationStatus.Name
		productCertStatus = strings.TrimSpace(productCertStatus)

		productEdition := chplEntry.Edition.Name
		productEdition = strings.TrimSpace(productEdition)

		certificationDateTime, err := time.Parse("2006-01-02", chplEntry.CertificationDate)
		if err != nil {
			log.Fatal("converting certification date to time failed")
		}
		certificationDateInt := certificationDateTime.Unix()

		var criteriaMetArr []int
		for _, criteriaEntry := range chplEntry.CriteriaMet {
			criteriaMetArr = append(criteriaMetArr, criteriaEntry.ID)
		}

		var apiDocURLArr []string
		for _, apiURLEntry := range chplEntry.APIDocumentation {
			apiDocURLArr = append(apiDocURLArr, apiURLEntry.Value)
		}

		var entry endpointEntry

		urlString := chplEntry.ServiceBaseUrlList.Value
		urlString = strings.TrimSpace(urlString)

		var productEntry chplCertifiedProductEntry

		productEntry.Product = productName
		productEntry.ChplProductNumber = productNumber
		productEntry.Version = productVersion
		productEntry.CertificationStatus = productCertStatus
		productEntry.CertificationDate = certificationDateInt
		productEntry.Edition = productEdition
		productEntry.CriteriaMet = criteriaMetArr
		productEntry.APIDocumentation = apiDocURLArr
		productEntry.Developer = developerName

		softwareContained, softwareIndex := containsSoftware(softwareInfoList, urlString)
		if !softwareContained {
			var softwareInfoEntry softwareInfo
			softwareInfoEntry.ListSourceURL = urlString
			softwareInfoEntry.SoftwareProducts = append(softwareInfoEntry.SoftwareProducts, productEntry)
			softwareInfoList = append(softwareInfoList, softwareInfoEntry)
		} else {
			softwareInfoList[softwareIndex].SoftwareProducts = append(softwareInfoList[softwareIndex].SoftwareProducts, productEntry)
		}

		if !containsEndpoint(endpointEntryList, urlString) {

			entry.URL = urlString

			entry.EndpointName = developerName

			// Get fileName from URL domain name
			fileName := urlString
			if strings.Count(urlString, ".") > 1 {
				index := strings.Index(urlString, ".")
				fileName = urlString[index+1:]
			} else {
				index := strings.Index(urlString, "://")
				fileName = urlString[index+3:]
			}

			index := strings.Index(fileName, ".")
			fileName = fileName[:index]

			entry.FileName = fileName + "EndpointSources.json"
			entry.FormatType = "Lantern"

			endpointEntryList = append(endpointEntryList, entry)
		}

	}

	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteToCHPLList, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

	finalFormatJSONSoftware, err := json.MarshalIndent(softwareInfoList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteToSoftwareInfo, finalFormatJSONSoftware, 0644)
	if err != nil {
		log.Fatal(err)
	}

}

func containsEndpoint(endpointEntryList []endpointEntry, url string) bool {
	for _, e := range endpointEntryList {
		if e.URL == url {
			return true
		}
	}
	return false
}

func containsSoftware(softwareProductList []softwareInfo, url string) (bool, int) {
	for index, e := range softwareProductList {
		if e.ListSourceURL == url {
			return true, index
		}
	}
	return false, -1
}
