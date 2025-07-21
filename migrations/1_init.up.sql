CREATE TYPE skill AS ENUM(
    'backend',
    'frontend',
    'go',
    'docker'
);

CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,
    email VARCHAR(50) UNIQUE,
    name VARCHAR(50),
    password TEXT,
    about TEXT,
    skills skill[],
    avatar_url TEXT,
    is_email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);