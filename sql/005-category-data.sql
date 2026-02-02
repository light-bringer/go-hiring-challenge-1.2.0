-- Insert 3 categories
INSERT INTO categories (code, name) VALUES
    ('clothing', 'Clothing'),
    ('shoes', 'Shoes'),
    ('accessories', 'Accessories')
ON CONFLICT (code) DO NOTHING;

-- Update products with categories
UPDATE products SET category_id = (SELECT id FROM categories WHERE code = 'clothing')
WHERE code IN ('PROD001', 'PROD004', 'PROD007')
    AND category_id IS NULL;

UPDATE products SET category_id = (SELECT id FROM categories WHERE code = 'shoes')
WHERE code IN ('PROD002', 'PROD006')
    AND category_id IS NULL;

UPDATE products SET category_id = (SELECT id FROM categories WHERE code = 'accessories')
WHERE code IN ('PROD003', 'PROD005', 'PROD008')
    AND category_id IS NULL;
