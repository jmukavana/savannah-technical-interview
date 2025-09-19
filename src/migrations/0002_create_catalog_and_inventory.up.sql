CREATE TABLE
    categories (
        id UUID PRIMARY KEY,
        name VARCHAR(150) NOT NULL,
        slug VARCHAR(160) NOT NULL UNIQUE,
        description TEXT,
        parent_id UUID REFERENCES categories (id) ON DELETE SET NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE TABLE
    products (
        id UUID PRIMARY KEY,
        sku VARCHAR(64) NOT NULL UNIQUE,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        category_id UUID REFERENCES categories (id) ON DELETE SET NULL,
        price NUMERIC(18, 4) NOT NULL,
        currency CHAR(3) NOT NULL DEFAULT 'KES',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE TABLE
    inventory (
        id UUID PRIMARY KEY,
        product_id UUID REFERENCES products (id) ON DELETE CASCADE,
        warehouse VARCHAR(100) NOT NULL,
        quantity INT NOT NULL DEFAULT 0,
        reserved INT NOT NULL DEFAULT 0,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (product_id, warehouse)
    );

CREATE TABLE
    stock_transactions (
        id UUID PRIMARY KEY,
        inventory_id UUID REFERENCES inventory (id) ON DELETE CASCADE,
        change INT NOT NULL,
        reason VARCHAR(255) NOT NULL,
        reference VARCHAR(255),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );