package nppesquerier

import (
	"encoding/csv"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

// Provider organization .csv downloaded from http://download.cms.gov/nppes/NPI_Files.html
type NPICsvLine struct {
	NPI                                                                    string
	Entity_Type_Code                                                       string
	Replacement_NPI                                                        string
	Employer_Identification_Number_EIN                                     string
	Provider_Organization_Name_Legal_Business_Name                         string
	Provider_Last_Name_Legal_Name                                          string
	Provider_First_Name                                                    string
	Provider_Middle_Name                                                   string
	Provider_Name_Prefix_Text                                              string
	Provider_Name_Suffix_Text                                              string
	Provider_Credential_Text                                               string
	Provider_Other_Organization_Name                                       string
	Provider_Other_Organization_Name_Type_Code                             string
	Provider_Other_Last_Name                                               string
	Provider_Other_First_Name                                              string
	Provider_Other_Middle_Name                                             string
	Provider_Other_Name_Prefix_Text                                        string
	Provider_Other_Name_Suffix_Text                                        string
	Provider_Other_Credential_Text                                         string
	Provider_Other_Last_Name_Type_Code                                     string
	Provider_First_Line_Business_Mailing_Address                           string
	Provider_Second_Line_Business_Mailing_Address                          string
	Provider_Business_Mailing_Address_City_Name                            string
	Provider_Business_Mailing_Address_State_Name                           string
	Provider_Business_Mailing_Address_Postal_Code                          string
	Provider_Business_Mailing_Address_Country_Code_If_outside_US           string
	Provider_Business_Mailing_Address_Telephone_Number                     string
	Provider_Business_Mailing_Address_Fax_Number                           string
	Provider_First_Line_Business_Practice_Location_Address                 string
	Provider_Second_Line_Business_Practice_Location_Address                string
	Provider_Business_Practice_Location_Address_City_Name                  string
	Provider_Business_Practice_Location_Address_State_Name                 string
	Provider_Business_Practice_Location_Address_Postal_Code                string
	Provider_Business_Practice_Location_Address_Country_Code_If_outside_US string
	Provider_Business_Practice_Location_Address_Telephone_Number           string
	Provider_Business_Practice_Location_Address_Fax_Number                 string
	Provider_Enumeration_Date                                              string
	Last_Update_Date                                                       string
	NPI_Deactivation_Reason_Code                                           string
	NPI_Deactivation_Date                                                  string
	NPI_Reactivation_Date                                                  string
	Provider_Gender_Code                                                   string
	Authorized_Official_Last_Name                                          string
	Authorized_Official_First_Name                                         string
	Authorized_Official_Middle_Name                                        string
	Authorized_Official_Title_or_Position                                  string
	Authorized_Official_Telephone_Number                                   string
	Healthcare_Provider_Taxonomy_Code_1                                    string
	Provider_License_Number_1                                              string
	Provider_License_Number_State_Code_1                                   string
	Healthcare_Provider_Primary_Taxonomy_Switch_1                          string
	Healthcare_Provider_Taxonomy_Code_2                                    string
	Provider_License_Number_2                                              string
	Provider_License_Number_State_Code_2                                   string
	Healthcare_Provider_Primary_Taxonomy_Switch_2                          string
	Healthcare_Provider_Taxonomy_Code_3                                    string
	Provider_License_Number_3                                              string
	Provider_License_Number_State_Code_3                                   string
	Healthcare_Provider_Primary_Taxonomy_Switch_3                          string
	Healthcare_Provider_Taxonomy_Code_4                                    string
	Provider_License_Number_4                                              string
	Provider_License_Number_State_Code_4                                   string
	Healthcare_Provider_Primary_Taxonomy_Switch_4                          string
	Healthcare_Provider_Taxonomy_Code_5                                    string
	Provider_License_Number_5                                              string
	Provider_License_Number_State_Code_5                                   string
	Healthcare_Provider_Primary_Taxonomy_Switch_5                          string
	Healthcare_Provider_Taxonomy_Code_6                                    string
	Provider_License_Number_6                                              string
	Provider_License_Number_State_Code_6                                   string
	Healthcare_Provider_Primary_Taxonomy_Switch_6                          string
	Healthcare_Provider_Taxonomy_Code_7                                    string
	Provider_License_Number_7                                              string
	Provider_License_Number_State_Code_7                                   string
	Healthcare_Provider_Primary_Taxonomy_Switch_7                          string
	Healthcare_Provider_Taxonomy_Code_8                                    string
	Provider_License_Number_8                                              string
	Provider_License_Number_State_Code_8                                   string
	Healthcare_Provider_Primary_Taxonomy_Switch_8                          string
	Healthcare_Provider_Taxonomy_Code_9                                    string
	Provider_License_Number_9                                              string
	Provider_License_Number_State_Code_9                                   string
	Healthcare_Provider_Primary_Taxonomy_Switch_9                          string
	Healthcare_Provider_Taxonomy_Code_10                                   string
	Provider_License_Number_10                                             string
	Provider_License_Number_State_Code_10                                  string
	Healthcare_Provider_Primary_Taxonomy_Switch_10                         string
	Healthcare_Provider_Taxonomy_Code_11                                   string
	Provider_License_Number_11                                             string
	Provider_License_Number_State_Code_11                                  string
	Healthcare_Provider_Primary_Taxonomy_Switch_11                         string
	Healthcare_Provider_Taxonomy_Code_12                                   string
	Provider_License_Number_12                                             string
	Provider_License_Number_State_Code_12                                  string
	Healthcare_Provider_Primary_Taxonomy_Switch_12                         string
	Healthcare_Provider_Taxonomy_Code_13                                   string
	Provider_License_Number_13                                             string
	Provider_License_Number_State_Code_13                                  string
	Healthcare_Provider_Primary_Taxonomy_Switch_13                         string
	Healthcare_Provider_Taxonomy_Code_14                                   string
	Provider_License_Number_14                                             string
	Provider_License_Number_State_Code_14                                  string
	Healthcare_Provider_Primary_Taxonomy_Switch_14                         string
	Healthcare_Provider_Taxonomy_Code_15                                   string
	Provider_License_Number_15                                             string
	Provider_License_Number_State_Code_15                                  string
	Healthcare_Provider_Primary_Taxonomy_Switch_15                         string
	Other_Provider_Identifier_1                                            string
	Other_Provider_Identifier_Type_Code_1                                  string
	Other_Provider_Identifier_State_1                                      string
	Other_Provider_Identifier_Issuer_1                                     string
	Other_Provider_Identifier_2                                            string
	Other_Provider_Identifier_Type_Code_2                                  string
	Other_Provider_Identifier_State_2                                      string
	Other_Provider_Identifier_Issuer_2                                     string
	Other_Provider_Identifier_3                                            string
	Other_Provider_Identifier_Type_Code_3                                  string
	Other_Provider_Identifier_State_3                                      string
	Other_Provider_Identifier_Issuer_3                                     string
	Other_Provider_Identifier_4                                            string
	Other_Provider_Identifier_Type_Code_4                                  string
	Other_Provider_Identifier_State_4                                      string
	Other_Provider_Identifier_Issuer_4                                     string
	Other_Provider_Identifier_5                                            string
	Other_Provider_Identifier_Type_Code_5                                  string
	Other_Provider_Identifier_State_5                                      string
	Other_Provider_Identifier_Issuer_5                                     string
	Other_Provider_Identifier_6                                            string
	Other_Provider_Identifier_Type_Code_6                                  string
	Other_Provider_Identifier_State_6                                      string
	Other_Provider_Identifier_Issuer_6                                     string
	Other_Provider_Identifier_7                                            string
	Other_Provider_Identifier_Type_Code_7                                  string
	Other_Provider_Identifier_State_7                                      string
	Other_Provider_Identifier_Issuer_7                                     string
	Other_Provider_Identifier_8                                            string
	Other_Provider_Identifier_Type_Code_8                                  string
	Other_Provider_Identifier_State_8                                      string
	Other_Provider_Identifier_Issuer_8                                     string
	Other_Provider_Identifier_9                                            string
	Other_Provider_Identifier_Type_Code_9                                  string
	Other_Provider_Identifier_State_9                                      string
	Other_Provider_Identifier_Issuer_9                                     string
	Other_Provider_Identifier_10                                           string
	Other_Provider_Identifier_Type_Code_10                                 string
	Other_Provider_Identifier_State_10                                     string
	Other_Provider_Identifier_Issuer_10                                    string
	Other_Provider_Identifier_11                                           string
	Other_Provider_Identifier_Type_Code_11                                 string
	Other_Provider_Identifier_State_11                                     string
	Other_Provider_Identifier_Issuer_11                                    string
	Other_Provider_Identifier_12                                           string
	Other_Provider_Identifier_Type_Code_12                                 string
	Other_Provider_Identifier_State_12                                     string
	Other_Provider_Identifier_Issuer_12                                    string
	Other_Provider_Identifier_13                                           string
	Other_Provider_Identifier_Type_Code_13                                 string
	Other_Provider_Identifier_State_13                                     string
	Other_Provider_Identifier_Issuer_13                                    string
	Other_Provider_Identifier_14                                           string
	Other_Provider_Identifier_Type_Code_14                                 string
	Other_Provider_Identifier_State_14                                     string
	Other_Provider_Identifier_Issuer_14                                    string
	Other_Provider_Identifier_15                                           string
	Other_Provider_Identifier_Type_Code_15                                 string
	Other_Provider_Identifier_State_15                                     string
	Other_Provider_Identifier_Issuer_15                                    string
	Other_Provider_Identifier_16                                           string
	Other_Provider_Identifier_Type_Code_16                                 string
	Other_Provider_Identifier_State_16                                     string
	Other_Provider_Identifier_Issuer_16                                    string
	Other_Provider_Identifier_17                                           string
	Other_Provider_Identifier_Type_Code_17                                 string
	Other_Provider_Identifier_State_17                                     string
	Other_Provider_Identifier_Issuer_17                                    string
	Other_Provider_Identifier_18                                           string
	Other_Provider_Identifier_Type_Code_18                                 string
	Other_Provider_Identifier_State_18                                     string
	Other_Provider_Identifier_Issuer_18                                    string
	Other_Provider_Identifier_19                                           string
	Other_Provider_Identifier_Type_Code_19                                 string
	Other_Provider_Identifier_State_19                                     string
	Other_Provider_Identifier_Issuer_19                                    string
	Other_Provider_Identifier_20                                           string
	Other_Provider_Identifier_Type_Code_20                                 string
	Other_Provider_Identifier_State_20                                     string
	Other_Provider_Identifier_Issuer_20                                    string
	Other_Provider_Identifier_21                                           string
	Other_Provider_Identifier_Type_Code_21                                 string
	Other_Provider_Identifier_State_21                                     string
	Other_Provider_Identifier_Issuer_21                                    string
	Other_Provider_Identifier_22                                           string
	Other_Provider_Identifier_Type_Code_22                                 string
	Other_Provider_Identifier_State_22                                     string
	Other_Provider_Identifier_Issuer_22                                    string
	Other_Provider_Identifier_23                                           string
	Other_Provider_Identifier_Type_Code_23                                 string
	Other_Provider_Identifier_State_23                                     string
	Other_Provider_Identifier_Issuer_23                                    string
	Other_Provider_Identifier_24                                           string
	Other_Provider_Identifier_Type_Code_24                                 string
	Other_Provider_Identifier_State_24                                     string
	Other_Provider_Identifier_Issuer_24                                    string
	Other_Provider_Identifier_25                                           string
	Other_Provider_Identifier_Type_Code_25                                 string
	Other_Provider_Identifier_State_25                                     string
	Other_Provider_Identifier_Issuer_25                                    string
	Other_Provider_Identifier_26                                           string
	Other_Provider_Identifier_Type_Code_26                                 string
	Other_Provider_Identifier_State_26                                     string
	Other_Provider_Identifier_Issuer_26                                    string
	Other_Provider_Identifier_27                                           string
	Other_Provider_Identifier_Type_Code_27                                 string
	Other_Provider_Identifier_State_27                                     string
	Other_Provider_Identifier_Issuer_27                                    string
	Other_Provider_Identifier_28                                           string
	Other_Provider_Identifier_Type_Code_28                                 string
	Other_Provider_Identifier_State_28                                     string
	Other_Provider_Identifier_Issuer_28                                    string
	Other_Provider_Identifier_29                                           string
	Other_Provider_Identifier_Type_Code_29                                 string
	Other_Provider_Identifier_State_29                                     string
	Other_Provider_Identifier_Issuer_29                                    string
	Other_Provider_Identifier_30                                           string
	Other_Provider_Identifier_Type_Code_30                                 string
	Other_Provider_Identifier_State_30                                     string
	Other_Provider_Identifier_Issuer_30                                    string
	Other_Provider_Identifier_31                                           string
	Other_Provider_Identifier_Type_Code_31                                 string
	Other_Provider_Identifier_State_31                                     string
	Other_Provider_Identifier_Issuer_31                                    string
	Other_Provider_Identifier_32                                           string
	Other_Provider_Identifier_Type_Code_32                                 string
	Other_Provider_Identifier_State_32                                     string
	Other_Provider_Identifier_Issuer_32                                    string
	Other_Provider_Identifier_33                                           string
	Other_Provider_Identifier_Type_Code_33                                 string
	Other_Provider_Identifier_State_33                                     string
	Other_Provider_Identifier_Issuer_33                                    string
	Other_Provider_Identifier_34                                           string
	Other_Provider_Identifier_Type_Code_34                                 string
	Other_Provider_Identifier_State_34                                     string
	Other_Provider_Identifier_Issuer_34                                    string
	Other_Provider_Identifier_35                                           string
	Other_Provider_Identifier_Type_Code_35                                 string
	Other_Provider_Identifier_State_35                                     string
	Other_Provider_Identifier_Issuer_35                                    string
	Other_Provider_Identifier_36                                           string
	Other_Provider_Identifier_Type_Code_36                                 string
	Other_Provider_Identifier_State_36                                     string
	Other_Provider_Identifier_Issuer_36                                    string
	Other_Provider_Identifier_37                                           string
	Other_Provider_Identifier_Type_Code_37                                 string
	Other_Provider_Identifier_State_37                                     string
	Other_Provider_Identifier_Issuer_37                                    string
	Other_Provider_Identifier_38                                           string
	Other_Provider_Identifier_Type_Code_38                                 string
	Other_Provider_Identifier_State_38                                     string
	Other_Provider_Identifier_Issuer_38                                    string
	Other_Provider_Identifier_39                                           string
	Other_Provider_Identifier_Type_Code_39                                 string
	Other_Provider_Identifier_State_39                                     string
	Other_Provider_Identifier_Issuer_39                                    string
	Other_Provider_Identifier_40                                           string
	Other_Provider_Identifier_Type_Code_40                                 string
	Other_Provider_Identifier_State_40                                     string
	Other_Provider_Identifier_Issuer_40                                    string
	Other_Provider_Identifier_41                                           string
	Other_Provider_Identifier_Type_Code_41                                 string
	Other_Provider_Identifier_State_41                                     string
	Other_Provider_Identifier_Issuer_41                                    string
	Other_Provider_Identifier_42                                           string
	Other_Provider_Identifier_Type_Code_42                                 string
	Other_Provider_Identifier_State_42                                     string
	Other_Provider_Identifier_Issuer_42                                    string
	Other_Provider_Identifier_43                                           string
	Other_Provider_Identifier_Type_Code_43                                 string
	Other_Provider_Identifier_State_43                                     string
	Other_Provider_Identifier_Issuer_43                                    string
	Other_Provider_Identifier_44                                           string
	Other_Provider_Identifier_Type_Code_44                                 string
	Other_Provider_Identifier_State_44                                     string
	Other_Provider_Identifier_Issuer_44                                    string
	Other_Provider_Identifier_45                                           string
	Other_Provider_Identifier_Type_Code_45                                 string
	Other_Provider_Identifier_State_45                                     string
	Other_Provider_Identifier_Issuer_45                                    string
	Other_Provider_Identifier_46                                           string
	Other_Provider_Identifier_Type_Code_46                                 string
	Other_Provider_Identifier_State_46                                     string
	Other_Provider_Identifier_Issuer_46                                    string
	Other_Provider_Identifier_47                                           string
	Other_Provider_Identifier_Type_Code_47                                 string
	Other_Provider_Identifier_State_47                                     string
	Other_Provider_Identifier_Issuer_47                                    string
	Other_Provider_Identifier_48                                           string
	Other_Provider_Identifier_Type_Code_48                                 string
	Other_Provider_Identifier_State_48                                     string
	Other_Provider_Identifier_Issuer_48                                    string
	Other_Provider_Identifier_49                                           string
	Other_Provider_Identifier_Type_Code_49                                 string
	Other_Provider_Identifier_State_49                                     string
	Other_Provider_Identifier_Issuer_49                                    string
	Other_Provider_Identifier_50                                           string
	Other_Provider_Identifier_Type_Code_50                                 string
	Other_Provider_Identifier_State_50                                     string
	Other_Provider_Identifier_Issuer_50                                    string
	Is_Sole_Proprietor                                                     string
	Is_Organization_Subpart                                                string
	Parent_Organization_LBN                                                string
	Parent_Organization_TIN                                                string
	Authorized_Official_Name_Prefix_Text                                   string
	Authorized_Official_Name_Suffix_Text                                   string
	Authorized_Official_Credential_Text                                    string
	Healthcare_Provider_Taxonomy_Group_1                                   string
	Healthcare_Provider_Taxonomy_Group_2                                   string
	Healthcare_Provider_Taxonomy_Group_3                                   string
	Healthcare_Provider_Taxonomy_Group_4                                   string
	Healthcare_Provider_Taxonomy_Group_5                                   string
	Healthcare_Provider_Taxonomy_Group_6                                   string
	Healthcare_Provider_Taxonomy_Group_7                                   string
	Healthcare_Provider_Taxonomy_Group_8                                   string
	Healthcare_Provider_Taxonomy_Group_9                                   string
	Healthcare_Provider_Taxonomy_Group_10                                  string
	Healthcare_Provider_Taxonomy_Group_11                                  string
	Healthcare_Provider_Taxonomy_Group_12                                  string
	Healthcare_Provider_Taxonomy_Group_13                                  string
	Healthcare_Provider_Taxonomy_Group_14                                  string
	Healthcare_Provider_Taxonomy_Group_15                                  string
}

func ParseNPIdataLine(line []string) NPICsvLine {
	data := NPICsvLine{
		NPI:              line[0],
		Entity_Type_Code: line[1],
		Provider_Organization_Name_Legal_Business_Name:          line[4],
		Provider_Other_Organization_Name:                        line[11],
		Provider_First_Line_Business_Practice_Location_Address:  line[28],
		Provider_Second_Line_Business_Practice_Location_Address: line[29],
		Provider_Business_Practice_Location_Address_City_Name:   line[30],
		Provider_Business_Practice_Location_Address_State_Name:  line[31],
		Provider_Business_Practice_Location_Address_Postal_Code: line[32],
		Healthcare_Provider_Taxonomy_Code_1:                     line[47],
	}
	return data
}

func BuildNPIOrgFromNPICsvLine(data NPICsvLine) *endpointmanager.NPIOrganization {
	npi_org := &endpointmanager.NPIOrganization{
		NPI_ID:        data.NPI,
		Name:          data.Provider_Organization_Name_Legal_Business_Name,
		SecondaryName: data.Provider_Other_Organization_Name,
		Location: &endpointmanager.Location{
			Address1: data.Provider_First_Line_Business_Practice_Location_Address,
			Address2: data.Provider_Second_Line_Business_Practice_Location_Address,
			City:     data.Provider_Business_Practice_Location_Address_City_Name,
			State:    data.Provider_Business_Practice_Location_Address_State_Name,
			ZipCode:  data.Provider_Business_Practice_Location_Address_Postal_Code},
		Taxonomy: data.Healthcare_Provider_Taxonomy_Code_1}
	return npi_org
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

// ReadCsv accepts a file and returns its content as a multi-dimentional type
// with lines and each column. Only parses to string type.
func ReadCsv(filename string) ([][]string, error) {

	// Open CSV file
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}
	// return lines without header line
	return lines[1:], nil
}

// Parses NPI Org data out of fname, writes it to store and returns the number of organizations processed
func ParseAndStoreNPIFile(fname string, store *postgresql.Store) int {
	// Provider organization .csv downloaded from http://download.cms.gov/nppes/NPI_Files.html
	lines, err := ReadCsv(fname)
	panicOnErr(err)
	added_or_updated := 0
	// Loop through lines & turn into object
	for _, line := range lines {
		data := ParseNPIdataLine(line)
		// We will only parse out organizations (entiy_type_code == 2), not individual providers
		if data.Entity_Type_Code == "2" {
			npi_org := BuildNPIOrgFromNPICsvLine(data)
			err = store.AddOrUpdateNPIOrganization(npi_org)
			if err != nil {
				if err != nil {
					log.Printf("%s", err)
				}
			} else {
				added_or_updated += 1
			}
		}
	}
	return added_or_updated
}
