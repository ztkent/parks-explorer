<div align="center">
  <h3>National Parks Explorer - Dashboard for Park Data and News</h3>
  <p>
    <a href="https://parksexplorer.us">🌲 Parks Explorer</a> • 
    <a href="#quick-start">Quick Start</a> • 
    <a href="#api-endpoints">API</a> 
  </p>
</div>

## Overview

Insights into our National Park Service, via the [National Park Service API](https://www.nps.gov/subjects/developer/index.htm).  
Provides detailed park information, activities, events, camping options, and real-time updates.

## Features

- **Complete Park Database**: Access to all 400+ National Park Service sites including parks, monuments, battlefields, and historic sites
- **Interactive Exploration**: Browse parks by state, activity type, or search functionality
- **Comprehensive Park Data**: Including activities, camping, events, news, visitor centers, and amenities
- **Real-Time Updates**: Live park alerts, weather conditions, and webcam feeds
- **Google OAuth Authentication**: Secure user authentication with personalized tracking
- **Image Proxy Service**: Secure image serving with caching

## Architecture

```
parks-explorer/
├── main.go                   
├── internal/
│   ├── dashboard/           # Core dashboard logic and HTTP handlers
│   │   ├── auth.go          # Google OAuth 2.0 authentication
│   └── database/            # Database layer and schema management
├── web/
│   ├── static/              # Frontend assets and resources
│   │   └── assets/          # Images, favicons, and media files
│   └── templates/           # HTML template partials
└── data/                    # SQLite database storage
```

## Technology Stack

### Backend
- **Language**: Go 1.24
- **Framework**: Chi router with middleware
- **Database**: SQLite with indexing and caching
- **Authentication**: Google OAuth 2.0 with secure session management
- **API Integration**: National Park Service API via [go-nps](https://github.com/ztkent/go-nps) package
- **Caching**: In-memory caching with [replay](https://github.com/ztkent/replay) library

### Frontend
- **Framework**: HTML/CSS/JavaScript with HTMX
- **Components**: Modular HTML templates with server-side rendering

## Quick Start

### Prerequisites
- Go 1.24 or higher
- Docker and Docker Compose
- National Park Service API key
- Google OAuth credentials

### Environment Setup

1. **Clone the repository**:
```bash
git clone https://github.com/ztkent/parks-explorer.git
cd parks-explorer
```

2. **Configure environment variables**:
```bash
# Required for National Park Service API access
export NPS_API_KEY="your_nps_api_key"

# Required for Google OAuth authentication
export GOOGLE_CLIENT_ID="your_google_client_id"
export GOOGLE_CLIENT_SECRET="your_google_client_secret"
export GOOGLE_REDIRECT_URI="https://your-domain.com/api/auth/google/callback"

# Optional configuration
export SERVER_PORT="8086"
export DB_PATH="./data/dashboard.db"
export ENV="dev"
```

3. **Build and run with Docker**:
```bash
make app-up
```

4. **Run locally for development**:
```bash
go mod download
go run main.go
```

5. **Access the application**:
- Development: `http://localhost:8086`
- Production: `https://parksexplorer.us`

### Database Schema

The application uses SQLite with the following core tables:
- `parks`: Complete park information and metadata
- `users`: User accounts and authentication data
- `sessions`: Secure session management
- `park_activities`: Cached activities and things-to-do data
- `park_media`: Images, videos, and webcam feeds
- `park_news`: News articles, alerts, and events
- `park_details`: Visitor centers, campgrounds, and amenities

## API Endpoints

### Public Endpoints
- `GET /` - Main dashboard interface
- `GET /parks/{slug}` - Individual park pages
- `GET /things-to-do` - Activities and attractions
- `GET /events` - Park events and programs
- `GET /camping` - Campground information
- `GET /news` - Park news and updates

### Authentication Endpoints
- `GET /api/auth/google` - Initiate Google OAuth flow
- `GET /api/auth/google/callback` - OAuth callback handler
- `GET /api/auth/logout` - User logout
- `GET /api/user-info` - Current user information
- `GET /api/auth-status` - Authentication status check

### Data API Endpoints
- `GET /api/parks` - Paginated parks listing
- `GET /api/parks/featured` - Featured parks carousel
- `GET /api/parks/search` - Park search functionality
- `GET /api/things-to-do/search` - Activities search
- `GET /api/events/search` - Events search and filtering
- `GET /api/camping/search` - Campground search
- `GET /api/news/search` - News articles search

### Park-Specific Endpoints
- `GET /api/parks/{parkCode}/overview` - Park overview data
- `GET /api/parks/{parkCode}/activities` - Park activities and tours
- `GET /api/parks/{parkCode}/media` - Park galleries and webcams
- `GET /api/parks/{parkCode}/news` - Park-specific news and alerts
- `GET /api/parks/{parkCode}/details` - Visitor centers and amenities

### Utility Endpoints
- `GET /api/image-proxy` - Secure image serving with caching
- `GET /api/avatar` - User avatar proxy service
- `GET /api/analytics/config` - Analytics configuration

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NPS_API_KEY` | National Park Service API key | - | Yes |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | - | Yes |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | - | Yes |
| `GOOGLE_REDIRECT_URI` | OAuth redirect URI | - | Yes |
| `SERVER_PORT` | Server port | `8086` | No |
| `DB_PATH` | SQLite database path | `./data/dashboard.db` | No |
| `ENV` | Environment mode (`dev`/`prod`) | `dev` | No |

## Development

### Getting API Keys
- **NPS API Key**: Register at the [NPS Developer Portal](https://www.nps.gov/subjects/developer/get-started.htm)
- **Google OAuth**: Configure credentials in [Google Cloud Console](https://console.cloud.google.com/)

### Local Development
```bash
# Install dependencies
go mod download

# Run with hot reload (development mode)
ENV=dev go run main.go

# Run tests
go test ./...

# Build for production
go build -o parks main.go
```

### Docker Development
```bash
# Start all services
make app-up

# Fresh build and restart
make app-fresh

# Stop services
make app-down
```

## Resources

- [National Park Service API Documentation](https://www.nps.gov/subjects/developer/guides.htm)
- [NPS Data Guidelines and Terms](https://www.nps.gov/aboutus/disclaimer.htm)
- [Go-NPS Package Documentation](https://github.com/ztkent/go-nps)
- [Replay Caching Library](https://github.com/ztkent/replay)
