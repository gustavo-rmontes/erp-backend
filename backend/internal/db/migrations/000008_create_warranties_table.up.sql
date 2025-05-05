CREATE TABLE IF NOT EXISTS warranties (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL,
    duration_months INT NOT NULL,
    price NUMERIC(10, 2) NOT NULL CHECK (price > 0),
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);