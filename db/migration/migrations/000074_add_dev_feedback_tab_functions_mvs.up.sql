BEGIN;

-- Function to validate NPI using Luhn algorithm 
CREATE OR REPLACE FUNCTION validate_npi_luhn(npi TEXT) 
RETURNS BOOLEAN AS $$
DECLARE
    digits INTEGER[];
    checksum INTEGER := 0;
    doubled INTEGER;
    i INTEGER;
BEGIN
    -- Check if NPI is exactly 10 digits
    IF length(npi) != 10 OR npi !~ '^[0-9]{10}$' THEN
        RETURN FALSE;
    END IF;
    
    -- Convert string to array of integers
    FOR i IN 1..10 LOOP
        digits[i] := substring(npi FROM i FOR 1)::INTEGER;
    END LOOP;
    
    -- Double digits in positions 1,3,5,7,9 (R uses 1-indexed positions)
    FOR i IN 1..5 LOOP
        doubled := digits[i*2-1] * 2;
        IF doubled > 9 THEN
            doubled := doubled - 9;
        END IF;
        checksum := checksum + doubled;
    END LOOP;
    
    -- Add digits in positions 2,4,6,8,10
    FOR i IN 1..5 LOOP
        checksum := checksum + digits[i*2];
    END LOOP;
    
    -- Add 24 and check modulo 10
    checksum := checksum + 24;
    RETURN (checksum % 10) = 0;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to check if name is address-like 
CREATE OR REPLACE FUNCTION is_address_like(name_text TEXT)
RETURNS BOOLEAN AS $$
DECLARE
    clean_name TEXT;
    score INTEGER := 0;
BEGIN
    IF name_text IS NULL OR name_text = '' THEN
        RETURN FALSE;
    END IF;
    
    -- Clean the name (remove HTML tags)
    clean_name := regexp_replace(name_text, '<[^>]+>', '', 'g');
    clean_name := trim(clean_name);
    
    -- Strong address indicators
    IF clean_name ~ '^[0-9]+' THEN 
        score := score + 3; -- Starts with number
    END IF;
    
    IF clean_name ~* '\b(St|Street|Ave|Avenue|Blvd|Boulevard|Rd|Road|Dr|Drive|Ln|Lane|Ct|Court|Cir|Circle|Way|Pl|Place|Pkwy|Parkway|Ter|Terrace)\b' THEN 
        score := score + 3; -- Street suffix
    END IF;
    
    IF clean_name ~* '\b(Suite|Ste|Apt|Apartment|Unit|Floor|Fl|Room|Rm|Building|Bldg|#)\b' THEN 
        score := score + 2; -- Unit/suite
    END IF;
    
    IF clean_name ~ '\b[0-9]{5}(-[0-9]{4})?\b' THEN 
        score := score + 3; -- ZIP code
    END IF;
    
    IF clean_name ~* '\b(AL|AK|AZ|AR|CA|CO|CT|DE|FL|GA|HI|ID|IL|IN|IA|KS|KY|LA|ME|MD|MA|MI|MN|MS|MO|MT|NE|NV|NH|NJ|NM|NY|NC|ND|OH|OK|OR|PA|RI|SC|SD|TN|TX|UT|VT|VA|WA|WV|WI|WY)\b' THEN 
        score := score + 2; -- State abbreviation
    END IF;
    
    -- Directional indicators
    IF clean_name ~* '\b(North|South|East|West|N|S|E|W)\b' THEN 
        score := score + 1;
    END IF;
    
    -- Multiple commas (address format)
    IF length(clean_name) - length(replace(clean_name, ',', '')) >= 2 THEN 
        score := score + 2;
    END IF;
    
    -- Healthcare/organization keywords reduce address likelihood
    IF clean_name ~* '\b(Hospital|Clinic|Center|Centre|Health|Medical|System|Services|LLC|Corp|Corporation|Inc|Incorporated|Ltd|Limited|Associates|Group|Foundation|Institute|University|College|Pharmacy|Laboratory|Labs?)\b' THEN
        score := score - 3;
    END IF;
    
    -- Additional org-like terms specific to healthcare
    IF clean_name ~* '\b(Family|Internal|Primary|Urgent|Emergency|Pediatric|Cardiology|Orthopedic|Dental|Vision|Eye|Care)\b' THEN
        score := score - 2;
    END IF;
    
    RETURN score >= 4;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to validate individual identifier 
