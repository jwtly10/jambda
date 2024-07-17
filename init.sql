-- Create the database if it does not exist
CREATE DATABASE jambda_db WITH OWNER = postgres ENCODING = 'UTF8' CONNECTION
LIMIT
    = -1 TEMPLATE template0;

-- Connect to the new database
\ c jambda_db;

-- Create the table if it does not exist
CREATE TABLE IF NOT EXISTS functions_tb (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    external_id VARCHAR(8) NOT NULL UNIQUE,
    state VARCHAR(10) NOT NULL,
    configuration JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);