package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	http "net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type softwareInfo struct {
	ListSourceURL    string                      `json:"listSourceURL"`
	SoftwareProducts []chplCertifiedProductEntry `json:"softwareProducts"`
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
	PageSize    int                 `json:"pageSize"`
	PageNumber  int                 `json:"pageNumber"`
	RecordCount int                 `json:"recordCount"`
	Results     []CHPLEndpointEntry `json:"results"`
}

type CHPLEndpointEntry struct {
	ID                  int              `json:"id"`
	Developer           details          `json:"developer"`
	Product             details          `json:"product"`
	Version             details          `json:"version"`
	CertificationStatus details          `json:"certificationStatus"`
	CertificationDate   string           `json:"certificationDate"`
	Edition             details          `json:"edition"`
	CHPLProductNumber   string           `json:"chplProductNumber"`
	CriteriaMet         []certCriteria   `json:"criteriaMet"`
	ServiceBaseUrlList  serviceBaseURL   `json:"serviceBaseUrlList"`
	APIDocumentation    []serviceBaseURL `json:"apiDocumentation"`
	ACB                 details          `json:"certificationBody"`
}

type chplCertifiedProductEntry struct {
	ID                  int              `json:"id"`
	ChplProductNumber   string           `json:"chplProductNumber"`
	Edition             details          `json:"edition"`
	PracticeType        details          `json:"practiceType"`
	Developer           details          `json:"developer"`
	Product             details          `json:"product"`
	Version             details          `json:"version"`
	CertificationDate   string           `json:"certificationDate"`
	CertificationStatus details          `json:"certificationStatus"`
	CriteriaMet         []certCriteria   `json:"criteriaMet"`
	APIDocumentation    []serviceBaseURL `json:"apiDocumentation"`
	ACB                 string           `json:"acb"`
}

func main() {
	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 60,
	}

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

	pageSize := 100
	pageNumber := 0
	savedEntries := 0

	for {
		respBody, err := getEndpointListJSON(chplURL, pageSize, pageNumber, ctx, client)
		if err != nil {
			log.Fatal(err)
		}

		var chplJSON CHPLEndpointList
		err = json.Unmarshal(respBody, &chplJSON)
		if err != nil {
			log.Fatal(err)
		}

		if savedEntries >= chplJSON.RecordCount {
			break
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

			certificationDateTime := chplEntry.CertificationDate

			criteriaMetArr := chplEntry.CriteriaMet

			apiDocURLArr := chplEntry.APIDocumentation

			var entry endpointEntry

			urlString := chplEntry.ServiceBaseUrlList.Value
			urlString = strings.TrimSpace(urlString)

			var productEntry chplCertifiedProductEntry

			productEntry.ID = chplEntry.ID
			productEntry.Product = chplEntry.Product
			productEntry.ChplProductNumber = productNumber
			productEntry.Version = chplEntry.Version
			productEntry.CertificationStatus = chplEntry.CertificationStatus
			productEntry.CertificationDate = certificationDateTime
			productEntry.Edition = chplEntry.Edition
			productEntry.CriteriaMet = criteriaMetArr
			productEntry.APIDocumentation = apiDocURLArr
			productEntry.Developer = chplEntry.Developer
			productEntry.ACB = chplEntry.ACB.Name

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

				// Get fileName from developer name
				re, err := regexp.Compile(`[^\w\s\']|_`)
				if err != nil {
					log.Fatal(err)
				}

				developerNameNormalized := re.ReplaceAllString(developerName, "")
				fileNameArr := strings.Fields(developerNameNormalized)
				fileName := ""
				if len(fileNameArr) > 0 {
					for _, s := range fileNameArr {
						fileName = fileName + s + "_"
					}
				} else {
					fileName = "Unknown_Developer_"
				}

				matchedFiles := containsFileName(endpointEntryList, fileName)
				// Ensure we do not have any file names that are the same
				if matchedFiles > 0 {
					fileName = fileName + strconv.Itoa(matchedFiles) + "_"
				}

				entry.FileName = fileName + "EndpointSources.json"
				entry.FormatType = "Lantern"

				endpointEntryList = append(endpointEntryList, entry)
			}

		}

		pageNumber = pageNumber + 1
		savedEntries = savedEntries + len(chplJSON.Results)
	}

	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteToCHPLList, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Save a copy of CHPL Endpoint Lists file in the dev resources folder
	devfinalFormatJSONEndpoints, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/dev_resources/"+fileToWriteToCHPLList, devfinalFormatJSONEndpoints, 0644)
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

	// Save a copy of software products info file in the dev resources folder
	devfinalFormatJSONSoftware, err := json.MarshalIndent(softwareInfoList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/dev_resources/"+fileToWriteToSoftwareInfo, devfinalFormatJSONSoftware, 0644)
	if err != nil {
		log.Fatal(err)
	}

}

func getEndpointListJSON(chplURL string, pageSize int, pageNumber int, ctx context.Context, client *http.Client) ([]byte, error) {

	chplURL = chplURL + "&pageSize=" + strconv.Itoa(pageSize) + "&pageNumber=" + strconv.Itoa(pageNumber)

	req, err := http.NewRequest("GET", chplURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req = req.WithContext(ctx)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func containsEndpoint(endpointEntryList []endpointEntry, url string) bool {
	for _, e := range endpointEntryList {
		if e.URL == url {
			return true
		}
	}
	return false
}

func containsFileName(endpointEntryList []endpointEntry, filename string) int {
	matchedFiles := 0
	for _, e := range endpointEntryList {
		if strings.Contains(e.FileName, filename) {
			matchedFiles = matchedFiles + 1
		}
	}
	return matchedFiles
}

func containsSoftware(softwareProductList []softwareInfo, url string) (bool, int) {
	for index, e := range softwareProductList {
		if e.ListSourceURL == url {
			return true, index
		}
	}
	return false, -1
}
