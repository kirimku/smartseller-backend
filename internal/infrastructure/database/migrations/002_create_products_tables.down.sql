-- Drop product-related tables, triggers, and functions

-- Drop triggers
DROP TRIGGER IF EXISTS validate_variant_options_trigger ON product_variants;
DROP TRIGGER IF EXISTS set_variant_name_trigger ON product_variants;
DROP TRIGGER IF EXISTS update_product_variants_updated_at ON product_variants;
DROP TRIGGER IF EXISTS update_product_variant_options_updated_at ON product_variant_options;
DROP TRIGGER IF EXISTS update_product_categories_updated_at ON product_categories;
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

-- Drop functions
DROP FUNCTION IF EXISTS validate_variant_options();
DROP FUNCTION IF EXISTS set_variant_name();
DROP FUNCTION IF EXISTS generate_variant_name(JSONB);

-- Drop tables
DROP TABLE IF EXISTS product_variants CASCADE;
DROP TABLE IF EXISTS product_variant_options CASCADE;
DROP TABLE IF EXISTS product_images CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS product_categories CASCADE;
