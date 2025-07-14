package dashboard

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ParkPageData represents the data structure for the park page template
type ParkPageData struct {
	Name        string
	ImageURL    string
	Description string
	States      string
	Designation string
	URL         string
}

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
		parks, err := dm.parkService.GetAllParks()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		parkList := ParkList{}
		for _, park := range parks {
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
			<a href="/api/auth/google" class="login-btn" hx-boost="false">
				<svg viewBox="-1 0 22 22" width="20" height="20" fill="none" xmlns="http://www.w3.org/2000/svg">
					<path fill-rule="evenodd" clip-rule="evenodd" d="M10 0C13.3137 0 16 2.68629 16 6C16 9.3137 13.3137 12 10 12C6.68629 12 4 9.3137 4 6C4 2.68629 6.68629 0 10 0zM1 22.099C0.44772 22.099 0 21.6513 0 21.099V19C0 16.2386 2.23858 14 5 14H15.0007C17.7621 14 20.0007 16.2386 20.0007 19V21.099C20.0007 21.6513 19.553 22.099 19.0007 22.099C18.4484 22.099 1.55228 22.099 1 22.099z" fill="currentColor"/>
				</svg>
			</a>
		`))
		return
	}

	// User authenticated
	html := fmt.Sprintf(`
		<div class="user-dropdown">
			<img src="/api/avatar" alt="%s" class="user-avatar" onclick="toggleDropdown(event)">
			<div class="dropdown-content">
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
	parks, err := dm.parkService.GetFeaturedParks()
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="loading">Error loading parks</div>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")

	if len(parks) == 0 {
		w.Write([]byte(`<div class="loading">No parks available</div>`))
		return
	}

	var html strings.Builder
	for _, park := range parks {
		// Get the first available image or use a placeholder
		imageUrl := ""
		imageAlt := park.Name
		if len(park.Images) > 0 {
			imageUrl = park.Images[0].URL
			if park.Images[0].AltText != "" {
				imageAlt = park.Images[0].AltText
			}
		}

		// Build the image HTML
		imageHTML := `<div class="park-image"></div>`
		if imageUrl != "" {
			imageHTML = fmt.Sprintf(`<div class="park-image" style="background-image: url('%s');" title="%s"></div>`, imageUrl, imageAlt)
		}

		html.WriteString(fmt.Sprintf(`
			<div class="park-card" data-park="%s">
				<a href="/parks/%s" class="park-card-link">
					%s
					<div class="park-content">
						<h3 class="park-title">%s</h3>
						<p class="park-description">
							Explore the natural beauty and unique features of %s.
						</p>
						<div class="park-location">United States</div>
					</div>
				</a>
			</div>
		`, park.Slug, park.Slug, imageHTML, park.Name, park.Name))
	}

	w.Write([]byte(html.String()))
}

// ParkSearchHandler returns HTML for park search results
func (dm *Dashboard) ParkSearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	w.Header().Set("Content-Type", "text/html")

	// If no query, show featured parks
	if strings.TrimSpace(query) == "" {
		dm.FeaturedParksHandler(w, r)
		return
	}

	// Search parks using the park service
	parks, err := dm.parkService.SearchParks(query)
	if err != nil {
		w.Write([]byte(`<div >Error searching parks</div>`))
		return
	}

	if len(parks) == 0 {
		w.Write([]byte(`<div class="no-results">
			<div class="no-results-icon">🏔️</div>
			<h3 class="no-results-title">No parks found</h3>
			<p class="no-results-message">We couldn't find any national parks matching "<strong>` + html.EscapeString(query) + `</strong>". Try adjusting your search terms.</p>
		</div>`))
		return
	}

	var html strings.Builder
	for _, park := range parks {
		// Get the first available image or use a placeholder
		imageUrl := ""
		imageAlt := park.Name
		if len(park.Images) > 0 {
			imageUrl = park.Images[0].URL
			if park.Images[0].AltText != "" {
				imageAlt = park.Images[0].AltText
			}
		}

		// Build the image HTML
		imageHTML := `<div class="park-image"></div>`
		if imageUrl != "" {
			imageHTML = fmt.Sprintf(`<div class="park-image" style="background-image: url('%s');" title="%s"></div>`, imageUrl, imageAlt)
		}

		html.WriteString(fmt.Sprintf(`
			<div class="park-card" data-park="%s">
				<a href="/parks/%s" class="park-card-link">
					%s
					<div class="park-content">
						<h3 class="park-title">%s</h3>
						<p class="park-description">
							Explore the natural beauty and unique features of %s.
						</p>
						<div class="park-location">United States</div>
					</div>
				</a>
			</div>
		`, park.Slug, park.Slug, imageHTML, park.Name, park.Name))
	}

	w.Write([]byte(html.String()))
}

