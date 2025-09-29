-- Seed Product Categories
-- This script creates a comprehensive hierarchical category structure for testing

-- Clear existing categories (optional - uncomment if you want to reset)
-- DELETE FROM product_categories;

-- Root Categories (Level 1)
INSERT INTO product_categories (id, name, description, slug, parent_id, sort_order, is_active) VALUES
('11111111-1111-1111-1111-111111111111', 'Electronics', 'Electronic devices and accessories', 'electronics', NULL, 1, true),
('22222222-2222-2222-2222-222222222222', 'Clothing & Fashion', 'Apparel and fashion accessories', 'clothing-fashion', NULL, 2, true),
('33333333-3333-3333-3333-333333333333', 'Home & Garden', 'Home improvement and garden supplies', 'home-garden', NULL, 3, true),
('44444444-4444-4444-4444-444444444444', 'Sports & Outdoors', 'Sports equipment and outdoor gear', 'sports-outdoors', NULL, 4, true),
('55555555-5555-5555-5555-555555555555', 'Books & Media', 'Books, movies, music and digital media', 'books-media', NULL, 5, true),
('66666666-6666-6666-6666-666666666666', 'Health & Beauty', 'Health products and beauty items', 'health-beauty', NULL, 6, true),
('77777777-7777-7777-7777-777777777777', 'Automotive', 'Car parts and automotive accessories', 'automotive', NULL, 7, true),
('88888888-8888-8888-8888-888888888888', 'Toys & Games', 'Toys, games and entertainment', 'toys-games', NULL, 8, true);

-- Electronics Subcategories (Level 2)
INSERT INTO product_categories (id, name, description, slug, parent_id, sort_order, is_active) VALUES
('11111111-1111-1111-1111-111111111112', 'Smartphones', 'Mobile phones and accessories', 'smartphones', '11111111-1111-1111-1111-111111111111', 1, true),
('11111111-1111-1111-1111-111111111113', 'Laptops & Computers', 'Laptops, desktops and computer accessories', 'laptops-computers', '11111111-1111-1111-1111-111111111111', 2, true),
('11111111-1111-1111-1111-111111111114', 'Audio & Video', 'Headphones, speakers, cameras and video equipment', 'audio-video', '11111111-1111-1111-1111-111111111111', 3, true),
('11111111-1111-1111-1111-111111111115', 'Gaming', 'Gaming consoles, games and accessories', 'gaming', '11111111-1111-1111-1111-111111111111', 4, true),
('11111111-1111-1111-1111-111111111116', 'Smart Home', 'Smart home devices and IoT products', 'smart-home', '11111111-1111-1111-1111-111111111111', 5, true);

-- Clothing & Fashion Subcategories (Level 2)
INSERT INTO product_categories (id, name, description, slug, parent_id, sort_order, is_active) VALUES
('22222222-2222-2222-2222-222222222223', 'Men''s Clothing', 'Men''s apparel and accessories', 'mens-clothing', '22222222-2222-2222-2222-222222222222', 1, true),
('22222222-2222-2222-2222-222222222224', 'Women''s Clothing', 'Women''s apparel and accessories', 'womens-clothing', '22222222-2222-2222-2222-222222222222', 2, true),
('22222222-2222-2222-2222-222222222225', 'Kids & Baby', 'Children''s clothing and baby items', 'kids-baby', '22222222-2222-2222-2222-222222222222', 3, true),
('22222222-2222-2222-2222-222222222226', 'Shoes', 'Footwear for all ages', 'shoes', '22222222-2222-2222-2222-222222222222', 4, true),
('22222222-2222-2222-2222-222222222227', 'Accessories', 'Fashion accessories and jewelry', 'accessories', '22222222-2222-2222-2222-222222222222', 5, true);

