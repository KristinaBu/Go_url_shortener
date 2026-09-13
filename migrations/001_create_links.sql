CREATE TABLE IF NOT EXISTS links (
    short_code VARCHAR(10) PRIMARY KEY,
    original_url TEXT NOT NULL UNIQUE
    );