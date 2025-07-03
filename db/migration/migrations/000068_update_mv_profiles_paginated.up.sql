BEGIN;

-- Create index to optimize filtering in profiles module
CREATE INDEX idx_profiles_filters ON mv_profiles_paginated (fhir_version, vendor_name, resource, profileurl, url);

COMMIT;

