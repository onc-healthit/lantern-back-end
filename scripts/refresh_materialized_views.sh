#!/bin/sh

# Refresh mv_contact_information
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_contact_information;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_contact_information." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_contact_information_uniq;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_contact_information_uniq." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX mv_contact_information_uniq ON mv_contact_information (url, requested_fhir_version, contact_rank);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_contact_information_uniq." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_contact_information_fhir_version_idx ON mv_contact_information (fhir_version);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_contact_information_fhir_version_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_contact_information_vendor_name_idx ON mv_contact_information (vendor_name);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_contact_information_vendor_name_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_contact_information_has_contact_idx ON mv_contact_information (has_contact);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_contact_information_has_contact_idx." >> $log_file
}