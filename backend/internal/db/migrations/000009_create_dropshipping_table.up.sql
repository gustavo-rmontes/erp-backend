CREATE TABLE IF NOT EXISTS dropshipping (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL,
    warranty_id INTEGER,
    cliente VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL CHECK (price > 0),
    quantity INTEGER NOT NULL CHECK (quantity >= 0),
    total_price DECIMAL(10, 2),
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_dropshipping_product ON dropshipping (product_id);
CREATE INDEX IF NOT EXISTS idx_dropshipping_cliente ON dropshipping (cliente);
CREATE INDEX IF NOT EXISTS idx_dropshipping_product_id ON dropshipping (product_id);
CREATE INDEX IF NOT EXISTS idx_dropshipping_warranty_id ON dropshipping (warranty_id);