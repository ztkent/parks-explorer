-- Users Table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    google_id TEXT UNIQUE,
    avatar_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT 1
);

-- Sessions Table for managing auth tokens
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    session_token TEXT UNIQUE NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Parks Table for caching park data from NPS API
CREATE TABLE parks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    park_code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    full_name TEXT,
    slug TEXT NOT NULL,
    states TEXT,
    designation TEXT,
    description TEXT,
    weather_info TEXT,
    directions_info TEXT,
    url TEXT,
    directions_url TEXT,
    latitude TEXT,
    longitude TEXT,
    lat_long TEXT,
    relevance_score REAL,
    api_data TEXT, -- JSON blob of complete API response
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Park Images Table for caching park images
CREATE TABLE park_images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    park_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    title TEXT,
    alt_text TEXT,
    caption TEXT,
    credit TEXT,
    image_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (park_id) REFERENCES parks(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX idx_parks_park_code ON parks(park_code);
CREATE INDEX idx_parks_slug ON parks(slug);
CREATE INDEX idx_parks_last_fetched ON parks(last_fetched_at);
CREATE INDEX idx_park_images_park_id ON park_images(park_id);
CREATE INDEX idx_park_images_order ON park_images(park_id, image_order);