-- Home & Garden Subcategories (Level 2)
INSERT INTO product_categories (id, name, description, slug, parent_id, sort_order, is_active) VALUES
('33333333-3333-3333-3333-333333333334', 'Furniture', 'Home and office furniture', 'furniture', '33333333-3333-3333-3333-333333333333', 1, true),
('33333333-3333-3333-3333-333333333335', 'Kitchen & Dining', 'Kitchen appliances and dining accessories', 'kitchen-dining', '33333333-3333-3333-3333-333333333333', 2, true),
('33333333-3333-3333-3333-333333333336', 'Home Decor', 'Decorative items and home accessories', 'home-decor', '33333333-3333-3333-3333-333333333333', 3, true),
('33333333-3333-3333-3333-333333333337', 'Garden & Outdoor', 'Gardening tools and outdoor furniture', 'garden-outdoor', '33333333-3333-3333-3333-333333333333', 4, true),
('33333333-3333-3333-3333-333333333338', 'Tools & Hardware', 'Tools and hardware supplies', 'tools-hardware', '33333333-3333-3333-3333-333333333333', 5, true);

-- Electronics Level 3 Categories (Smartphones subcategories)
INSERT INTO product_categories (id, name, description, slug, parent_id, sort_order, is_active) VALUES
('11111111-1111-1111-1111-111111111117', 'iPhone', 'Apple iPhone devices', 'iphone', '11111111-1111-1111-1111-111111111112', 1, true),
('11111111-1111-1111-1111-111111111118', 'Samsung Galaxy', 'Samsung Galaxy smartphones', 'samsung-galaxy', '11111111-1111-1111-1111-111111111112', 2, true),
('11111111-1111-1111-1111-111111111119', 'Phone Cases', 'Protective cases and covers', 'phone-cases', '11111111-1111-1111-1111-111111111112', 3, true),
('11111111-1111-1111-1111-11111111111A', 'Chargers & Cables', 'Charging accessories and cables', 'chargers-cables', '11111111-1111-1111-1111-111111111112', 4, true);

-- Electronics Level 3 Categories (Audio & Video subcategories)
INSERT INTO product_categories (id, name, description, slug, parent_id, sort_order, is_active) VALUES
('11111111-1111-1111-1111-11111111111B', 'Headphones', 'Wired and wireless headphones', 'headphones', '11111111-1111-1111-1111-111111111114', 1, true),
('11111111-1111-1111-1111-11111111111C', 'Speakers', 'Bluetooth and wired speakers', 'speakers', '11111111-1111-1111-1111-111111111114', 2, true),
('11111111-1111-1111-1111-11111111111D', 'Cameras', 'Digital cameras and accessories', 'cameras', '11111111-1111-1111-1111-111111111114', 3, true),
('11111111-1111-1111-1111-11111111111E', 'TV & Monitors', 'Televisions and computer monitors', 'tv-monitors', '11111111-1111-1111-1111-111111111114', 4, true);

