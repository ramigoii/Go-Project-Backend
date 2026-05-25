CREATE TABLE IF NOT EXISTS reviews (
                                   id SERIAL PRIMARY KEY,
                                   user_id INTEGER NOT NULL,
                                   movie_id INTEGER NOT NULL,
                                   rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 10),
comment TEXT,
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
                         FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
UNIQUE (user_id, movie_id)
);