CREATE OR REPLACE FUNCTION validate_identifier_value(identifier_type TEXT, identifier_value TEXT)
RETURNS TABLE(valid BOOLEAN, error_msg TEXT) AS $$
BEGIN
    IF identifier_value IS NULL OR trim(identifier_value) = '' THEN
        RETURN QUERY SELECT FALSE, 'Missing identifier value';
        RETURN;
    END IF;
    
    identifier_type := upper(trim(identifier_type));
    identifier_value := trim(identifier_value);
    
    IF identifier_type = 'NPI' THEN
        -- us-core-16: NPI must be 10 digits
        IF identifier_value !~ '^[0-9]{10}$' THEN
            RETURN QUERY SELECT FALSE, 'NPI must be exactly 10 digits';
            RETURN;
        END IF;
        
        -- us-core-17: NPI check digit must be valid (Luhn algorithm)
        IF NOT validate_npi_luhn(identifier_value) THEN
            RETURN QUERY SELECT FALSE, 'NPI check digit is invalid (Luhn algorithm failed)';
            RETURN;
        END IF;
        
        RETURN QUERY SELECT TRUE, NULL::TEXT;
        
    ELSIF identifier_type = 'CLIA' THEN
        -- us-core-18: CLIA number must be 10 digits with a letter "D" in third position
        IF identifier_value !~ '^[0-9]{2}D[0-9]{7}$' THEN
            RETURN QUERY SELECT FALSE, 'CLIA must be 10 characters: 2 digits + ''D'' + 7 digits';
            RETURN;
        END IF;
        
        RETURN QUERY SELECT TRUE, NULL::TEXT;
        
    ELSIF identifier_type = 'NAIC' THEN
        -- us-core-19: NAIC must be 5 digits
        IF identifier_value !~ '^[0-9]{5}$' THEN
            RETURN QUERY SELECT FALSE, 'NAIC must be exactly 5 digits';
            RETURN;
        END IF;
        
        RETURN QUERY SELECT TRUE, NULL::TEXT;
        
    ELSE
        -- Non-standard identifier type
        RETURN QUERY SELECT FALSE, 'Non-standard identifier type (should use NPI, CLIA, or NAIC)';
    END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function for comprehensive name validation 
CREATE OR REPLACE FUNCTION is_valid_organization_name(org_name TEXT)
RETURNS BOOLEAN AS $$
DECLARE
    clean_name TEXT;
    special_chars INTEGER;
    total_chars INTEGER;
BEGIN
    IF org_name IS NULL OR org_name = '' THEN
        RETURN FALSE;
    END IF;
    
    -- Remove HTML tags and clean
    clean_name := regexp_replace(org_name, '<[^>]+>', '', 'g');
    clean_name := trim(clean_name);
    
    -- Minimum length check
    IF length(clean_name) < 3 THEN
        RETURN FALSE;
    END IF;
    
    -- Placeholder patterns
    IF upper(clean_name) IN ('-', '.', 'N/A', 'NA', 'UNKNOWN', 'TEST', 'EXAMPLE', 'TBD', 'TODO') THEN
        RETURN FALSE;
    END IF;
    
    -- Reject if all digits
    IF clean_name ~ '^[0-9]+$' THEN
        RETURN FALSE;
    END IF;
    
    -- Reject digits with only separators/symbols
    IF clean_name ~ '^[0-9()/.\\-]+$' THEN
        RETURN FALSE;
    END IF;
    
    -- Reject only non-word characters
    IF clean_name ~ '^\W+$' THEN
        RETURN FALSE;
    END IF;
    
    -- Reject phone number patterns
    IF clean_name ~ '^\(?[0-9]{3}\)?[- ]?[0-9]{3}[- ]?[0-9]{4}$' THEN
        RETURN FALSE;
    END IF;
    
    -- Reject if it looks like an address
    IF is_address_like(clean_name) THEN
        RETURN FALSE;
    END IF;
    
    -- Check special character ratio (equivalent to original R logic)
    total_chars := length(clean_name);
    special_chars := length(regexp_replace(clean_name, '[a-zA-Z0-9 ]', '', 'g'));
    
    IF special_chars::DECIMAL / total_chars::DECIMAL > 0.3 THEN
        RETURN FALSE;
    END IF;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function for comprehensive address validation 
