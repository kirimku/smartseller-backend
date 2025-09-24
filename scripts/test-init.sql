-- Test database initialization script
-- This script sets up the test database with required extensions and initial data

-- Enable required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create test users with known credentials for integration tests
-- Note: In production, these would be created through proper user registration flows

-- Test User (Regular User)
INSERT INTO users (
    id, 
    email, 
    password_hash, 
    first_name, 
    last_name, 
    role, 
    status, 
    email_verified_at,
    created_at, 
    updated_at
) VALUES (
    uuid_generate_v4(),
    'testuser@example.com',
    crypt('testpassword123', gen_salt('bf')),
    'Test',
    'User',
    'user',
    'active',
    NOW(),
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Admin User
INSERT INTO users (
    id, 
    email, 
    password_hash, 
    first_name, 
    last_name, 
    role, 
    status, 
    email_verified_at,
    created_at, 
    updated_at
) VALUES (
    uuid_generate_v4(),
    'admin@example.com',
    crypt('adminpassword123', gen_salt('bf')),
    'Admin',
    'User',
    'admin',
    'active',
    NOW(),
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Create test product categories
INSERT INTO product_categories (id, name, slug, description, created_at, updated_at) VALUES
(uuid_generate_v4(), 'Electronics', 'electronics', 'Electronic devices and accessories', NOW(), NOW()),
(uuid_generate_v4(), 'Clothing', 'clothing', 'Clothing and fashion items', NOW(), NOW()),
(uuid_generate_v4(), 'Books', 'books', 'Books and educational materials', NOW(), NOW()),
(uuid_generate_v4(), 'Sports', 'sports', 'Sports and outdoor equipment', NOW(), NOW()),
(uuid_generate_v4(), 'Home & Garden', 'home-garden', 'Home and garden supplies', NOW(), NOW())
ON CONFLICT (slug) DO NOTHING;

-- Create subcategories
WITH parent_categories AS (
    SELECT id, slug FROM product_categories WHERE parent_id IS NULL
)
INSERT INTO product_categories (id, name, slug, description, parent_id, created_at, updated_at)
SELECT 
    uuid_generate_v4(),
    subcategory.name,
    subcategory.slug,
    subcategory.description,
    pc.id,
    NOW(),
    NOW()
FROM parent_categories pc
CROSS JOIN (VALUES
    ('Smartphones', 'smartphones', 'Mobile phones and accessories', 'electronics'),
    ('Laptops', 'laptops', 'Laptops and computers', 'electronics'),
    ('Headphones', 'headphones', 'Audio equipment and headphones', 'electronics'),
    ('Men''s Clothing', 'mens-clothing', 'Clothing for men', 'clothing'),
    ('Women''s Clothing', 'womens-clothing', 'Clothing for women', 'clothing'),
    ('Shoes', 'shoes', 'Footwear for all ages', 'clothing'),
    ('Fiction', 'fiction', 'Fiction books and novels', 'books'),
    ('Non-Fiction', 'non-fiction', 'Educational and reference books', 'books'),
    ('Children''s Books', 'childrens-books', 'Books for children', 'books'),
    ('Fitness', 'fitness', 'Fitness and exercise equipment', 'sports'),
    ('Outdoor', 'outdoor', 'Outdoor and camping gear', 'sports'),
    ('Furniture', 'furniture', 'Home furniture', 'home-garden'),
    ('Gardening', 'gardening', 'Gardening tools and supplies', 'home-garden')
) AS subcategory(name, slug, description, parent_slug)
WHERE pc.slug = subcategory.parent_slug
ON CONFLICT (slug) DO NOTHING;

-- Create test variant options for common product types
INSERT INTO product_variant_options (id, name, values, display_name, sort_order, required, created_at, updated_at) VALUES
(uuid_generate_v4(), 'color', '["Red", "Blue", "Green", "Black", "White", "Yellow", "Purple", "Orange"]', 'Color', 1, true, NOW(), NOW()),
(uuid_generate_v4(), 'size', '["XS", "S", "M", "L", "XL", "XXL"]', 'Size', 2, true, NOW(), NOW()),
(uuid_generate_v4(), 'storage', '["64GB", "128GB", "256GB", "512GB", "1TB"]', 'Storage', 3, false, NOW(), NOW()),
(uuid_generate_v4(), 'material', '["Cotton", "Polyester", "Leather", "Wool", "Silk", "Denim"]', 'Material', 4, false, NOW(), NOW()),
(uuid_generate_v4(), 'style', '["Casual", "Formal", "Sport", "Business", "Party"]', 'Style', 5, false, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Insert sample products for testing (will be cleaned up between tests)
WITH category_ids AS (
    SELECT id, slug FROM product_categories WHERE slug IN ('smartphones', 'laptops', 'mens-clothing', 'womens-clothing', 'fiction')
)
INSERT INTO products (
    id, 
    name, 
    sku, 
    description, 
    price, 
    cost_price,
    category_id, 
    status, 
    weight,
    dimensions,
    stock_quantity,
    stock_threshold,
    created_at, 
    updated_at
)
SELECT 
    uuid_generate_v4(),
    p.name,
    p.sku,
    p.description,
    p.price,
    p.cost_price,
    c.id,
    'active',
    p.weight,
    jsonb_build_object('length', p.length, 'width', p.width, 'height', p.height),
    p.stock_quantity,
    p.stock_threshold,
    NOW(),
    NOW()
FROM category_ids c
CROSS JOIN (VALUES
    ('iPhone 15 Pro', 'IPHONE-15-PRO', 'Latest iPhone with advanced features', 999.99, 750.00, 200.0, 14.7, 7.1, 0.8, 50, 10, 'smartphones'),
    ('Samsung Galaxy S24', 'GALAXY-S24', 'Samsung flagship smartphone', 849.99, 650.00, 190.0, 14.6, 7.0, 0.8, 30, 5, 'smartphones'),
    ('MacBook Pro M3', 'MBP-M3-14', 'Professional laptop with M3 chip', 1999.99, 1500.00, 1600.0, 31.3, 22.1, 1.6, 20, 3, 'laptops'),
    ('Dell XPS 13', 'DELL-XPS13', 'Ultrabook for professionals', 1299.99, 950.00, 1200.0, 29.6, 19.9, 1.8, 15, 2, 'laptops'),
    ('Men''s Cotton T-Shirt', 'MENS-TSHIRT-001', 'Comfortable cotton t-shirt', 29.99, 15.00, 150.0, 28.0, 20.0, 1.0, 100, 20, 'mens-clothing'),
    ('Men''s Jeans', 'MENS-JEANS-001', 'Classic blue jeans', 79.99, 45.00, 600.0, 32.0, 12.0, 4.0, 75, 15, 'mens-clothing'),
    ('Women''s Dress', 'WOMENS-DRESS-001', 'Elegant summer dress', 89.99, 50.00, 300.0, 95.0, 25.0, 2.0, 40, 8, 'womens-clothing'),
    ('The Great Gatsby', 'BOOK-GATSBY', 'Classic American novel', 12.99, 7.00, 200.0, 18.0, 11.0, 2.0, 200, 25, 'fiction'),
    ('To Kill a Mockingbird', 'BOOK-MOCKINGBIRD', 'Pulitzer Prize winning novel', 14.99, 8.50, 180.0, 17.5, 10.5, 1.8, 150, 20, 'fiction')
) AS p(name, sku, description, price, cost_price, weight, length, width, height, stock_quantity, stock_threshold, category_slug)
WHERE c.slug = p.category_slug
ON CONFLICT (sku) DO NOTHING;

-- Create some product variants for testing
WITH product_data AS (
    SELECT p.id as product_id, vo.id as option_id, vo.values, p.sku
    FROM products p
    CROSS JOIN product_variant_options vo
    WHERE p.sku IN ('IPHONE-15-PRO', 'MENS-TSHIRT-001', 'MENS-JEANS-001')
    AND ((p.sku = 'IPHONE-15-PRO' AND vo.name IN ('color', 'storage'))
         OR (p.sku IN ('MENS-TSHIRT-001', 'MENS-JEANS-001') AND vo.name IN ('color', 'size')))
)
INSERT INTO product_variants (
    id,
    product_id,
    sku,
    name,
    options,
    price_adjustment,
    cost_price_adjustment,
    stock_quantity,
    weight,
    dimensions,
    is_default,
    created_at,
    updated_at
) VALUES
-- iPhone variants
(uuid_generate_v4(), 
 (SELECT id FROM products WHERE sku = 'IPHONE-15-PRO'), 
 'IPHONE-15-PRO-BLK-256', 
 'iPhone 15 Pro - Black 256GB', 
 '{"color": "Black", "storage": "256GB"}', 
 0.00, 
 0.00, 
 25, 
 200.0,
 '{"length": 14.7, "width": 7.1, "height": 0.8}',
 true,
 NOW(), 
 NOW()),
(uuid_generate_v4(), 
 (SELECT id FROM products WHERE sku = 'IPHONE-15-PRO'), 
 'IPHONE-15-PRO-WHT-512', 
 'iPhone 15 Pro - White 512GB', 
 '{"color": "White", "storage": "512GB"}', 
 200.00, 
 150.00, 
 15, 
 200.0,
 '{"length": 14.7, "width": 7.1, "height": 0.8}',
 false,
 NOW(), 
 NOW()),
-- T-shirt variants
(uuid_generate_v4(), 
 (SELECT id FROM products WHERE sku = 'MENS-TSHIRT-001'), 
 'MENS-TSHIRT-001-BLU-M', 
 'Men''s Cotton T-Shirt - Blue M', 
 '{"color": "Blue", "size": "M"}', 
 0.00, 
 0.00, 
 50, 
 150.0,
 '{"length": 28.0, "width": 20.0, "height": 1.0}',
 true,
 NOW(), 
 NOW()),
(uuid_generate_v4(), 
 (SELECT id FROM products WHERE sku = 'MENS-TSHIRT-001'), 
 'MENS-TSHIRT-001-RED-L', 
 'Men''s Cotton T-Shirt - Red L', 
 '{"color": "Red", "size": "L"}', 
 0.00, 
 0.00, 
 30, 
 150.0,
 '{"length": 28.0, "width": 20.0, "height": 1.0}',
 false,
 NOW(), 
 NOW())
ON CONFLICT (sku) DO NOTHING;

-- Create some product images (placeholder URLs for testing)
WITH product_images AS (
    SELECT id, sku FROM products WHERE sku IN ('IPHONE-15-PRO', 'MENS-TSHIRT-001', 'MBP-M3-14')
)
INSERT INTO product_images (
    id,
    product_id,
    variant_id,
    image_url,
    cloudinary_public_id,
    alt_text,
    is_primary,
    sort_order,
    file_size,
    width,
    height,
    format,
    created_at,
    updated_at
) VALUES
-- iPhone images
(uuid_generate_v4(), (SELECT id FROM products WHERE sku = 'IPHONE-15-PRO'), NULL, 'https://example.com/images/iphone-15-pro-1.jpg', 'iphone-15-pro-1', 'iPhone 15 Pro front view', true, 1, 245760, 800, 600, 'jpg', NOW(), NOW()),
(uuid_generate_v4(), (SELECT id FROM products WHERE sku = 'IPHONE-15-PRO'), NULL, 'https://example.com/images/iphone-15-pro-2.jpg', 'iphone-15-pro-2', 'iPhone 15 Pro back view', false, 2, 198432, 800, 600, 'jpg', NOW(), NOW()),
-- T-shirt images
(uuid_generate_v4(), (SELECT id FROM products WHERE sku = 'MENS-TSHIRT-001'), NULL, 'https://example.com/images/mens-tshirt-1.jpg', 'mens-tshirt-1', 'Men''s cotton t-shirt', true, 1, 156789, 600, 800, 'jpg', NOW(), NOW()),
-- MacBook images
(uuid_generate_v4(), (SELECT id FROM products WHERE sku = 'MBP-M3-14'), NULL, 'https://example.com/images/macbook-pro-m3.jpg', 'macbook-pro-m3', 'MacBook Pro with M3 chip', true, 1, 312456, 1024, 768, 'jpg', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Create indexes for better test performance
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
CREATE INDEX IF NOT EXISTS idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX IF NOT EXISTS idx_product_variants_sku ON product_variants(sku);
CREATE INDEX IF NOT EXISTS idx_product_images_product_id ON product_images(product_id);
CREATE INDEX IF NOT EXISTS idx_product_categories_slug ON product_categories(slug);
CREATE INDEX IF NOT EXISTS idx_product_categories_parent_id ON product_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Update table statistics for better query planning during tests
ANALYZE products;
ANALYZE product_categories;
ANALYZE product_variants;
ANALYZE product_images;
ANALYZE product_variant_options;
ANALYZE users;
