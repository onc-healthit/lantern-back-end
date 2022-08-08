package main

import (
	"context"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	log.Info("Starting to export endpoints")

	err := config.SetupConfig()
	helpers.FailOnError("Error setting up config", err)
	ctx := context.Background()

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("Error creating store", err)

	// get report data and export to tmp/export.csv
	getCHPLEndpointsStatement := `
	COPY (
        SELECT DISTINCT
   			f.list_source AS service_base_url,
   			f.url as endpoint,
			vendors.name as developer,
   			CASE WHEN COUNT(metadata.http_response)=0
        		THEN NULL
        		ELSE ROUND(100 - COUNT(NULLIF(metadata.http_response, 200))::numeric/COUNT(metadata.http_response) * 100,2)
   			END AS server_status_percent,
   			string_agg(DISTINCT metadata.http_response::varchar, ',') AS server_status_info,
   			CASE WHEN hist.capability_statement = 'null'
        		THEN 'missing'
        		ELSE 'valid'
   			END AS cap_stat_status,
   			ROUND(COUNT(NULLIF(hist.capability_statement, 'null'))::numeric / COUNT(hist.capability_statement) * 100,2) AS cap_stat_status_days,
			CASE WHEN metadata.http_response != 200
				THEN 'Unreachable'
			    ELSE 'Doesn''t service Capability Statement request'
			END AS reason
		FROM fhir_endpoints f
		LEFT JOIN list_source_info ON f.list_source = list_source_info.list_source
		LEFT JOIN (
    		SELECT
    			url, requested_fhir_version, metadata_id, capability_statement::jsonb, updated_at, vendor_id
    		FROM
    		fhir_endpoints_info_history 
		) AS hist ON f.url = hist.url
		LEFT JOIN vendors ON hist.vendor_id = vendors.id
		LEFT JOIN fhir_endpoints_metadata AS metadata ON hist.metadata_id = metadata.id
		WHERE list_source_info.is_chpl = true AND age(hist.updated_at) < '30 days'
		GROUP BY
			f.list_source,
			f.url,
			hist.capability_statement,
			metadata.http_response,
			vendors.name
		HAVING hist.capability_statement = 'null' OR metadata.http_response != 200
	)
	TO '/tmp/export.csv'
	DELIMITER ',' CSV HEADER
	;
	`

	_, err = store.DB.ExecContext(ctx, getCHPLEndpointsStatement)
	helpers.FailOnError("Error exporting csv. Error: ", err)

}
