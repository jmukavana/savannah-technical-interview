-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,
    
    CONSTRAINT check_name_length CHECK (LENGTH(name) >= 2),
    CONSTRAINT check_slug_length CHECK (LENGTH(slug) >= 2)
);

-- Create indexes for categories
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);
CREATE INDEX IF NOT EXISTS idx_categories_created_at ON categories(created_at);


-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    sku VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,
    
    CONSTRAINT check_product_name_length CHECK (LENGTH(name) >= 2),
    CONSTRAINT check_sku_length CHECK (LENGTH(sku) >= 1),
    CONSTRAINT check_currency_length CHECK (LENGTH(currency) = 3)
);

-- Create indexes for products
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);

-- Create GIN index for full-text search on products
CREATE INDEX IF NOT EXISTS idx_products_search ON products USING GIN (
    to_tsvector('english', name || ' ' || COALESCE(description, '') || ' ' || sku)
);

-- Create GIN index for full-text search on categories
CREATE INDEX IF NOT EXISTS idx_categories_search ON categories USING GIN (
    to_tsvector('english', name || ' ' || COALESCE(description, '') || ' ' || slug)
);

-- CREATE TABLE
--     inventory (
--         id UUID PRIMARY KEY,
--         product_id UUID REFERENCES products (id) ON DELETE CASCADE,
--         warehouse VARCHAR(100) NOT NULL,
--         quantity INT NOT NULL DEFAULT 0,
--         reserved INT NOT NULL DEFAULT 0,
--         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--         updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--         UNIQUE (product_id, warehouse)
--     );

-- CREATE TABLE
--     stock_transactions (
--         id UUID PRIMARY KEY,
--         inventory_id UUID REFERENCES inventory (id) ON DELETE CASCADE,
--         change INT NOT NULL,
--         reason VARCHAR(255) NOT NULL,
--         reference VARCHAR(255),
--         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
--     );