// ParkDetailsHandler returns HTML for a specific park's details
func (dm *Dashboard) ParkDetailsHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[len("/api/parks/"):]
	slug = slug[:len(slug)-len("/details")]

	// Get park from cache
	park, err := dm.parkService.GetParkBySlug(slug)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="modal-content"><h2>Park not found</h2></div>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")

	// Get the first available image
	imageUrl := ""
	if len(park.Images) > 0 {
		imageUrl = park.Images[0].URL
	}

	// Build park details HTML
	imageHTML := ""
	if imageUrl != "" {
		imageHTML = fmt.Sprintf(`<img src="%s" alt="%s" style="width: 100%%; height: 300px; object-fit: cover; border-radius: 8px; margin-bottom: 1rem;">`, imageUrl, park.Name)
	}

	description := park.Description
	if description == "" {
		description = fmt.Sprintf("Discover the natural beauty and unique features of %s.", park.Name)
	}

	// Truncate description if it's too long
	if len(description) > 300 {
		description = description[:300] + "..."
	}

	html := fmt.Sprintf(`
		<div class="modal-content" style="padding: 2rem; max-width: 600px;">
			<div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1.5rem;">
				<h2 style="margin: 0; color: #111827; font-size: 1.8rem;">%s</h2>
				<button onclick="closeModal()" style="background: none; border: none; font-size: 1.5rem; cursor: pointer; color: #6b7280;">&times;</button>
			</div>
			%s
			<p style="color: #6b7280; line-height: 1.6; margin-bottom: 1.5rem;">%s</p>
			<div style="display: flex; gap: 1rem; flex-wrap: wrap;">
				<div style="background: #f3f4f6; padding: 0.75rem 1rem; border-radius: 6px; font-size: 0.9rem;">
					<strong>State(s):</strong> %s
				</div>
				<div style="background: #f3f4f6; padding: 0.75rem 1rem; border-radius: 6px; font-size: 0.9rem;">
					<strong>Designation:</strong> %s
				</div>
			</div>
			<div style="margin-top: 1.5rem; padding-top: 1.5rem; border-top: 1px solid #e5e7eb;">
				<a href="%s" target="_blank" style="background: #16a34a; color: white; padding: 0.75rem 1.5rem; border-radius: 6px; text-decoration: none; font-weight: 600; display: inline-block;">
					Visit Official Site
				</a>
			</div>
		</div>
	`, park.Name, imageHTML, description, park.States, park.Designation, park.URL)

	w.Write([]byte(html))
}

// ParkPageHandler returns a full HTML page for a specific park
func (dm *Dashboard) ParkPageHandler(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Get park from cache
	park, err := dm.parkService.GetParkBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	// Get the first available image
	imageUrl := ""
	if len(park.Images) > 0 {
		imageUrl = park.Images[0].URL
	}

	description := park.Description
	if description == "" {
		description = fmt.Sprintf("Explore a variety of activities available at %s, from hiking and rock climbing to guided tours and stargazing.", park.Name)
	}

	// Create template data
	data := ParkPageData{
		Name:        park.Name,
		ImageURL:    imageUrl,
		Description: description,
		States:      park.States,
		Designation: park.Designation,
		URL:         park.URL,
	}

	// Parse and execute the template from file
	tmpl, err := template.ParseFiles("web/templates/park.html")
	if err != nil {
		http.Error(w, "Failed to load park page template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to render park page", http.StatusInternalServerError)
		return
	}
}

// TemplateHandler serves HTML template fragments
func (dm *Dashboard) TemplateHandler(w http.ResponseWriter, r *http.Request) {
	templateName := chi.URLParam(r, "template")

	w.Header().Set("Content-Type", "text/html")

	switch templateName {
	case "header":
		tmpl, err := template.ParseFiles("web/templates/header.html")
		if err != nil {
			http.Error(w, "Failed to load header template", http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Failed to render header template", http.StatusInternalServerError)
			return
		}
	case "footer":
		tmpl, err := template.ParseFiles("web/templates/footer.html")
		if err != nil {
			http.Error(w, "Failed to load footer template", http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Failed to render footer template", http.StatusInternalServerError)
			return
		}
	default:
		http.NotFound(w, r)
	}
}

// ParksHandler returns HTML for parks with pagination support
func (dm *Dashboard) ParksHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	limit := 12 // Default page size

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	parks, err := dm.parkService.GetParksWithPagination(offset, limit)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="loading">Error loading parks</div>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")

	if len(parks) == 0 {
		w.Write([]byte(`<div class="loading">No more parks available</div>`))
		return
	}

	var html strings.Builder
	for _, park := range parks {
		// Get the first available image or use a placeholder
		imageUrl := ""
		imageAlt := park.Name
		if len(park.Images) > 0 {
			imageUrl = park.Images[0].URL
			if park.Images[0].AltText != "" {
				imageAlt = park.Images[0].AltText
			}
		}

		// Build the image HTML
		imageHTML := `<div class="park-image"></div>`
		if imageUrl != "" {
			imageHTML = fmt.Sprintf(`<div class="park-image" style="background-image: url('%s');" title="%s"></div>`, imageUrl, imageAlt)
		}

		html.WriteString(fmt.Sprintf(`
			<div class="park-card" data-park="%s">
				<a href="/parks/%s" class="park-card-link">
					%s
					<div class="park-content">
						<h3 class="park-title">%s</h3>
						<p class="park-description">
							Explore the natural beauty and unique features of %s.
						</p>
						<div class="park-location">United States</div>
					</div>
				</a>
			</div>
		`, park.Slug, park.Slug, imageHTML, park.Name, park.Name))
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
