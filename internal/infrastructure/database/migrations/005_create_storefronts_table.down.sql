-- Drop storefronts table and related objects
DROP TRIGGER IF EXISTS update_storefronts_updated_at ON storefronts;
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS validate_slug(TEXT);
DROP TABLE IF EXISTS storefronts CASCADE;