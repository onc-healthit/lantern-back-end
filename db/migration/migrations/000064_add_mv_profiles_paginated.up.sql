BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_profiles_paginated CASCADE;

CREATE MATERIALIZED VIEW mv_profiles_paginated AS
SELECT 
  row_number() OVER (ORDER BY vendor_name, url, profileurl) AS page_id,
  url,
  profileurl,
  profilename,
  resource,
  fhir_version,
  vendor_name
FROM (
  SELECT DISTINCT 
    url,
    profileurl,
    profilename,
    resource,
    fhir_version,
    vendor_name
  FROM endpoint_supported_profiles_mv
) distinct_profiles
ORDER BY vendor_name, url, profileurl;

-- Create indexes for fast filtering and pagination
CREATE UNIQUE INDEX mv_profiles_paginated_page_id_idx ON mv_profiles_paginated(page_id);
CREATE INDEX mv_profiles_paginated_fhir_version_idx ON mv_profiles_paginated(fhir_version);
CREATE INDEX mv_profiles_paginated_vendor_name_idx ON mv_profiles_paginated(vendor_name);
CREATE INDEX mv_profiles_paginated_resource_idx ON mv_profiles_paginated(resource);
CREATE INDEX mv_profiles_paginated_profileurl_idx ON mv_profiles_paginated(profileurl);

-- Composite index for common filter combinations
CREATE INDEX mv_profiles_paginated_composite_idx ON mv_profiles_paginated(vendor_name, fhir_version, resource);

COMMIT;