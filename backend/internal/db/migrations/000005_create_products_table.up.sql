CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    detailed_name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL CHECK (status IN ('ativo', 'desativado', 'descontinuado')),
    sku TEXT,
    barcode TEXT,
    external_id TEXT,
    coin TEXT NOT NULL CHECK (coin IN ('BRL', 'USD', 'EUR', 'CAD', 'ADOBE_USD')),
    price NUMERIC NOT NULL CHECK (price >= 0),
    sales_price NUMERIC CHECK (sales_price >= 0),
    cost_price NUMERIC CHECK (cost_price >= 0),
    stock INT NOT NULL CHECK (stock >= 0),
    type TEXT,
    product_group TEXT,
    product_category TEXT,
    product_subcategory TEXT,
    tags TEXT[], 
    manufacturer TEXT,
    manufacturer_code TEXT,
    ncm TEXT,
    cest TEXT,
    cnae TEXT,
    origin TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP, -- Adiciona suporte para Soft Delete
    images TEXT[], 
    documents TEXT[] 
);