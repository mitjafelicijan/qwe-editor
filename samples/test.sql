-- Create a new table called 'users'
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT TRUE
);

-- Insert some sample data
INSERT INTO users (username, email) VALUES
    ('john_doe', 'john@example.com'),
    ('jane_smith', 'jane@example.com'),
    ('bob_jones', 'bob@test.com');

-- Select data with a condition
SELECT * FROM users
WHERE active = TRUE AND created_at > '2023-01-01';

-- Update a record
UPDATE users
SET active = FALSE
WHERE username = 'bob_jones';

-- Join with another table
SELECT u.username, p.title
FROM users u
JOIN posts p ON u.id = p.user_id
WHERE u.active = TRUE
ORDER BY p.created_at DESC;

-- Create an index
CREATE INDEX idx_users_email ON users(email);

-- Create a view
CREATE OR REPLACE VIEW active_users AS
SELECT id, username, email
FROM users
WHERE active = TRUE;

-- Delete a record
DELETE FROM users WHERE id = 100;

-- Drop the table
-- DROP TABLE users;
