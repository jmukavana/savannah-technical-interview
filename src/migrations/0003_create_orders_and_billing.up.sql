-- Orders, Order items, Invoices, Payments
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    customer_id UUID REFERENCES customers(id) ON DELETE
    SET NULL,
        status VARCHAR(30) NOT NULL DEFAULT 'CREATED',
        -- CREATED, CONFIRMED, SHIPPED, CANCELLED, COMPLETED
        subtotal NUMERIC(18, 4) NOT NULL,
        tax NUMERIC(18, 4) NOT NULL DEFAULT 0,
        shipping NUMERIC(18, 4) NOT NULL DEFAULT 0,
        total NUMERIC(18, 4) NOT NULL,
        currency CHAR(3) NOT NULL DEFAULT 'USD',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        version INT NOT NULL DEFAULT 1
);
CREATE TABLE order_items (
    id UUID PRIMARY KEY,
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id) ON DELETE
    SET NULL,
        sku VARCHAR(64),
        name VARCHAR(255),
        unit_price NUMERIC(18, 4) NOT NULL,
        quantity INT NOT NULL,
        line_total NUMERIC(18, 4) NOT NULL
);
CREATE TABLE invoices (
    id UUID PRIMARY KEY,
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    invoice_number VARCHAR(64) UNIQUE NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'UNPAID',
    -- UNPAID, PAID, VOID, REFUNDED
    amount NUMERIC(18, 4) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ
);
CREATE TABLE payments (
    id UUID PRIMARY KEY,
    invoice_id UUID REFERENCES invoices(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_payment_id VARCHAR(255),
    amount NUMERIC(18, 4) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'KES',
    status VARCHAR(30) NOT NULL DEFAULT 'PROCESSING',
    -- PROCESSING, SUCCESS, FAILED
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Optional indexes
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_inventory_product ON inventory(product_id);
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_invoices_order ON invoices(order_id);