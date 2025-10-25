-- Script to reset sample user passwords with proper bcrypt hashes
-- Run this if you already have the database set up

-- Delete existing sample users (this will also delete their posts due to CASCADE)
DELETE FROM users WHERE username IN ('john_doe', 'jane_smith', 'bob_wilson');

-- Re-insert with proper bcrypt hashes
-- All users have password: "password123"
INSERT INTO users (name, username, password, profile_photo) VALUES
('John Doe', 'john_doe', '$2a$10$YyMu9W4FgeVKLsKzRofXIesu4vSVF1Rk8yjfzrF.q4fZNxMsVZSoG', 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=1'),
('Jane Smith', 'jane_smith', '$2a$10$l1IHmn2gJobbgZtYQ6Qmi.DtsReGVD.udsk4Im3zc1/EIDP4rMI1O', 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=2'),
('Bob Wilson', 'bob_wilson', '$2a$10$azP54b5aw6iLZZhm4kHQ/uLuSIyvNNiegCYa5atFjOojEqhFfLI/W', 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=3');

-- Re-insert sample posts
INSERT INTO posts (publisher_id, publisher_name, header, content_text, is_private, vote_count) VALUES
(1, 'John Doe', 'First Post', 'Hello world! This is my first post.', false, 5),
(2, 'Jane Smith', 'Welcome', 'Great to be here!', false, 3),
(1, 'John Doe', 'Update', 'Just had a great day learning DevOps!', false, 8),
(1, 'John Doe', 'Private Note', 'This is my private thought.', true, 0);

-- Re-insert sample votes
INSERT INTO user_votes (post_id, user_id, upvoted) VALUES
(1, 2, true),
(1, 3, true),
(2, 1, true),
(3, 2, true),
(3, 3, false);