CREATE OR REPLACE FUNCTION is_valid_organization_address(address_text TEXT)
RETURNS BOOLEAN AS $$
DECLARE
    clean_address TEXT;
    comma_count INTEGER;
BEGIN
    IF address_text IS NULL OR address_text = '' THEN
        RETURN FALSE;
    END IF;
    
    -- Remove HTML tags and clean
    clean_address := regexp_replace(address_text, '<[^>]+>', '', 'g');
    clean_address := trim(clean_address);
    
    -- Minimum length check
    IF length(clean_address) < 10 THEN
        RETURN FALSE;
    END IF;
    
    -- Check for placeholder addresses (equivalent to original R logic)
    IF upper(clean_address) ~ '123 (MAIN|TEST) ST' OR upper(clean_address) ~ '123 (MAIN|TEST) STREET' THEN
        RETURN FALSE;
    END IF;
    
    -- Must have street number
    IF NOT (clean_address ~ '[0-9]+') THEN
        RETURN FALSE;
    END IF;
    
    -- Must have city, state structure (at least 2 commas)
    comma_count := length(clean_address) - length(replace(clean_address, ',', ''));
    IF comma_count < 2 THEN
        RETURN FALSE;
    END IF;
    
    -- Must have ZIP code
    IF NOT (clean_address ~ '[0-9]{5}') THEN
        RETURN FALSE;
    END IF;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- 1. Complete organization quality materialized view with ALL original validations
