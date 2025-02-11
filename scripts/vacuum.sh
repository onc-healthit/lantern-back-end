#!/bin/sh

total_start=$(date +%s)

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE fhir_endpoints" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "fhir_endpoints -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE certification_criteria" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "certification_criteria -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE endpoint_organization" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "endpoint_organization -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE fhir_endpoint_organizations" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "fhir_endpoint_organizations -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE fhir_endpoint_organizations_map" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "fhir_endpoint_organizations_map -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE fhir_endpoints_availability" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "fhir_endpoints_availability -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE fhir_endpoints_info" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "fhir_endpoints_info -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE fhir_endpoints_info_history" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "fhir_endpoints_info_history -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE fhir_endpoints_metadata" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "fhir_endpoints_metadata -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE healthit_products" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "healthit_products -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE healthit_products_map" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "healthit_products_map -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE info_history_pruning_metadata" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "info_history_pruning_metadata -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE list_source_info" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "list_source_info -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE npi_contacts" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "npi_contacts -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE npi_organizations" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "npi_organizations -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE product_criteria" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "product_criteria -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE schema_migrations" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "schema_migrations -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE validation_results" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "validation_results -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE validations" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "validations -- Elapsed time: ${elapsed} seconds"

start=$(date +%s)

docker exec -t lantern-back-end_postgres_1 psql -t -c "VACUUM FULL VERBOSE vendors" -U lantern -d lantern

end=$(date +%s)
elapsed=$(( end - start ))
echo "vendors -- Elapsed time: ${elapsed} seconds"

total_end=$(date +%s)
elapsed=$(( total_end - total_start ))
echo "FINISHED -- Total Elapsed time: ${elapsed} seconds"