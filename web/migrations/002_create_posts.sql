CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    content TEXT,
    post_type VARCHAR(50) NOT NULL,
    file_url VARCHAR(550),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT posts_post_type_check CHECK (post_type IN ('blog', 'poem', 'song')),
    CONSTRAINT posts_title_not_blank CHECK (btrim(title) <> '')
);

CREATE INDEX IF NOT EXISTS posts_user_id_idx ON posts (user_id);
CREATE INDEX IF NOT EXISTS posts_created_at_idx ON posts (created_at DESC);
CREATE INDEX IF NOT EXISTS posts_post_type_created_at_idx ON posts (post_type, created_at DESC);
