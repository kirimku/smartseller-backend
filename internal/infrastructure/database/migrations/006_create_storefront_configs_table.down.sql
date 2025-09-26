-- Drop storefront_configs table
DROP TRIGGER IF EXISTS update_storefront_configs_updated_at ON storefront_configs;
DROP TABLE IF EXISTS storefront_configs CASCADE;