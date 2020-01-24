package main

import (
	"bufio"
	"encoding/csv"
	"strings"
	"io"
	"log"
	"regexp"
	"os"
)

func failOnError(errString string, err error) {
	if err != nil {
		log.Fatalf("%s %s", errString, err)
	}
}

func main() {
	// Open the file
	csvfile, err := os.Open("npidata_pfile_20050523-20191110.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	// Parse the file
	r := csv.NewReader(bufio.NewReader(csvfile))

	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		entity_type_code := record[1]
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if(entity_type_code == "2"){
			// npi_number := record[0]
			legal_business_name := record[4]
			other_organizxation_name := record[11]
			//fmt.Printf("%s,%s\n",legal_business_name,other_organizxation_name)
			//normalizeOrgName(legal_business_name)
			//normalizeOrgName(other_organizxation_name)
		}
	}
}