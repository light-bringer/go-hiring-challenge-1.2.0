-- Ensure products.code is non-null and unique for /catalog/{code}

-- Backfill missing or empty codes to avoid constraint failures
UPDATE products
SET code = CONCAT('unknown-', id)
WHERE code IS NULL OR code = '';

-- Enforce NOT NULL on code (idempotent)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_attribute a
        JOIN pg_class t ON a.attrelid = t.oid
        WHERE t.relname = 'products'
          AND a.attname = 'code'
          AND a.attnotnull = false
    ) THEN
        ALTER TABLE products ALTER COLUMN code SET NOT NULL;
    END IF;
END
$$;

-- Enforce UNIQUE on code (idempotent)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON c.conrelid = t.oid
        WHERE c.conname = 'products_code_unique'
          AND t.relname = 'products'
    ) THEN
        ALTER TABLE products ADD CONSTRAINT products_code_unique UNIQUE (code);
    END IF;
END
$$;
