CREATE DATABASE `jambda-db` WITH OWNER = postgres ENCODING = 'UTF8' CONNECTION
LIMIT
    = -1 IS_TEMPLATE = False;

\ connect jambda_db;

CREATE TABLE functions_tb (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    external_id VARCHAR(8) NOT NULL UNIQUE,
    state VARCHAR(10) NOT NULL,
    configuration JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);