-- Clothing Level 3 Categories (Men's Clothing subcategories)
INSERT INTO product_categories (id, name, description, slug, parent_id, sort_order, is_active) VALUES
('22222222-2222-2222-2222-222222222228', 'Men''s Shirts', 'Dress shirts, t-shirts and casual shirts', 'mens-shirts', '22222222-2222-2222-2222-222222222223', 1, true),
('22222222-2222-2222-2222-222222222229', 'Men''s Pants', 'Jeans, trousers and casual pants', 'mens-pants', '22222222-2222-2222-2222-222222222223', 2, true),
('22222222-2222-2222-2222-22222222222A', 'Men''s Jackets', 'Coats, jackets and outerwear', 'mens-jackets', '22222222-2222-2222-2222-222222222223', 3, true),
('22222222-2222-2222-2222-22222222222B', 'Men''s Underwear', 'Underwear and sleepwear', 'mens-underwear', '22222222-2222-2222-2222-222222222223', 4, true);

-- Clothing Level 3 Categories (Women's Clothing subcategories)
INSERT INTO product_categories (id, name, description, slug, parent_id, sort_order, is_active) VALUES
('22222222-2222-2222-2222-22222222222C', 'Women''s Dresses', 'Casual and formal dresses', 'womens-dresses', '22222222-2222-2222-2222-222222222224', 1, true),
('22222222-2222-2222-2222-22222222222D', 'Women''s Tops', 'Blouses, t-shirts and tank tops', 'womens-tops', '22222222-2222-2222-2222-222222222224', 2, true),
('22222222-2222-2222-2222-22222222222E', 'Women''s Bottoms', 'Jeans, skirts and pants', 'womens-bottoms', '22222222-2222-2222-2222-222222222224', 3, true),
('22222222-2222-2222-2222-22222222222F', 'Women''s Lingerie', 'Underwear and intimate apparel', 'womens-lingerie', '22222222-2222-2222-2222-222222222224', 4, true);

-- Sports & Outdoors Subcategories (Level 2)
INSERT INTO product_categories (id, name, description, slug, parent_id, sort_order, is_active) VALUES
('44444444-4444-4444-4444-444444444445', 'Fitness Equipment', 'Exercise and fitness gear', 'fitness-equipment', '44444444-4444-4444-4444-444444444444', 1, true),
('44444444-4444-4444-4444-444444444446', 'Outdoor Recreation', 'Camping, hiking and outdoor gear', 'outdoor-recreation', '44444444-4444-4444-4444-444444444444', 2, true),
('44444444-4444-4444-4444-444444444447', 'Team Sports', 'Equipment for team sports', 'team-sports', '44444444-4444-4444-4444-444444444444', 3, true),
('44444444-4444-4444-4444-444444444448', 'Water Sports', 'Swimming and water activity gear', 'water-sports', '44444444-4444-4444-4444-444444444444', 4, true);

-- Health & Beauty Subcategories (Level 2)
INSERT INTO product_categories (id, name, description, slug, parent_id, sort_order, is_active) VALUES
('66666666-6666-6666-6666-666666666667', 'Skincare', 'Skincare products and treatments', 'skincare', '66666666-6666-6666-6666-666666666666', 1, true),
('66666666-6666-6666-6666-666666666668', 'Makeup', 'Cosmetics and makeup products', 'makeup', '66666666-6666-6666-6666-666666666666', 2, true),
('66666666-6666-6666-6666-666666666669', 'Hair Care', 'Shampoo, conditioner and styling products', 'hair-care', '66666666-6666-6666-6666-666666666666', 3, true),
('66666666-6666-6666-6666-66666666666A', 'Health Supplements', 'Vitamins and health supplements', 'health-supplements', '66666666-6666-6666-6666-666666666666', 4, true);

-- Add some inactive categories for testing
INSERT INTO product_categories (id, name, description, slug, parent_id, sort_order, is_active) VALUES
('99999999-9999-9999-9999-999999999999', 'Discontinued Items', 'Products no longer available', 'discontinued-items', NULL, 99, false),
('99999999-9999-9999-9999-999999999998', 'Test Category', 'Category for testing purposes', 'test-category', '99999999-9999-9999-9999-999999999999', 1, false);

-- Update timestamps to show realistic creation times
UPDATE product_categories SET 
    created_at = CURRENT_TIMESTAMP - INTERVAL '30 days' + (RANDOM() * INTERVAL '30 days'),
    updated_at = CURRENT_TIMESTAMP - INTERVAL '7 days' + (RANDOM() * INTERVAL '7 days')
WHERE parent_id IS NULL;

UPDATE product_categories SET 
    created_at = CURRENT_TIMESTAMP - INTERVAL '20 days' + (RANDOM() * INTERVAL '20 days'),
    updated_at = CURRENT_TIMESTAMP - INTERVAL '5 days' + (RANDOM() * INTERVAL '5 days')
WHERE parent_id IS NOT NULL;

-- Display summary of seeded categories
SELECT 
    'Category Seeding Complete!' as status,
    COUNT(*) as total_categories,
    COUNT(CASE WHEN parent_id IS NULL THEN 1 END) as root_categories,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_categories,
    COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_categories
FROM product_categories;