CREATE MATERIALIZED VIEW mv_organization_quality AS
WITH base_org_data AS (
    SELECT
        org_id,
        organization_name,
        identifier_types_html,
        identifier_values_html,
        addresses_html,
        vendor_names_array,
        urls_array,
        endpoint_urls_html
    FROM mv_organizations_final
),
identifier_parsing AS (
    SELECT 
        org_id,
        organization_name,
        identifier_types_html,
        identifier_values_html,
        addresses_html,
        vendor_names_array,
        urls_array,
        endpoint_urls_html,
        
        -- Check if identifier data exists (equivalent to original R logic)
        CASE 
            WHEN (identifier_types_html IS NULL OR identifier_types_html = '') AND 
                 (identifier_values_html IS NULL OR identifier_values_html = '') THEN 'no_identifiers'
            WHEN (identifier_types_html IS NULL OR identifier_types_html = '') OR 
                 (identifier_values_html IS NULL OR identifier_values_html = '') THEN 'incomplete_data'
            ELSE 'has_data'
        END as identifier_data_status,
        
        -- Parse identifier types and values from HTML format
        CASE 
            WHEN identifier_types_html IS NULL OR identifier_types_html = '' THEN ARRAY[]::TEXT[]
            ELSE string_to_array(
                regexp_replace(
                    regexp_replace(identifier_types_html, '<br/?>', '|', 'gi'), 
                    '\s+', ' ', 'g'
                ), '|'
            )
        END as parsed_types,
        
        CASE 
            WHEN identifier_values_html IS NULL OR identifier_values_html = '' THEN ARRAY[]::TEXT[]
            ELSE string_to_array(
                regexp_replace(
                    regexp_replace(identifier_values_html, '<br/?>', '|', 'gi'), 
                    '\s+', ' ', 'g'
                ), '|'
            )
        END as parsed_values
        
    FROM base_org_data
),
identifier_validation_detailed AS (
    SELECT *,
        -- Count total identifiers
        CASE 
            WHEN identifier_data_status = 'no_identifiers' THEN 0
            WHEN parsed_types IS NULL THEN 0
            ELSE COALESCE(array_length(parsed_types, 1), 0)
        END as total_identifier_count,
        
        -- Comprehensive identifier validation using the new functions
        CASE 
            WHEN identifier_data_status = 'no_identifiers' THEN 0
            WHEN identifier_data_status = 'incomplete_data' THEN 0
            WHEN parsed_types IS NULL OR parsed_values IS NULL THEN 0
            WHEN COALESCE(array_length(parsed_types, 1), 0) != COALESCE(array_length(parsed_values, 1), 0) THEN 0
            ELSE (
                SELECT COUNT(*)::INT
                FROM unnest(parsed_types, parsed_values) AS t(itype, ivalue)
                WHERE (SELECT valid FROM validate_identifier_value(itype, ivalue)) = TRUE
            )
        END as conformant_identifier_count,
        
        -- Detailed identifier counts by type with full validation
        COALESCE((SELECT COUNT(*) FROM unnest(parsed_types) t WHERE upper(trim(t)) = 'NPI'), 0)::INT as npi_count,
        COALESCE((SELECT COUNT(*) FROM unnest(parsed_types) t WHERE upper(trim(t)) = 'CLIA'), 0)::INT as clia_count,
        COALESCE((SELECT COUNT(*) FROM unnest(parsed_types) t WHERE upper(trim(t)) = 'NAIC'), 0)::INT as naic_count,
        COALESCE((SELECT COUNT(*) FROM unnest(parsed_types) t WHERE upper(trim(t)) NOT IN ('NPI', 'CLIA', 'NAIC') AND trim(t) != ''), 0)::INT as other_count,
        
        -- Valid counts by type using full validation functions
        COALESCE((
            SELECT COUNT(*)::INT
            FROM unnest(parsed_types, parsed_values) AS t(itype, ivalue)
            WHERE upper(trim(itype)) = 'NPI' 
              AND (SELECT valid FROM validate_identifier_value(itype, ivalue)) = TRUE
        ), 0) as npi_valid,
        
        COALESCE((
            SELECT COUNT(*)::INT
            FROM unnest(parsed_types, parsed_values) AS t(itype, ivalue)
            WHERE upper(trim(itype)) = 'CLIA' 
              AND (SELECT valid FROM validate_identifier_value(itype, ivalue)) = TRUE
        ), 0) as clia_valid,
        
        COALESCE((
            SELECT COUNT(*)::INT
            FROM unnest(parsed_types, parsed_values) AS t(itype, ivalue)
            WHERE upper(trim(itype)) = 'NAIC' 
              AND (SELECT valid FROM validate_identifier_value(itype, ivalue)) = TRUE
        ), 0) as naic_valid
        
    FROM identifier_parsing
),
quality_calculations AS (
    SELECT *,
        -- Identifier validation results
        conformant_identifier_count > 0 as has_valid_identifiers,
        CASE 
            WHEN total_identifier_count = 0 THEN 0.0
            ELSE (conformant_identifier_count::DECIMAL / total_identifier_count::DECIMAL * 100)
        END as identifier_conformance_rate,
        
        -- Name validation using the comprehensive function
        is_valid_organization_name(organization_name) as has_valid_name,
        
        -- Address validation using the comprehensive function
        is_valid_organization_address(addresses_html) as has_valid_address,
        
        -- Calculate invalid counts
        GREATEST(0, npi_count - npi_valid) as npi_invalid,
        GREATEST(0, clia_count - clia_valid) as clia_invalid,
        GREATEST(0, naic_count - naic_valid) as naic_invalid
        
    FROM identifier_validation_detailed
)
SELECT 
    org_id,
    organization_name,
    identifier_types_html,
    identifier_values_html,
    addresses_html,
    vendor_names_array,
    urls_array,
    endpoint_urls_html,
    identifier_data_status,
    total_identifier_count,
    conformant_identifier_count,
    ROUND(identifier_conformance_rate::NUMERIC, 1) as identifier_conformance_rate,
    has_valid_identifiers,
    has_valid_name,
    has_valid_address,
    
    -- Identifier counts by type
    npi_count,
    clia_count,
    naic_count,
    other_count,
    npi_valid,
    clia_valid,
    naic_valid,
    npi_invalid,
    clia_invalid,
    naic_invalid,
    other_count as other_invalid, -- All "other" types are invalid
    
    -- Overall quality score (0-3)
    (CASE WHEN has_valid_identifiers THEN 1 ELSE 0 END +
     CASE WHEN has_valid_name THEN 1 ELSE 0 END +
     CASE WHEN has_valid_address THEN 1 ELSE 0 END) as overall_quality_score,
     
    -- Conformance categories
    CASE 
        WHEN identifier_conformance_rate = 100 THEN 'Fully Conformant'
        WHEN identifier_conformance_rate >= 50 THEN 'Partially Conformant'
        WHEN conformant_identifier_count > 0 THEN 'Minimally Conformant'
        ELSE 'Non-Conformant'
    END as identifier_conformance_category,
    
    -- Status categories
    CASE 
        WHEN identifier_data_status = 'no_identifiers' THEN 'no_identifiers'
        WHEN conformant_identifier_count = 0 THEN 'invalid_only'
        WHEN conformant_identifier_count = total_identifier_count THEN 'all_valid'
        ELSE 'mixed_valid_invalid'
    END as identifier_status
    
