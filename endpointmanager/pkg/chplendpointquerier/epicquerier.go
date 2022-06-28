package chplendpointquerier

import (
	"encoding/json"
	"io/ioutil"
	http "net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func EpicQuerier(epicURL string, fileToWriteTo string) {

	DSTU2URL := strings.Join(strings.Split(epicURL, "/")[:3], "/") + "/Endpoints/DSTU2"
	R4URL := strings.Join(strings.Split(epicURL, "/")[:3], "/") + "/Endpoints/R4"

	var endpointEntryList EndpointList

	client := &http.Client{}
	req, err := http.NewRequest("GET", DSTU2URL, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	endpointEntryList.Endpoints = BundleToLanternFormat(respBody)

	client = &http.Client{}
	req, err = http.NewRequest("GET", R4URL, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	respBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, BundleToLanternFormat(respBody)...)

	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

}