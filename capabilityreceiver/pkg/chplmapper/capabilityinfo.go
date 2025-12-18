package chplmapper

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/pkg/errors"
)

type details struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ChplEndpointListProductInfo struct {
	ListSourceURL    string                 `json:"listSourceURL"`
	SoftwareProducts []ChplCertifiedProduct `json:"softwareProducts"`
}

type ChplCertifiedProduct struct {
	ChplProductNumber string  `json:"chplProductNumber"`
	Developer         details `json:"developer"`
}

type ChplMapResults struct {
	ChplProductIDs []string
	ChplDeveloper  []string
}

// MatchEndpointToProduct creates the database association between the endpoint and the HealthITProduct,
func MatchEndpointToProduct(ctx context.Context, ep *endpointmanager.FHIREndpointInfo, store *postgresql.Store, matchFile string, productIds []string) error {

	softwareName := ""
	softwareVersion := ""
	chplIDArr := []string{}

	if ep.CapabilityStatement != nil {
		chplProductNameVersion, err := openProductLinksFile(matchFile)
		if err != nil {
			return errors.Wrap(err, "error matching the capability statement to a CHPL product")
		}

		softwareName, err = ep.CapabilityStatement.GetSoftwareName()
		if err != nil {
			return errors.Wrap(err, "error matching the capability statement to a CHPL product")
		}
		softwareVersion, err = ep.CapabilityStatement.GetSoftwareVersion()
		if err != nil {
			return errors.Wrap(err, "error matching the capability statement to a CHPL product")
		}

		chplIDMatchFile := chplProductNameVersion[softwareName][softwareVersion]

		if len(chplIDMatchFile) != 0 {
			chplIDArr = append(chplIDArr, chplIDMatchFile)
		}
	}

	if len(productIds) > 0 {
		chplIDArr = append(chplIDArr, productIds...)
	}

	var healthITProductsArr []*endpointmanager.HealthITProduct
	var err error
	if len(softwareName) != 0 {
		healthITProductsArr, err = store.GetActiveHealthITProductsUsingName(ctx, softwareName)
		if err != nil {
			return err
		}
	}

	for _, healthITProduct := range healthITProductsArr {
		if len(softwareVersion) == 0 {
			if !helpers.StringArrayContains(chplIDArr, healthITProduct.CHPLID) {
				chplIDArr = append(chplIDArr, healthITProduct.CHPLID)
			}
		} else {
			if strings.EqualFold(healthITProduct.Version, softwareVersion) {
				if !helpers.StringArrayContains(chplIDArr, healthITProduct.CHPLID) {
					chplIDArr = append(chplIDArr, healthITProduct.CHPLID)
				}
			}
		}
	}

	log.Info("chplIDArr: ", chplIDArr, "\n")

	for _, chplID := range chplIDArr {
		healthITProductID, err := store.GetHealthITProductIDByCHPLID(ctx, chplID)
		// No errors thrown means a healthit product with CHPLID was found and can be set on ep
		if err == nil {
			healthITMapID, err := store.AddHealthITProductMap(ctx, ep.HealthITProductID, healthITProductID)
			if err != nil {
				return err
			}
			ep.HealthITProductID = healthITMapID
		}
	}

	return nil
}

func openProductLinksFile(filepath string) (map[string]map[string]string, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	var softwareNameVersion []map[string]string
	byteValueFile, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var chplMap = make(map[string]map[string]string)
	if len(byteValueFile) != 0 {
		err = json.Unmarshal(byteValueFile, &softwareNameVersion)
		if err != nil {
			return nil, err
		}
		for _, obj := range softwareNameVersion {
			var name = obj["name"]
			var version = obj["version"]
			var chplID = obj["CHPLID"]
			if name != "" && version != "" && chplID != "" {
				if chplMap[name] == nil {
					chplMap[name] = make(map[string]string)
				}
				chplMap[name][version] = chplID
			}
		}
	}

	return chplMap, nil
}

func OpenCHPLEndpointListInfoFile(filepath string) (map[string]ChplMapResults, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	var softwareListMap = make(map[string]ChplMapResults)

	byteValueFile, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var chplMap []ChplEndpointListProductInfo
	if len(byteValueFile) != 0 {
		err = json.Unmarshal(byteValueFile, &chplMap)
		if err != nil {
			return nil, err
		}
		for _, obj := range chplMap {
			var listSource = obj.ListSourceURL
			var softwareProducts = obj.SoftwareProducts

			chplMapResult := ChplMapResults{ChplProductIDs: []string{}, ChplDeveloper: []string{}}

			chplID := ""

			for _, prod := range softwareProducts {
				chplID = prod.ChplProductNumber

				if chplID != "" {
					chplMapResult.ChplProductIDs = append(chplMapResult.ChplProductIDs, chplID)
				}
			}

			if listSource != "" {
				for _, softwareProduct := range softwareProducts {
					chplMapResult.ChplDeveloper = append(chplMapResult.ChplDeveloper, softwareProduct.Developer.Name)
				}

				softwareListMap[listSource] = chplMapResult
			}

		}
	}

	return softwareListMap, nil
}
