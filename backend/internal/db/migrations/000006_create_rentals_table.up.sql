CREATE TABLE IF NOT EXISTS rentals (
    id SERIAL PRIMARY KEY,
    client_name TEXT NOT NULL,
    equipment TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    billing_type TEXT NOT NULL
);