FROM quality_calculations;

-- Create indexes
CREATE UNIQUE INDEX idx_mv_organization_quality_complete_org_id ON mv_organization_quality(org_id);
CREATE INDEX idx_mv_organization_quality_complete_vendor ON mv_organization_quality USING GIN(vendor_names_array);
CREATE INDEX idx_mv_organization_quality_complete_valid_id ON mv_organization_quality(has_valid_identifiers);
CREATE INDEX idx_mv_organization_quality_complete_valid_name ON mv_organization_quality(has_valid_name);
CREATE INDEX idx_mv_organization_quality_complete_valid_address ON mv_organization_quality(has_valid_address);
CREATE INDEX idx_mv_organization_quality_complete_conformance ON mv_organization_quality(identifier_conformance_category);
CREATE INDEX idx_mv_organization_quality_complete_status ON mv_organization_quality(identifier_status);
CREATE INDEX idx_mv_organization_quality_complete_score ON mv_organization_quality(overall_quality_score);

-- 2. Update summary views to use the complete validation data
CREATE MATERIALIZED VIEW mv_organization_quality_summary AS
SELECT 
    vendor_name,
    COUNT(*) as total_organizations,
    
    -- Identifier validation summary
    COUNT(*) FILTER (WHERE has_valid_identifiers) as organizations_with_valid_identifiers,
    COUNT(*) FILTER (WHERE identifier_status = 'no_identifiers') as organizations_with_no_identifiers,
    COUNT(*) FILTER (WHERE identifier_status = 'invalid_only') as organizations_with_invalid_only,
    COUNT(*) FILTER (WHERE identifier_status = 'all_valid') as organizations_all_valid,
    COUNT(*) FILTER (WHERE identifier_status = 'mixed_valid_invalid') as organizations_mixed_valid,
    
    -- Quality metrics
    COUNT(*) FILTER (WHERE has_valid_name) as organizations_with_valid_names,
    COUNT(*) FILTER (WHERE has_valid_address) as organizations_with_valid_addresses,
    COUNT(*) FILTER (WHERE overall_quality_score >= 2) as high_quality_organizations,
    COUNT(*) FILTER (WHERE overall_quality_score <= 1) as low_quality_organizations,
    
    -- Conformance breakdown
    COUNT(*) FILTER (WHERE identifier_conformance_category = 'Fully Conformant') as fully_conformant,
    COUNT(*) FILTER (WHERE identifier_conformance_category = 'Partially Conformant') as partially_conformant,
    COUNT(*) FILTER (WHERE identifier_conformance_category = 'Minimally Conformant') as minimally_conformant,
    COUNT(*) FILTER (WHERE identifier_conformance_category = 'Non-Conformant') as non_conformant,
    
    -- Averages
    ROUND(AVG(identifier_conformance_rate), 1) as avg_conformance_rate,
    ROUND(AVG(overall_quality_score), 2) as avg_quality_score,
    
    -- Percentages
    ROUND(COUNT(*) FILTER (WHERE has_valid_identifiers)::DECIMAL / COUNT(*)::DECIMAL * 100, 1) as identifier_percentage,
    ROUND(COUNT(*) FILTER (WHERE has_valid_name)::DECIMAL / COUNT(*)::DECIMAL * 100, 1) as name_percentage,
    ROUND(COUNT(*) FILTER (WHERE has_valid_address)::DECIMAL / COUNT(*)::DECIMAL * 100, 1) as address_percentage

