package dashboard

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// HomeHandler serves the main HTML page
func (dm *Dashboard) HomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Serve the static HTML file
		http.ServeFile(w, r, "web/static/index.html")
	}
}

// StaticFileHandler serves static files
func (dm *Dashboard) StaticFileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the file path from URL
		filePath := r.URL.Path[len("/static/"):]
		fullPath := filepath.Join("web/static", filePath)

		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, fullPath)
	}
}

// LiveParkCamsHandler returns a list of live park cameras.
type ParkCam struct {
	Title string `json:"title"`
	Image string `json:"image"`
	Link  string `json:"link"`
}

func (dm *Dashboard) LiveParkCamsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parkCams := []ParkCam{
			{
				Title: "Park Cam 1",
				Image: "https://via.placeholder.com/100",
				Link:  "#",
			},
			{
				Title: "Park Cam 2",
				Image: "https://via.placeholder.com/100",
				Link:  "#",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(parkCams)
	}
}

// ParkListHandler returns a list of parks.
type ParkList struct {
	Parks []string `json:"parks"`
}

func (dm *Dashboard) ParkListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := dm.npsApi.GetParks(nil, nil, 0, 500, "", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		parkList := ParkList{}
		for _, park := range res.Data {
			parkList.Parks = append(parkList.Parks, park.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(parkList)
	}
}

// Unique identifier for the user, stored in a cookie.
func (dm *Dashboard) EnsureUUIDHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("uuid")
		if err == http.ErrNoCookie {
			// Cookie does not exist, set it
			token := uuid.New().String()
			http.SetCookie(w, &http.Cookie{
				Name:     "uuid",
				Value:    token,
				HttpOnly: true,
				Secure:   true, // Set to true if your site uses HTTPS
				SameSite: http.SameSiteStrictMode,
			})
		} else if err != nil {
			// Some other error occurred
			http.Error(w, "Failed to read cookie", http.StatusInternalServerError)
		}
	}
}

// AvatarProxyHandler avoids CORS issues by fetching the avatar on the server side
func (dm *Dashboard) AvatarProxyHandler(w http.ResponseWriter, r *http.Request) {
	user, err := dm.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if user.AvatarURL == "" {
		http.Error(w, "No avatar URL", http.StatusNotFound)
		return
	}

	// Fetch the image from Google
	resp, err := http.Get(user.AvatarURL)
	if err != nil {
		http.Error(w, "Failed to fetch avatar", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy headers
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	// Copy the image data
	io.Copy(w, resp.Body)
}

// AuthStatusHandler returns HTML for the authentication status
func (dm *Dashboard) AuthStatusHandler(w http.ResponseWriter, r *http.Request) {
	user, err := dm.GetCurrentUser(r)
	w.Header().Set("Content-Type", "text/html")

	if err != nil {
		// User not authenticated
		w.Write([]byte(`
			<a href="/api/auth/google" class="google-signin-icon tooltip" data-tooltip="Sign in with Google" hx-boost="false">
				<svg viewBox="0 0 24 24" fill="currentColor">
					<path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
					<path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
					<path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
					<path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
				</svg>
			</a>
		`))
		return
	}

	// User authenticated
	html := fmt.Sprintf(`
		<div class="user-dropdown">
			<img src="/api/avatar" alt="%s" class="user-avatar" id="userAvatar">
			<div class="dropdown-content" id="userDropdown">
				<div class="dropdown-header">
					<div class="dropdown-user-name">%s</div>
					<div class="dropdown-user-email">%s</div>
				</div>
				<a href="/api/auth/logout" class="dropdown-item logout" hx-boost="false">Sign out</a>
			</div>
		</div>
	`, user.Username, user.Username, user.Email)

	w.Write([]byte(html))
}

// FeaturedParksHandler returns HTML for featured parks
func (dm *Dashboard) FeaturedParksHandler(w http.ResponseWriter, r *http.Request) {
	res, err := dm.npsApi.GetParks(nil, nil, 0, 500, "", nil)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="loading">Error loading parks</div>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")

	if len(res.Data) == 0 {
		w.Write([]byte(`<div class="loading">No parks available</div>`))
		return
	}

	// Show first 6 parks as featured
	featuredParks := res.Data
	if len(featuredParks) > 6 {
		featuredParks = featuredParks[:6]
	}

	var html strings.Builder
	for _, park := range featuredParks {
		parkSlug := createSlug(park.Name)
		html.WriteString(fmt.Sprintf(`
			<div class="park-card" data-park="%s" 
				 hx-get="/api/parks/%s/details" 
				 hx-target="#park-modal" 
				 hx-swap="innerHTML">
				<div class="park-image"></div>
				<div class="park-content">
					<h3 class="park-title">%s</h3>
					<p class="park-description">
						Explore the natural beauty and unique features of %s.
					</p>
					<div class="park-location">United States</div>
				</div>
			</div>
		`, parkSlug, parkSlug, park.Name, park.Name))
	}

	w.Write([]byte(html.String()))
}

// ParkSearchHandler returns HTML for park search results
func (dm *Dashboard) ParkSearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	res, err := dm.npsApi.GetParks(nil, nil, 0, 500, "", nil)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="loading">Error loading parks</div>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")

	// If no query, show featured parks
	if strings.TrimSpace(query) == "" {
		dm.FeaturedParksHandler(w, r)
		return
	}

	// Filter parks by query
	queryLower := strings.ToLower(query)
	var html strings.Builder
	matchCount := 0

	for _, park := range res.Data {
		if strings.Contains(strings.ToLower(park.Name), queryLower) {
			parkSlug := createSlug(park.Name)
			html.WriteString(fmt.Sprintf(`
				<div class="park-card" data-park="%s">
					<div class="park-image"></div>
					<div class="park-content">
						<h3 class="park-title">%s</h3>
						<p class="park-description">
							Explore the natural beauty and unique features of %s.
						</p>
						<div class="park-location">United States</div>
					</div>
				</div>
			`, parkSlug, park.Name, park.Name))
			matchCount++
		}
	}

	if matchCount == 0 {
		w.Write([]byte(`<div class="loading">No parks found matching your search</div>`))
		return
	}

	w.Write([]byte(html.String()))
}

// Helper function to create slugs
func createSlug(text string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, "&", "and")
	return slug
}
