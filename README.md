# NPS Dashboard

A web dashboard for exploring National Parks data using the National Park Service API.

## Features

- **Beautiful Front-end**: Clean, responsive design showcasing national parks
- **Google OAuth**: Secure authentication with Google accounts
- **Park Data**: Comprehensive park information from the NPS API
- **User Tracking**: Anonymous visitor tracking with email hashing for authenticated users
- **Caching**: Intelligent response caching with configurable TTL
- **Rate Limiting**: Built-in rate limiting to protect API endpoints

## Setup

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd nps-dashboard
   ```

2. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your actual values
   ```

3. **Get API Keys**
   - **NPS API Key**: Visit the [NPS Developer Portal](https://www.nps.gov/subjects/developer/get-started.htm) to request an API key
   - **Google OAuth**: Set up OAuth credentials in the [Google Cloud Console](https://console.cloud.google.com/)

4. **Install dependencies**

   ```bash
   go mod download
   ```

5. **Run the application**

   ```bash
   go run main.go
   ```

6. **Access the dashboard**
   Open your browser to `http://localhost:8080`

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `ENV` | Environment mode | `dev` or `prod` |
| `SERVER_PORT` | Port to run the server | `8080` |
| `DB_PATH` | Path to SQLite database | `./data/dashboard.db` |
| `NPS_API_KEY` | National Park Service API key | Required |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | Required for auth |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | Required for auth |
| `GOOGLE_REDIRECT_URI` | OAuth redirect URI | `http://localhost:8080/api/auth/callback` |

## API Endpoints

### Front-end

- `GET /` - Main dashboard page
- `GET /static/*` - Static assets

### Authentication

- `GET /api/auth/google` - Initiate Google OAuth
- `GET /api/auth/callback` - OAuth callback handler
- `GET /api/auth/logout` - Logout user
- `GET /api/user-info` - Get current user info

### Data

- `GET /park-list` - Get list of all parks
- `GET /park-cams` - Get live park cameras

## Architecture

The application is built with:

- **Backend**: Go with Chi router
- **Frontend**: Vanilla JavaScript with modern CSS
- **Database**: SQLite for user sessions and tracking
- **Caching**: In-memory caching with configurable policies
- **Authentication**: Google OAuth 2.0

## Development

### Adding New Features

1. **New API endpoints**: Add handlers in `internal/dashboard/routes.go`
2. **Database changes**: Update `internal/database/schema.sql`
3. **Front-end updates**: Modify `web/static/index.html`

### Project Structure

```text
nps-dashboard/
├── main.go              # Application entry point
├── internal/
│   ├── dashboard/       # Dashboard handlers and logic
│   │   ├── auth.go      # Google OAuth implementation
│   │   ├── dashboard.go # Core dashboard struct
│   │   ├── routes.go    # HTTP route handlers
│   │   └── tracking.go  # User tracking middleware
│   └── database/        # Database layer
│       ├── database.go  # Database connection and setup
│       └── schema.sql   # Database schema
├── web/
│   └── static/          # Static front-end files
│       └── index.html   # Main dashboard page
└── data/                # SQLite database storage
```

## About

Insights into our National Park Service, via the [National Park Service API](https://www.nps.gov/subjects/developer/index.htm)

- Gather general park data.
- Check upcoming event information.
- Investigate required passes and fees.
- View park alerts & live feeds.

## Resources

[NPS API](https://www.nps.gov/subjects/developer/guides.htm)   
[NPS Data Guidelines](https://www.nps.gov/aboutus/disclaimer.htm)   
[Visitor Use Statistics](https://www.nps.gov/subjects/socialscience/nps-visitor-use-statistics-definitions.htm)   
[Economic Valuation Information](https://www.nps.gov/subjects/socialscience/economic-valuation.htm)

## License

MIT License - see LICENSE file for details.