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

-- Park Activities Cache
CREATE TABLE park_activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    park_id INTEGER NOT NULL,
    data_type TEXT NOT NULL, -- 'activities', 'things_to_do', 'tours'
    api_data TEXT NOT NULL, -- JSON blob of API response
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (park_id) REFERENCES parks(id) ON DELETE CASCADE,
    UNIQUE(park_id, data_type)
);

-- Park Media Cache
CREATE TABLE park_media (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    park_id INTEGER NOT NULL,
    data_type TEXT NOT NULL, -- 'galleries', 'videos', 'audio', 'webcams'
    api_data TEXT NOT NULL, -- JSON blob of API response
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (park_id) REFERENCES parks(id) ON DELETE CASCADE,
    UNIQUE(park_id, data_type)
);

-- Park News Cache
CREATE TABLE park_news (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    park_id INTEGER NOT NULL,
    data_type TEXT NOT NULL, -- 'articles', 'alerts', 'events', 'news_releases'
    api_data TEXT NOT NULL, -- JSON blob of API response
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (park_id) REFERENCES parks(id) ON DELETE CASCADE,
    UNIQUE(park_id, data_type)
);

-- Park Details Cache
CREATE TABLE park_details (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    park_id INTEGER NOT NULL,
    data_type TEXT NOT NULL, -- 'visitor_centers', 'campgrounds', 'fees', 'amenities', 'parking_lots'
    api_data TEXT NOT NULL, -- JSON blob of API response
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (park_id) REFERENCES parks(id) ON DELETE CASCADE,
    UNIQUE(park_id, data_type)
);

-- Park Gallery Assets Cache
CREATE TABLE park_gallery_assets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    park_id INTEGER NOT NULL,
    gallery_id TEXT NOT NULL,
    asset_id TEXT,
    title TEXT,
    alt_text TEXT,
    caption TEXT,
    credit TEXT,
    url TEXT,
    asset_type TEXT, -- 'image', 'video', etc.
    file_size INTEGER,
    width INTEGER,
    height INTEGER,
    api_data TEXT NOT NULL, -- JSON blob of complete asset data
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (park_id) REFERENCES parks(id) ON DELETE CASCADE,
    UNIQUE(park_id, gallery_id, asset_id)
);

CREATE INDEX idx_park_gallery_assets_park_id ON park_gallery_assets(park_id);
CREATE INDEX idx_park_gallery_assets_gallery_id ON park_gallery_assets(park_id, gallery_id);
CREATE INDEX idx_park_gallery_assets_last_fetched ON park_gallery_assets(last_fetched_at);
CREATE INDEX idx_park_gallery_assets_type ON park_gallery_assets(asset_type);
CREATE INDEX idx_parks_park_code ON parks(park_code);
CREATE INDEX idx_parks_slug ON parks(slug);
CREATE INDEX idx_parks_last_fetched ON parks(last_fetched_at);
CREATE INDEX idx_park_images_park_id ON park_images(park_id);
CREATE INDEX idx_park_images_order ON park_images(park_id, image_order);
CREATE INDEX idx_park_activities_park_id ON park_activities(park_id);
CREATE INDEX idx_park_activities_type ON park_activities(park_id, data_type);
CREATE INDEX idx_park_activities_last_fetched ON park_activities(last_fetched_at);
CREATE INDEX idx_park_media_park_id ON park_media(park_id);
CREATE INDEX idx_park_media_type ON park_media(park_id, data_type);
CREATE INDEX idx_park_media_last_fetched ON park_media(last_fetched_at);
CREATE INDEX idx_park_news_park_id ON park_news(park_id);
CREATE INDEX idx_park_news_type ON park_news(park_id, data_type);
CREATE INDEX idx_park_news_last_fetched ON park_news(last_fetched_at);
CREATE INDEX idx_park_details_park_id ON park_details(park_id);
CREATE INDEX idx_park_details_type ON park_details(park_id, data_type);
CREATE INDEX idx_park_details_last_fetched ON park_details(last_fetched_at);
