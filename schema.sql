CREATE TABLE functions_tb
(
    id            SERIAL PRIMARY KEY,
    external_id   VARCHAR(8)   NOT NULL,
    state         VARCHAR(10)  NOT NULL,
    configuration JSONB,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

