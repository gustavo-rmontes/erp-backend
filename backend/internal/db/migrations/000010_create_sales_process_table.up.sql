-- Database schema for the sales process

-- Quotations Table
CREATE TABLE IF NOT EXISTS quotations (
    id SERIAL PRIMARY KEY,
    quotation_no VARCHAR(50) NOT NULL UNIQUE,
    contact_id INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry_date TIMESTAMP NOT NULL,
    subtotal DECIMAL(12, 2) NOT NULL,
    tax_total DECIMAL(12, 2) NOT NULL,
    discount_total DECIMAL(12, 2) NOT NULL,
    grand_total DECIMAL(12, 2) NOT NULL,
    notes TEXT,
    terms TEXT,
    CONSTRAINT valid_quotation_status CHECK (status IN ('draft', 'sent', 'accepted', 'rejected', 'expired', 'cancelled'))
);

-- Quotation Items Table
CREATE TABLE IF NOT EXISTS quotation_items (
    id SERIAL PRIMARY KEY,
    quotation_id INTEGER NOT NULL REFERENCES quotations(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    product_code VARCHAR(50),
    description TEXT,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(12, 2) NOT NULL,
    discount DECIMAL(12, 2) DEFAULT 0,
    tax DECIMAL(12, 2) DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL
);

-- Sales Orders Table
CREATE TABLE IF NOT EXISTS sales_orders (
    id SERIAL PRIMARY KEY,
    so_no VARCHAR(50) NOT NULL UNIQUE,
    quotation_id INTEGER REFERENCES quotations(id),
    contact_id INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expected_date TIMESTAMP,
    subtotal DECIMAL(12, 2) NOT NULL,
    tax_total DECIMAL(12, 2) NOT NULL,
    discount_total DECIMAL(12, 2) NOT NULL,
    grand_total DECIMAL(12, 2) NOT NULL,
    notes TEXT,
    payment_terms TEXT,
    shipping_address TEXT,
    CONSTRAINT valid_so_status CHECK (status IN ('draft', 'confirmed', 'processing', 'completed', 'cancelled'))
);

-- Sales Order Items Table
CREATE TABLE IF NOT EXISTS sales_order_items (
    id SERIAL PRIMARY KEY,
    sales_order_id INTEGER NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    product_code VARCHAR(50),
    description TEXT,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(12, 2) NOT NULL,
    discount DECIMAL(12, 2) DEFAULT 0,
    tax DECIMAL(12, 2) DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL
);

-- Purchase Orders Table
CREATE TABLE IF NOT EXISTS purchase_orders (
    id SERIAL PRIMARY KEY,
    po_no VARCHAR(50) NOT NULL UNIQUE,
    so_no VARCHAR(50),
    sales_order_id INTEGER REFERENCES sales_orders(id),
    contact_id INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expected_date TIMESTAMP,
    subtotal DECIMAL(12, 2) NOT NULL,
    tax_total DECIMAL(12, 2) NOT NULL,
    discount_total DECIMAL(12, 2) NOT NULL,
    grand_total DECIMAL(12, 2) NOT NULL,
    notes TEXT,
    payment_terms TEXT,
    shipping_address TEXT,
    CONSTRAINT valid_po_status CHECK (status IN ('draft', 'sent', 'confirmed', 'received', 'cancelled'))
);

-- Purchase Order Items Table
CREATE TABLE IF NOT EXISTS purchase_order_items (
    id SERIAL PRIMARY KEY,
    purchase_order_id INTEGER NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    product_code VARCHAR(50),
    description TEXT,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(12, 2) NOT NULL,
    discount DECIMAL(12, 2) DEFAULT 0,
    tax DECIMAL(12, 2) DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL
);

-- Deliveries Table
CREATE TABLE IF NOT EXISTS deliveries (
    id SERIAL PRIMARY KEY,
    delivery_no VARCHAR(50) NOT NULL UNIQUE,
    purchase_order_id INTEGER REFERENCES purchase_orders(id),
    po_no VARCHAR(50),
    sales_order_id INTEGER REFERENCES sales_orders(id),
    so_no VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivery_date TIMESTAMP,
    received_date TIMESTAMP,
    shipping_method VARCHAR(100),
    tracking_number VARCHAR(100),
    shipping_address TEXT,
    notes TEXT,
    CONSTRAINT valid_delivery_status CHECK (status IN ('pending', 'shipped', 'delivered', 'returned'))
);

-- Delivery Items Table
CREATE TABLE IF NOT EXISTS delivery_items (
    id SERIAL PRIMARY KEY,
    delivery_id INTEGER NOT NULL REFERENCES deliveries(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    product_code VARCHAR(50),
    description TEXT,
    quantity INTEGER NOT NULL,
    received_qty INTEGER DEFAULT 0,
    notes TEXT
);

-- Invoices Table
CREATE TABLE IF NOT EXISTS invoices (
    id SERIAL PRIMARY KEY,
    invoice_no VARCHAR(50) NOT NULL UNIQUE,
    sales_order_id INTEGER REFERENCES sales_orders(id),
    so_no VARCHAR(50),
    contact_id INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    issue_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    due_date TIMESTAMP NOT NULL,
    subtotal DECIMAL(12, 2) NOT NULL,
    tax_total DECIMAL(12, 2) NOT NULL,
    discount_total DECIMAL(12, 2) NOT NULL,
    grand_total DECIMAL(12, 2) NOT NULL,
    amount_paid DECIMAL(12, 2) DEFAULT 0,
    payment_terms TEXT,
    notes TEXT,
    CONSTRAINT valid_invoice_status CHECK (status IN ('draft', 'sent', 'partial', 'paid', 'overdue', 'cancelled'))
);

-- Invoice Items Table
CREATE TABLE IF NOT EXISTS invoice_items (
    id SERIAL PRIMARY KEY,
    invoice_id INTEGER NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    product_code VARCHAR(50),
    description TEXT,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(12, 2) NOT NULL,
    discount DECIMAL(12, 2) DEFAULT 0,
    tax DECIMAL(12, 2) DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL
);

-- Payments Table
CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    invoice_id INTEGER NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount DECIMAL(12, 2) NOT NULL,
    payment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    payment_method VARCHAR(50),
    reference VARCHAR(100),
    notes TEXT
);

-- Sales Process Table (linking all documents)
CREATE TABLE IF NOT EXISTS sales_processes (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total_value DECIMAL(12, 2) DEFAULT 0,
    profit DECIMAL(12, 2) DEFAULT 0,
    notes TEXT
);

-- Link tables for the sales process

-- Link table between sales_processes and quotations
CREATE TABLE IF NOT EXISTS process_quotations (
    process_id INTEGER NOT NULL REFERENCES sales_processes(id) ON DELETE CASCADE,
    quotation_id INTEGER NOT NULL REFERENCES quotations(id) ON DELETE CASCADE,
    PRIMARY KEY (process_id, quotation_id)
);

-- Link table between sales_processes and sales orders
CREATE TABLE IF NOT EXISTS process_sales_orders (
    process_id INTEGER NOT NULL REFERENCES sales_processes(id) ON DELETE CASCADE,
    sales_order_id INTEGER NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    PRIMARY KEY (process_id, sales_order_id)
);

-- Link table between sales_processes and purchase orders
CREATE TABLE IF NOT EXISTS process_purchase_orders (
    process_id INTEGER NOT NULL REFERENCES sales_processes(id) ON DELETE CASCADE,
    purchase_order_id INTEGER NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    PRIMARY KEY (process_id, purchase_order_id)
);

-- Link table between sales_processes and deliveries
CREATE TABLE IF NOT EXISTS process_deliveries (
    process_id INTEGER NOT NULL REFERENCES sales_processes(id) ON DELETE CASCADE,
    delivery_id INTEGER NOT NULL REFERENCES deliveries(id) ON DELETE CASCADE,
    PRIMARY KEY (process_id, delivery_id)
);

-- Link table between sales_processes and invoices
CREATE TABLE IF NOT EXISTS process_invoices (
    process_id INTEGER NOT NULL REFERENCES sales_processes(id) ON DELETE CASCADE,
    invoice_id INTEGER NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    PRIMARY KEY (process_id, invoice_id)
);