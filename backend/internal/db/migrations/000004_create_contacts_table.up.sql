CREATE TABLE IF NOT EXISTS contacts (
    id SERIAL PRIMARY KEY,
    person_type VARCHAR(2) NOT NULL CHECK (person_type IN ('pf', 'pj')),
    type VARCHAR(20) NOT NULL CHECK (type IN ('cliente', 'fornecedor', 'lead')),
    name VARCHAR(100) NOT NULL,
    company_name VARCHAR(150),
    trade_name VARCHAR(150),
    document VARCHAR(20) NOT NULL,
    secondary_doc VARCHAR(20),
    suframa VARCHAR(20),
    isento BOOLEAN DEFAULT FALSE,
    ccm VARCHAR(20),
    email VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    zip_code VARCHAR(10),
    street VARCHAR(150),
    number VARCHAR(20),
    complement VARCHAR(100),
    neighborhood VARCHAR(100),
    city VARCHAR(100),
    state VARCHAR(2),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
