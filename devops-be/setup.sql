-- DevOps BE Database Setup
-- This script creates all necessary tables for the application

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    profile_photo TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Images table
CREATE TABLE IF NOT EXISTS images (
    id SERIAL PRIMARY KEY,
    post_id INTEGER,
    path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Posts table
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    publisher_id INTEGER NOT NULL REFERENCES users(id),
    publisher_name VARCHAR(255) DEFAULT 'Anonymous',
    header TEXT NOT NULL,
    content_text TEXT NOT NULL,
    content_photo_id INTEGER REFERENCES images(id),
    is_private BOOLEAN DEFAULT FALSE,
    vote_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User votes table (for tracking upvotes/downvotes)
CREATE TABLE IF NOT EXISTS user_votes (
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    upvoted BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, user_id)
);

-- Add foreign key constraint to images table
ALTER TABLE images
ADD CONSTRAINT fk_images_post
FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE;

-- Sample data
-- NOTE: All sample users have the same password: "password123"
-- Passwords are stored as bcrypt hashes (never store plain text passwords!)
INSERT INTO users (name, username, password, profile_photo) VALUES
('John Doe', 'john_doe', '$2a$10$YyMu9W4FgeVKLsKzRofXIesu4vSVF1Rk8yjfzrF.q4fZNxMsVZSoG', 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=1'),
('Jane Smith', 'jane_smith', '$2a$10$l1IHmn2gJobbgZtYQ6Qmi.DtsReGVD.udsk4Im3zc1/EIDP4rMI1O', 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=2'),
('Bob Wilson', 'bob_wilson', '$2a$10$azP54b5aw6iLZZhm4kHQ/uLuSIyvNNiegCYa5atFjOojEqhFfLI/W', 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=3')
ON CONFLICT (username) DO NOTHING;

INSERT INTO posts (publisher_id, publisher_name, header, content_text, is_private, vote_count) VALUES
(1, 'John Doe', 'First Post', 'Hello world! This is my first post.', false, 5),
(2, 'Jane Smith', 'Welcome', 'Great to be here!', false, 3),
(1, 'John Doe', 'Update', 'Just had a great day learning DevOps!', false, 8),
(1, 'John Doe', 'Private Note', 'This is my private thought.', true, 0)
ON CONFLICT DO NOTHING;

INSERT INTO user_votes (post_id, user_id, upvoted) VALUES
(1, 2, true),
(1, 3, true),
(2, 1, true),
(3, 2, true),
(3, 3, false)
ON CONFLICT DO NOTHING;
