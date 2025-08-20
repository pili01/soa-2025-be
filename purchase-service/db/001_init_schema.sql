-- Create shopping_carts table
CREATE TABLE IF NOT EXISTS shopping_carts (
    id SERIAL PRIMARY KEY,
    tourist_id INTEGER NOT NULL,
    total_price DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create order_items table
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    cart_id INTEGER REFERENCES shopping_carts(id) ON DELETE CASCADE,
    tour_id INTEGER NOT NULL,
    tour_name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    quantity INTEGER DEFAULT 1,
    added_at TIMESTAMP DEFAULT NOW()
);

-- Create tour_purchase_tokens table
CREATE TABLE IF NOT EXISTS tour_purchase_tokens (
    id SERIAL PRIMARY KEY,
    tourist_id INTEGER NOT NULL,
    tour_id INTEGER NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    purchased_at TIMESTAMP DEFAULT NOW()
);

