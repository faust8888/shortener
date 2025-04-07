CREATE TABLE shortener (
    id SERIAL PRIMARY KEY,
    short_url VARCHAR(255) NOT NULL,
    full_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX unique_full_url_index ON shortener (full_url);