package dashboard

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
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
	Code        string
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
			log.Printf("Failed to get all parks: %v", err)
			http.Error(w, fmt.Sprintf("Failed to get parks: %v", err), http.StatusInternalServerError)
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
			log.Printf("Failed to read cookie: %v", err)
			http.Error(w, fmt.Sprintf("Failed to read cookie: %v", err), http.StatusInternalServerError)
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
		log.Printf("Failed to fetch avatar from %s: %v", user.AvatarURL, err)
		http.Error(w, fmt.Sprintf("Failed to fetch avatar: %v", err), http.StatusInternalServerError)
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
						<div class="park-location">%s</div>
					</div>
				</a>
			</div>
		`, park.Slug, park.Slug, imageHTML, park.Name, park.Name, formatStatesDisplay(park.States)))
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
						<div class="park-location">%s</div>
					</div>
				</a>
			</div>
		`, park.Slug, park.Slug, imageHTML, park.Name, park.Name, formatStatesDisplay(park.States)))
	}

	w.Write([]byte(html.String()))
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
		Code:        park.ParkCode,
		ImageURL:    imageUrl,
		Description: description,
		States:      park.States,
		Designation: park.Designation,
		URL:         park.URL,
	}

	// Parse and execute the template from file
	tmpl, err := template.ParseFiles("web/templates/park.html")
	if err != nil {
		log.Printf("Failed to load park page template: %v", err)
		http.Error(w, fmt.Sprintf("Failed to load park page template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to render park page for %s: %v", park.Name, err)
		http.Error(w, fmt.Sprintf("Failed to render park page: %v", err), http.StatusInternalServerError)
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
			log.Printf("Failed to load header template: %v", err)
			http.Error(w, fmt.Sprintf("Failed to load header template: %v", err), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, nil)
		if err != nil {
			log.Printf("Failed to render header template: %v", err)
			http.Error(w, fmt.Sprintf("Failed to render header template: %v", err), http.StatusInternalServerError)
			return
		}
	case "footer":
		tmpl, err := template.ParseFiles("web/templates/footer.html")
		if err != nil {
			log.Printf("Failed to load footer template: %v", err)
			http.Error(w, fmt.Sprintf("Failed to load footer template: %v", err), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, nil)
		if err != nil {
			log.Printf("Failed to render footer template: %v", err)
			http.Error(w, fmt.Sprintf("Failed to render footer template: %v", err), http.StatusInternalServerError)
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
						<div class="park-location">%s</div>
					</div>
				</a>
			</div>
		`, park.Slug, park.Slug, imageHTML, park.Name, park.Name, formatStatesDisplay(park.States)))
	}

	w.Write([]byte(html.String()))
}

// formatStatesDisplay formats the states string to show only first 5 states when there are more than 5
func formatStatesDisplay(states string) string {
	if states == "" {
		return ""
	}

	// Split states by comma and trim whitespace
	stateList := strings.Split(states, ",")
	for i := range stateList {
		stateList[i] = strings.TrimSpace(stateList[i])
	}

	// If 5 or fewer states, return as is
	if len(stateList) <= 5 {
		return strings.Join(stateList, ", ")
	}

	// If more than 5 states, show first 5 plus count of remaining
	firstFive := stateList[:5]
	remaining := len(stateList) - 5
	return strings.Join(firstFive, ", ") + fmt.Sprintf(", +%d", remaining)
}

// ParkOverviewHandler returns comprehensive overview data for a park
func (dm *Dashboard) ParkOverviewHandler(w http.ResponseWriter, r *http.Request) {
	parkCode := chi.URLParam(r, "parkCode")
	if parkCode == "" {
		http.Error(w, "Park code required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	thingsToDo, _ := dm.parkService.GetParkThingsToDo(parkCode)
	activities, _ := dm.parkService.GetParkActivities(parkCode)
	visitorCenters, _ := dm.parkService.GetParkVisitorCenters(parkCode)
	amenities, _ := dm.parkService.GetParkAmenities(parkCode)
	tours, _ := dm.parkService.GetParkTours(parkCode)
	events, _ := dm.parkService.GetParkEvents(parkCode)
	alerts, _ := dm.parkService.GetParkAlerts(parkCode)
	// Prepare data for the template with proper structure
	data := map[string]interface{}{
		"ThingsToDo":     thingsToDo,
		"Activities":     activities,
		"VisitorCenters": visitorCenters,
		"Amenities":      amenities,
		"ParkTours":      tours,
		"ParkEvents":     events,
		"ParkAlerts":     alerts,
		"ParkCode":       parkCode,
	}

	// Parse and execute the overview template
	tmpl, err := template.ParseFiles("web/templates/partials/park-overview.html")
	if err != nil {
		log.Printf("Failed to load overview template: %v", err)
		http.Error(w, fmt.Sprintf("Failed to load overview template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to render overview content for park %s: %v", parkCode, err)
		http.Error(w, fmt.Sprintf("Failed to render overview content: %v", err), http.StatusInternalServerError)
		return
	}
}

// ParkActivitiesHandler returns comprehensive activities and things to do data
func (dm *Dashboard) ParkActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	parkCode := chi.URLParam(r, "parkCode")
	if parkCode == "" {
		http.Error(w, "Park code required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	// Fetch comprehensive activity data with all available details
	thingsToDo, _ := dm.parkService.GetParkThingsToDo(parkCode)
	tours, _ := dm.parkService.GetParkTours(parkCode)
	events, _ := dm.parkService.GetParkEvents(parkCode)
	campgrounds, _ := dm.parkService.GetParkCampgrounds(parkCode)
	activities, _ := dm.parkService.GetParkActivities(parkCode)

	// Comprehensive data structure leveraging all available NPS API fields
	data := map[string]interface{}{
		"ThingsToDo":  thingsToDo,
		"Tours":       tours,
		"Events":      events,
		"Campgrounds": campgrounds,
		"Activities":  activities,
		"ParkCode":    parkCode,
	}

	// Parse and execute the activities template
	tmpl, err := template.New("activities").Funcs(template.FuncMap{
		"unescapeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).ParseFiles("web/templates/partials/park-activities.html")
	if err != nil {
		log.Printf("Failed to load activities template: %v", err)
		http.Error(w, fmt.Sprintf("Failed to load activities template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to render activities content for park %s: %v", parkCode, err)
		http.Error(w, fmt.Sprintf("Failed to render activities content: %v", err), http.StatusInternalServerError)
		return
	}
}

// ParkMediaHandler returns comprehensive multimedia content including all media types
func (dm *Dashboard) ParkMediaHandler(w http.ResponseWriter, r *http.Request) {
	parkCode := chi.URLParam(r, "parkCode")
	if parkCode == "" {
		http.Error(w, "Park code required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	// 	mediaTypes := []string{"galleries", "videos", "audio", "webcams"}
	galleries, _ := dm.parkService.GetParkMultimediaGalleries(parkCode)
	videos, _ := dm.parkService.GetParkMultimediaVideos(parkCode)
	audio, _ := dm.parkService.GetParkMultimediaAudio(parkCode)
	webcams, _ := dm.parkService.GetParkWebcams(parkCode)

	// Comprehensive media data structure
	data := map[string]interface{}{
		"Galleries": galleries, // Comprehensive gallery data with detailed metadata
		"Videos":    videos,    // Comprehensive video data with detailed metadata
		"Audio":     audio,     // Additional audio content with detailed metadata
		"Webcams":   webcams,   // Live webcam feeds with detailed metadata
		"ParkCode":  parkCode,
	}

	// Parse and execute the media template
	tmpl, err := template.New("park-media.html").Funcs(template.FuncMap{
		"formatDuration": func(durationMs interface{}) string {
			var ms int64
			switch v := durationMs.(type) {
			case int:
				ms = int64(v)
			case int64:
				ms = v
			case float64:
				ms = int64(v)
			default:
				return "Unknown"
			}

			if ms <= 0 {
				return "Unknown"
			}

			seconds := ms / 1000
			minutes := seconds / 60
			hours := minutes / 60

			if hours > 0 {
				remainingMinutes := minutes % 60
				return fmt.Sprintf("%d:%02d", hours, remainingMinutes)
			} else if minutes > 0 {
				return fmt.Sprintf("%d min", minutes)
			} else {
				return fmt.Sprintf("%d sec", seconds)
			}
		},
	}).ParseFiles("web/templates/partials/park-media.html")
	if err != nil {
		log.Printf("Failed to load media template: %v", err)
		http.Error(w, fmt.Sprintf("Failed to load media template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to render media content for park %s: %v", parkCode, err)
		http.Error(w, "Failed to render media content: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ParkNewsHandler returns comprehensive news, alerts, and updates with full content details
func (dm *Dashboard) ParkNewsHandler(w http.ResponseWriter, r *http.Request) {
	parkCode := chi.URLParam(r, "parkCode")
	if parkCode == "" {
		http.Error(w, "Park code required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	newsReleases, _ := dm.parkService.GetParkNewsReleases(parkCode)
	articles, _ := dm.parkService.GetParkArticles(parkCode)
	alerts, _ := dm.parkService.GetParkAlerts(parkCode)
	events, _ := dm.parkService.GetParkEvents(parkCode)

	data := map[string]interface{}{
		"NewsReleases": newsReleases, // Official press releases with detailed content
		"Articles":     articles,     // In-depth articles with full text, images, and metadata
		"Alerts":       alerts,       // Critical alerts with full descriptions, categories, and
		"Events":       events,       // Upcoming events with full details, including dates,
		"ParkCode":     parkCode,
	}

	// Parse and execute the news template
	tmpl, err := template.New("park-news.html").Funcs(template.FuncMap{
		"unescapeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).ParseFiles("web/templates/partials/park-news.html")
	if err != nil {
		log.Printf("Failed to load news template: %v", err)
		http.Error(w, fmt.Sprintf("Failed to load news template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to render news content for park %s: %v", parkCode, err)
		http.Error(w, fmt.Sprintf("Failed to render news content: %v", err), http.StatusInternalServerError)
		return
	}
}

// ParkEnhancedDetailsHandler returns comprehensive park details and facilities information
func (dm *Dashboard) ParkDetailsHandler(w http.ResponseWriter, r *http.Request) {
	parkCode := chi.URLParam(r, "parkCode")
	if parkCode == "" {
		http.Error(w, "Park code required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	visitors_centers, _ := dm.parkService.GetParkVisitorCenters(parkCode)
	campgrounds, _ := dm.parkService.GetParkCampgrounds(parkCode)
	fees, _ := dm.parkService.GetParkFees(parkCode)
	parking, _ := dm.parkService.GetParkParking(parkCode)

	// Comprehensive details data structure with all available NPS API fields
	data := map[string]interface{}{
		"VisitorCenters": visitors_centers,
		"Campgrounds":    campgrounds,
		"Fees":           fees,
		"Parking":        parking,
		"ParkCode":       parkCode,
	}

	// Parse and execute the details template
	tmpl, err := template.ParseFiles("web/templates/partials/park-details.html")
	if err != nil {
		log.Printf("Failed to load details template: %v", err)
		http.Error(w, fmt.Sprintf("Failed to load details template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Template execution error for park %s: %v", parkCode, err)
		http.Error(w, "Failed to render park details", http.StatusInternalServerError)
		return
	}
}
