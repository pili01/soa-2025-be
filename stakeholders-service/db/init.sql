CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('Admin', 'Guide', 'Tourist')),
    name VARCHAR(255),
    surname VARCHAR(255),
    biography TEXT,
    moto VARCHAR(255),
    photo_url VARCHAR(255),
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS positions (
    id SERIAL PRIMARY KEY,
    user_id INT UNIQUE NOT NULL,
    longitude FLOAT NOT NULL,
    latitude FLOAT NOT NULL
);