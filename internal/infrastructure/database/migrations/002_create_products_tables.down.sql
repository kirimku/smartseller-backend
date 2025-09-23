-- Drop product-related tables, triggers, and functions

-- Drop view
DROP VIEW IF EXISTS product_variants_with_pricing;

-- Drop triggers
DROP TRIGGER IF EXISTS validate_variant_options_trigger ON product_variants;
DROP TRIGGER IF EXISTS set_variant_name_trigger ON product_variants;

-- Drop functions
DROP FUNCTION IF EXISTS get_effective_price(DECIMAL(15,2), DECIMAL(15,2), DECIMAL(15,2), DECIMAL(15,2));
DROP FUNCTION IF EXISTS validate_variant_options();
DROP FUNCTION IF EXISTS set_variant_name();
DROP FUNCTION IF EXISTS generate_variant_name(JSONB);

-- Drop tables
DROP TABLE IF EXISTS product_variants CASCADE;
DROP TABLE IF EXISTS product_variant_options CASCADE;
DROP TABLE IF EXISTS product_images CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS product_categories CASCADE;
