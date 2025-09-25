-- DevOps BE PostgreSQL Database Setup

-- Create database (run this command separately as postgres superuser)
-- CREATE DATABASE devops_db;

-- Connect to devops_db database, then run the following:

-- 1) USERS
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    profile_photo TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2) IMAGES
CREATE TABLE IF NOT EXISTS images (
    id SERIAL PRIMARY KEY,
    post_id INT,
    path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3) POSTS
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    publisher_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    publisher_name VARCHAR(100) NOT NULL,
    content_text TEXT NOT NULL,
    content_photo_id INT REFERENCES images(id) ON DELETE SET NULL,
    vote_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4) USER_VOTES
CREATE TABLE IF NOT EXISTS user_votes (
    post_id INT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    upvoted BOOLEAN NOT NULL,
    PRIMARY KEY (post_id, user_id)
);

-- Add foreign key constraint for images table
ALTER TABLE images ADD CONSTRAINT fk_images_post_id 
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE;

-- Insert sample users with DiceBear profile photos
INSERT INTO users (name, username, password, profile_photo) VALUES 
    ('John Doe', 'john_doe', '$2a$10$example_hashed_password_1', 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=1'),
    ('Jane Smith', 'jane_smith', '$2a$10$example_hashed_password_2', 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=2'),
    ('Mike Developer', 'developer_mike', '$2a$10$example_hashed_password_3', 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=3'),
    ('Alice Johnson', 'alice', '$2a$10$example_hashed_password_4', 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=4'),
    ('Bob Wilson', 'bob', '$2a$10$example_hashed_password_5', 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=5');

-- Insert sample posts
INSERT INTO posts (publisher_id, publisher_name, content_text) VALUES 
    (1, 'John Doe', 'Hello world! This is my first post.'),
    (2, 'Jane Smith', 'Love this new platform! 🚀'),
    (3, 'Mike Developer', 'Just deployed a new feature. Hope you like it!'),
    (1, 'Johnny D', 'Another post from John'),
    (2, 'Jane S', 'Jane posting again!');

-- Add some sample votes
INSERT INTO user_votes (post_id, user_id, upvoted) VALUES
    (1, 4, true),  -- alice upvotes john's post
    (1, 5, true),  -- bob upvotes john's post
    (2, 4, true),  -- alice upvotes jane's post
    (2, 1, false), -- john downvotes jane's post
    (3, 5, true),  -- bob upvotes mike's post
    (3, 4, true),  -- alice upvotes mike's post
    (4, 2, true),  -- jane upvotes john's second post
    (5, 1, false); -- john downvotes jane's second post

-- Update vote counts based on votes
UPDATE posts SET vote_count = (
    SELECT COALESCE(SUM(CASE WHEN upvoted THEN 1 ELSE -1 END), 0)
    FROM user_votes WHERE user_votes.post_id = posts.id
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_publisher_id ON posts(publisher_id);
CREATE INDEX IF NOT EXISTS idx_user_votes_post_id ON user_votes(post_id);
CREATE INDEX IF NOT EXISTS idx_user_votes_user_id ON user_votes(user_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_images_post_id ON images(post_id); 