FROM (
    SELECT 
        UNNEST(vendor_names_array) as vendor_name,
        has_valid_identifiers,
        identifier_status,
        has_valid_name,
        has_valid_address,
        overall_quality_score,
        identifier_conformance_category,
        identifier_conformance_rate
    FROM mv_organization_quality
    
    UNION ALL
    
    -- Add "All Developers" summary
    SELECT 
        'All Developers' as vendor_name,
        has_valid_identifiers,
        identifier_status,
        has_valid_name,
        has_valid_address,
        overall_quality_score,
        identifier_conformance_category,
        identifier_conformance_rate
    FROM mv_organization_quality
) vendor_expanded
GROUP BY vendor_name;

-- Create index
CREATE UNIQUE INDEX idx_mv_org_quality_summary_complete_vendor ON mv_organization_quality_summary(vendor_name);

-- 3. Update identifier summary view
CREATE MATERIALIZED VIEW mv_organization_identifier_summary AS  
SELECT 
    vendor_name,
    SUM(npi_count) as total_npi,
    SUM(clia_count) as total_clia,
    SUM(naic_count) as total_naic,
    SUM(other_count) as total_other,
    SUM(CASE WHEN total_identifier_count = 0 THEN 1 ELSE 0 END) as total_no_identifiers,
    SUM(npi_valid) as total_npi_valid,
    SUM(clia_valid) as total_clia_valid,
    SUM(naic_valid) as total_naic_valid,
    SUM(npi_invalid) as total_npi_invalid,
    SUM(clia_invalid) as total_clia_invalid,
    SUM(naic_invalid) as total_naic_invalid,
    SUM(other_invalid) as total_other_invalid,
    SUM(total_identifier_count) as total_all_identifiers,
    SUM(conformant_identifier_count) as total_all_conformant,
    
    -- Percentages
    CASE 
        WHEN SUM(total_identifier_count) > 0 THEN ROUND(SUM(npi_count)::DECIMAL / SUM(total_identifier_count)::DECIMAL * 100, 1)
        ELSE 0
    END as npi_percentage,
    CASE 
        WHEN SUM(total_identifier_count) > 0 THEN ROUND(SUM(clia_count)::DECIMAL / SUM(total_identifier_count)::DECIMAL * 100, 1) 
        ELSE 0
    END as clia_percentage,
    CASE 
        WHEN SUM(total_identifier_count) > 0 THEN ROUND(SUM(naic_count)::DECIMAL / SUM(total_identifier_count)::DECIMAL * 100, 1)
        ELSE 0  
    END as naic_percentage,
    CASE 
        WHEN SUM(total_identifier_count) > 0 THEN ROUND(SUM(other_count)::DECIMAL / SUM(total_identifier_count)::DECIMAL * 100, 1)
        ELSE 0
    END as other_percentage,
    CASE 
        WHEN SUM(total_identifier_count) > 0 THEN ROUND(SUM(conformant_identifier_count)::DECIMAL / SUM(total_identifier_count)::DECIMAL * 100, 1)
        ELSE 0
    END as conformance_rate
    
FROM (
    SELECT 
        UNNEST(vendor_names_array) as vendor_name,
        npi_count, clia_count, naic_count, other_count,
        npi_valid, clia_valid, naic_valid, npi_invalid, clia_invalid, naic_invalid, other_invalid,
        total_identifier_count, conformant_identifier_count
    FROM mv_organization_quality
    
    UNION ALL
    
    -- Add "All Developers" summary  
    SELECT 
        'All Developers' as vendor_name,
        npi_count, clia_count, naic_count, other_count,
        npi_valid, clia_valid, naic_valid, npi_invalid, clia_invalid, naic_invalid, other_invalid,
        total_identifier_count, conformant_identifier_count
    FROM mv_organization_quality
) vendor_expanded
GROUP BY vendor_name;

-- Create index
CREATE UNIQUE INDEX idx_mv_org_identifier_summary_complete_vendor ON mv_organization_identifier_summary(vendor_name);

COMMIT;