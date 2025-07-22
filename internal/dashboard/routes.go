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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ztkent/go-nps"
	"github.com/ztkent/nps-dashboard/internal/database"
)

// UnifiedNewsItem represents a normalized news item that can handle NewsRelease, Article, and Alert types
type UnifiedNewsItem struct {
	ID           string               `json:"id"`
	Title        string               `json:"title"`
	URL          string               `json:"url"`
	Description  string               `json:"description"`
	ReleaseDate  string               `json:"releaseDate,omitempty"`
	Image        *UnifiedImageInfo    `json:"image,omitempty"`
	RelatedParks []UnifiedRelatedPark `json:"relatedParks,omitempty"`
	Category     string               `json:"category,omitempty"` // For alerts
}

// UnifiedImageInfo represents normalized image information
type UnifiedImageInfo struct {
	URL     string `json:"url"`
	AltText string `json:"altText"`
	Title   string `json:"title"`
}

// UnifiedRelatedPark represents normalized related park information
type UnifiedRelatedPark struct {
	Name        string `json:"name"`
	ParkCode    string `json:"parkCode"`
	States      string `json:"states"`
	FullName    string `json:"fullName"`
	URL         string `json:"url"`
	Designation string `json:"designation"`
}

// UnifiedNewsData represents the normalized response structure
type UnifiedNewsData struct {
	Total string            `json:"total"`
	Data  []UnifiedNewsItem `json:"data"`
	Limit string            `json:"limit"`
	Start string            `json:"start"`
}

// normalizeNewsReleaseData converts NewsReleaseResponse to UnifiedNewsData
func normalizeNewsReleaseData(data *nps.NewsReleaseResponse) *UnifiedNewsData {
	unified := &UnifiedNewsData{
		Total: data.Total,
		Limit: data.Limit,
		Start: data.Start,
		Data:  make([]UnifiedNewsItem, len(data.Data)),
	}

	for i, item := range data.Data {
		unifiedItem := UnifiedNewsItem{
			ID:          item.ID,
			Title:       item.Title,
			URL:         item.Url,
			Description: item.Abstract,
			ReleaseDate: item.ReleaseDate,
		}

		// Handle image
		if item.Image.Url != "" {
			unifiedItem.Image = &UnifiedImageInfo{
				URL:     item.Image.Url,
				AltText: item.Image.AltText,
				Title:   item.Image.Title,
			}
		}

		// Handle related parks
		unifiedItem.RelatedParks = make([]UnifiedRelatedPark, len(item.RelatedParks))
		for j, park := range item.RelatedParks {
			unifiedItem.RelatedParks[j] = UnifiedRelatedPark{
				Name:        park.Name,
				ParkCode:    park.ParkCode,
				States:      park.States,
				FullName:    park.FullName,
				URL:         park.Url,
				Designation: park.Designation,
			}
		}

		unified.Data[i] = unifiedItem
	}

	return unified
}

// normalizeArticleData converts ArticleData to UnifiedNewsData
func normalizeArticleData(data *nps.ArticleData) *UnifiedNewsData {
	unified := &UnifiedNewsData{
		Total: data.Total,
		Limit: data.Limit,
		Start: data.Start,
		Data:  make([]UnifiedNewsItem, len(data.Data)),
	}

	for i, item := range data.Data {
		unifiedItem := UnifiedNewsItem{
			ID:          item.ID,
			Title:       item.Title,
			URL:         item.URL,
			Description: item.ListingDescription,
		}

		// Handle listing image
		if item.ListingImage.URL != "" {
			unifiedItem.Image = &UnifiedImageInfo{
				URL:     item.ListingImage.URL,
				AltText: item.ListingImage.AltText,
				Title:   item.ListingImage.Title,
			}
		}

		// Handle related parks
		unifiedItem.RelatedParks = make([]UnifiedRelatedPark, len(item.RelatedParks))
		for j, park := range item.RelatedParks {
			unifiedItem.RelatedParks[j] = UnifiedRelatedPark{
				Name:        park.Name,
				ParkCode:    park.ParkCode,
				States:      park.States,
				FullName:    park.FullName,
				URL:         park.URL,
				Designation: park.Designation,
			}
		}

		unified.Data[i] = unifiedItem
	}

	return unified
}

// normalizeAlertData converts AlertResponse to UnifiedNewsData
func normalizeAlertData(data *nps.AlertResponse) *UnifiedNewsData {
	unified := &UnifiedNewsData{
		Total: data.Total,
		Limit: data.Limit,
		Start: data.Start,
		Data:  make([]UnifiedNewsItem, len(data.Data)),
	}

	for i, item := range data.Data {
		unifiedItem := UnifiedNewsItem{
			ID:          item.ID,
			Title:       item.Title,
			URL:         item.URL,
			Description: item.Description,
			Category:    item.Category,
		}

		// Alerts don't have images or related parks in the current structure
		unified.Data[i] = unifiedItem
	}

	return unified
}

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
		// Determine current page from referer or other context
		currentPage := getCurrentPageFromRequest(r)

		data := map[string]interface{}{
			"CurrentPage": currentPage,
		}

		tmpl, err := template.ParseFiles("web/templates/header.html")
		if err != nil {
			log.Printf("Failed to load header template: %v", err)
			http.Error(w, fmt.Sprintf("Failed to load header template: %v", err), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, data)
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

// getCurrentPageFromRequest determines the current page based on the request
func getCurrentPageFromRequest(r *http.Request) string {
	// Try to get referer header first
	referer := r.Header.Get("Referer")
	if referer != "" {
		if strings.Contains(referer, "/things-to-do") {
			return "things-to-do"
		}
		if strings.Contains(referer, "/get-involved") {
			return "get-involved"
		}
		if strings.Contains(referer, "/learn-more") {
			return "learn-more"
		}
		if strings.Contains(referer, "/news") {
			return "news"
		}
	}

	// Fallback: check if there's a specific header or query parameter
	currentPage := r.Header.Get("X-Current-Page")
	if currentPage != "" {
		return currentPage
	}

	currentPage = r.URL.Query().Get("current-page")
	if currentPage != "" {
		return currentPage
	}

	// Default to home if we can't determine
	return "home"
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
	// Prepare data for the template with proper structure
	data := map[string]interface{}{
		"ThingsToDo":     thingsToDo,
		"Activities":     activities,
		"VisitorCenters": visitorCenters,
		"Amenities":      amenities,
		"ParkTours":      tours,
		"ParkEvents":     events,
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
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
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

// ThingsToDoPageHandler returns the full Things To Do page
func (dm *Dashboard) ThingsToDoPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Get available parks for filter dropdown
	parks, err := dm.parkService.GetAllParks()
	if err != nil {
		log.Printf("Failed to get parks for Things To Do page: %v", err)
		parks = []database.CachedPark{}
	}

	// Get available activities for filter dropdown
	activities, err := dm.parkService.GetAllActivities()
	if err != nil {
		log.Printf("Failed to get activities for Things To Do page: %v", err)
		activities = &nps.ActivityResponse{Data: []nps.Activity{}}
	}

	// Get unique states from parks
	statesMap := make(map[string]bool)
	for _, park := range parks {
		states := strings.Split(park.States, ",")
		for _, state := range states {
			state = strings.TrimSpace(state)
			if state != "" {
				statesMap[state] = true
			}
		}
	}

	states := make([]string, 0, len(statesMap))
	for state := range statesMap {
		states = append(states, state)
	}

	// Sort states alphabetically
	sort.Strings(states)

	data := map[string]interface{}{
		"Parks":      parks,
		"Activities": activities.Data,
		"States":     states,
	}

	tmpl, err := template.ParseFiles("web/templates/things-to-do.html")
	if err != nil {
		log.Printf("Failed to load Things To Do template: %v", err)
		http.Error(w, fmt.Sprintf("Failed to load Things To Do template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to render Things To Do page: %v", err)
		http.Error(w, fmt.Sprintf("Failed to render Things To Do page: %v", err), http.StatusInternalServerError)
		return
	}
}

// ThingsToDoSearchHandler handles search and filtering for things to do
func (dm *Dashboard) ThingsToDoSearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Parse query parameters
	query := r.URL.Query().Get("q")
	if searchInput := r.URL.Query().Get("activity-search"); searchInput != "" {
		query = searchInput
	}

	parkCode := r.URL.Query().Get("parkCode")
	stateCode := r.URL.Query().Get("stateCode")
	activityId := r.URL.Query().Get("activityId")
	difficulty := r.URL.Query().Get("difficulty")

	startStr := r.URL.Query().Get("start")
	limitStr := r.URL.Query().Get("limit")

	start := 0
	limit := 3500

	if startStr != "" {
		if parsedStart, err := strconv.Atoi(startStr); err == nil && parsedStart >= 0 {
			start = parsedStart
		}
	}

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 500 {
			limit = parsedLimit
		}
	}

	// Call NPS API with filters
	thingsToDoResponse, err := dm.parkService.SearchThingsToDo(activityId, parkCode, stateCode, query, limit, start)
	if err != nil {
		log.Printf("Failed to search things to do: %v", err)
		thingsToDoResponse = &nps.ThingsToDoResponse{
			Total: "0",
			Data:  []nps.ThingsToDo{},
			Limit: strconv.Itoa(limit),
			Start: strconv.Itoa(start),
		}
	}

	// Filter by difficulty if specified (this might need to be done client-side since NPS API doesn't have difficulty filter)
	if difficulty != "" && thingsToDoResponse != nil {
		filteredData := []nps.ThingsToDo{}
		for _, activity := range thingsToDoResponse.Data {
			if matchesDifficulty(activity, difficulty) {
				filteredData = append(filteredData, activity)
			}
		}
		thingsToDoResponse.Data = filteredData
	}

	data := map[string]interface{}{
		"ThingsToDoData": thingsToDoResponse,
	}

	// Create template with function map and parse the file
	tmpl := template.New("things-to-do-results.html").Funcs(template.FuncMap{
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"slugify": func(s string) string {
			return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", "-"))
		},
		"atoi": func(s string) int {
			i, _ := strconv.Atoi(s)
			return i
		},
		"add": func(a, b int) int {
			return a + b
		},
		"unescapeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	})

	tmpl, err = tmpl.ParseFiles("web/templates/partials/things-to-do-results.html")

	if err != nil {
		log.Printf("Failed to load Things To Do results template: %v", err)
		http.Error(w, fmt.Sprintf("Failed to load Things To Do results template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to render Things To Do results: %v", err)
		http.Error(w, fmt.Sprintf("Failed to render Things To Do results: %v", err), http.StatusInternalServerError)
		return
	}
}

// Helper function to match difficulty (since NPS API doesn't provide difficulty levels)
func matchesDifficulty(activity nps.ThingsToDo, difficulty string) bool {
	description := strings.ToLower(activity.LongDescription + " " + activity.ShortDescription + " " + activity.ActivityDescription)

	switch difficulty {
	case "easy":
		return strings.Contains(description, "easy") ||
			strings.Contains(description, "beginner") ||
			strings.Contains(description, "gentle") ||
			strings.Contains(description, "accessible")
	case "moderate":
		return strings.Contains(description, "moderate") ||
			strings.Contains(description, "intermediate") ||
			strings.Contains(description, "some experience")
	case "difficult":
		return strings.Contains(description, "difficult") ||
			strings.Contains(description, "challenging") ||
			strings.Contains(description, "advanced") ||
			strings.Contains(description, "strenuous")
	default:
		return true
	}
}

// EventsPageHandler returns the full Events page
func (dm *Dashboard) EventsPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Get available parks for filter dropdown
	parks, err := dm.parkService.GetAllParks()
	if err != nil {
		log.Printf("Failed to get parks for Events page: %v", err)
		parks = []database.CachedPark{}
	}

	// Get unique states from parks
	statesMap := make(map[string]bool)
	for _, park := range parks {
		states := strings.Split(park.States, ",")
		for _, state := range states {
			state = strings.TrimSpace(state)
			if state != "" {
				statesMap[state] = true
			}
		}
	}

	states := make([]string, 0, len(statesMap))
	for state := range statesMap {
		states = append(states, state)
	}

	// Sort states alphabetically
	sort.Strings(states)

	// Execute the template with parks and states data
	data := map[string]interface{}{
		"CurrentPage": "events",
		"Parks":       parks,
		"States":      states,
	}

	// Load and parse the events template
	tmpl, err := template.ParseFiles("web/templates/events.html")
	if err != nil {
		log.Printf("Error parsing events template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing events template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// EventsSearchHandler handles search and filtering for events
func (dm *Dashboard) EventsSearchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query().Get("q")
	parkCode := r.URL.Query().Get("park")
	stateCode := r.URL.Query().Get("state")
	eventType := r.URL.Query().Get("event_type")
	dateStart := r.URL.Query().Get("date_start")
	dateEnd := r.URL.Query().Get("date_end")
	date := r.URL.Query().Get("date") // Single date selection

	limitStr := r.URL.Query().Get("limit")
	startStr := r.URL.Query().Get("start")

	limit := 50
	start := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if startStr != "" {
		if s, err := strconv.Atoi(startStr); err == nil {
			start = s
		}
	}

	// Handle single date selection
	if date != "" {
		dateStart = date
		dateEnd = date
	}

	// If no date range specified, default to next 3 months
	if dateStart == "" && dateEnd == "" {
		now := time.Now()
		dateStart = now.Format("2006-01-02")
		dateEnd = now.AddDate(0, 3, 0).Format("2006-01-02")
	}

	// Search events using the park service
	eventsData, err := dm.parkService.SearchEvents(query, parkCode, stateCode, eventType, dateStart, dateEnd, limit, start)
	if err != nil {
		log.Printf("Error searching events: %v", err)
		http.Error(w, "Error searching events", http.StatusInternalServerError)
		return
	}

	// Create a map of all unique 'event types' for filtering
	eventTypesMap := make(map[string]bool)
	for _, event := range eventsData.Data {
		if event.Types != nil {
			for _, eventType := range event.Types {
				eventTypesMap[eventType] = true
			}
		}
	}

	// Determine date range description for display
	var dateRange string
	if dateStart != "" && dateEnd != "" {
		if dateStart == dateEnd {
			if t, err := time.Parse("2006-01-02", dateStart); err == nil {
				dateRange = t.Format("January 2, 2006")
			}
		} else {
			if t1, err1 := time.Parse("2006-01-02", dateStart); err1 == nil {
				if t2, err2 := time.Parse("2006-01-02", dateEnd); err2 == nil {
					dateRange = fmt.Sprintf("%s - %s",
						t1.Format("January 2, 2006"),
						t2.Format("January 2, 2006"))
				}
			}
		}
	}

	// Create template data
	data := struct {
		EventsData *nps.EventResponse
		DateRange  string
	}{
		EventsData: eventsData,
		DateRange:  dateRange,
	}

	// Process events to filter future dates and limit additional dates
	today := time.Now().Format("2006-01-02")
	for i := range data.EventsData.Data {
		event := &data.EventsData.Data[i]
		if len(event.Dates) > 0 {
			// Filter dates to only include future dates
			var futureDates []string
			for _, dateStr := range event.Dates {
				if dateStr >= today {
					futureDates = append(futureDates, dateStr)
					// Limit to 10 additional dates
					if len(futureDates) >= 10 {
						break
					}
				}
			}
			event.Dates = futureDates
		}
	}
	// Remove any duplicate events based on ID
	uniqueEvents := make(map[string]nps.Event)
	for _, event := range data.EventsData.Data {
		if event.Title != "" {
			// Use event ID as key to ensure uniqueness
			uniqueEvents[event.Title] = event
		} else if event.EventID != "" {
			// Fallback to EventID if ID is not available
			uniqueEvents[event.EventID] = event
		}
	}

	data.EventsData.Data = make([]nps.Event, 0, len(uniqueEvents))
	for _, event := range uniqueEvents {
		data.EventsData.Data = append(data.EventsData.Data, event)
	}

	// Sort events by start date
	sort.Slice(data.EventsData.Data, func(i, j int) bool {
		if len(data.EventsData.Data[i].Dates) > 0 && len(data.EventsData.Data[j].Dates) > 0 {
			dateI, errI := time.Parse("2006-01-02", data.EventsData.Data[i].Dates[0])
			dateJ, errJ := time.Parse("2006-01-02", data.EventsData.Data[j].Dates[0])
			if errI == nil && errJ == nil {
				return dateI.Before(dateJ)
			}
		}
		return data.EventsData.Data[i].Title < data.EventsData.Data[j].Title
	})

	// Load and parse the events results template
	tmpl, err := template.New("events-results.html").Funcs(template.FuncMap{
		"unescapeHTML": func(s string) template.HTML {
			return template.HTML(html.UnescapeString(s))
		},
		"formatEventDate": func(dateStr string) string {
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				return t.Format("January 2, 2006")
			}
			if t, err := time.Parse("2006-01-02T15:04:05Z", dateStr); err == nil {
				return t.Format("January 2, 2006")
			}
			return dateStr
		},
		"fullImageURL": func(url string) string {
			if strings.HasPrefix(url, "http") {
				return url
			}
			if strings.HasPrefix(url, "/") {
				return "https://www.nps.gov" + url
			}
			return url
		},
		"lower": strings.ToLower,
		"atoi": func(s string) int {
			if i, err := strconv.Atoi(s); err == nil {
				return i
			}
			return 0
		},
		"add": func(a, b int) int {
			return a + b
		},
	}).ParseFiles("web/templates/partials/events-results.html")

	if err != nil {
		log.Printf("Error parsing events results template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing events results template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// CampingPageHandler returns the full Camping page
func (dm *Dashboard) CampingPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Get available parks for filter dropdown
	parks, err := dm.parkService.GetAllParks()
	if err != nil {
		log.Printf("Failed to get parks for Camping page: %v", err)
		parks = []database.CachedPark{}
	}

	// Get unique states from parks
	statesMap := make(map[string]bool)
	for _, park := range parks {
		states := strings.Split(park.States, ",")
		for _, state := range states {
			state = strings.TrimSpace(state)
			if state != "" {
				statesMap[state] = true
			}
		}
	}

	states := make([]string, 0, len(statesMap))
	for state := range statesMap {
		states = append(states, state)
	}

	// Sort states alphabetically
	sort.Strings(states)

	// Execute the template with parks and states data
	data := map[string]interface{}{
		"CurrentPage": "camping",
		"Parks":       parks,
		"States":      states,
	}

	// Load and parse the camping template
	tmpl, err := template.ParseFiles("web/templates/camping.html")
	if err != nil {
		log.Printf("Error parsing camping template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing camping template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// CampingSearchHandler handles search and filtering for campgrounds
func (dm *Dashboard) CampingSearchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query().Get("q")
	parkCode := r.URL.Query().Get("park")
	stateCode := r.URL.Query().Get("state")
	amenityType := r.URL.Query().Get("amenity_type")

	limitStr := r.URL.Query().Get("limit")
	startStr := r.URL.Query().Get("start")

	limit := 50
	start := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if startStr != "" {
		if s, err := strconv.Atoi(startStr); err == nil {
			start = s
		}
	}

	// Search campgrounds using the park service
	campgroundsData, err := dm.parkService.SearchCampgrounds(query, parkCode, stateCode, limit, start)
	if err != nil {
		log.Printf("Error searching campgrounds: %v", err)
		http.Error(w, "Error searching campgrounds", http.StatusInternalServerError)
		return
	}

	// Filter by amenity type if specified
	if amenityType != "" {
		filteredData := []nps.Campground{}
		for _, campground := range campgroundsData.Data {
			if matchesCampgroundAmenity(campground, amenityType) {
				filteredData = append(filteredData, campground)
			}
		}
		campgroundsData.Data = filteredData
		campgroundsData.Total = fmt.Sprintf("%d", len(filteredData))
	}

	// Create template data
	data := struct {
		CampgroundsData *nps.CampgroundData
	}{
		CampgroundsData: campgroundsData,
	}

	// Load and parse the camping results template
	tmpl, err := template.New("camping-results.html").Funcs(template.FuncMap{
		"unescapeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"formatCampgroundFee": func(fees []struct {
			Cost        string `json:"cost"`
			Description string `json:"description"`
			Title       string `json:"title"`
		}) string {
			if len(fees) == 0 {
				return "Contact campground for fee information"
			}
			return fees[0].Cost
		},
		"formatCampgroundAmenities": func(amenities struct {
			TrashRecyclingCollection   string   `json:"trashRecyclingCollection"`
			Toilets                    []string `json:"toilets"`
			InternetConnectivity       string   `json:"internetConnectivity"`
			Showers                    []string `json:"showers"`
			CellPhoneReception         string   `json:"cellPhoneReception"`
			Laundry                    string   `json:"laundry"`
			Amphitheater               string   `json:"amphitheater"`
			DumpStation                string   `json:"dumpStation"`
			CampStore                  string   `json:"campStore"`
			StaffOrVolunteerHostOnsite string   `json:"staffOrVolunteerHostOnsite"`
			PotableWater               []string `json:"potableWater"`
			IceAvailableForSale        string   `json:"iceAvailableForSale"`
			FirewoodForSale            string   `json:"firewoodForSale"`
			FoodStorageLockers         string   `json:"foodStorageLockers"`
		}) []string {
			var result []string
			if len(amenities.Toilets) > 0 {
				result = append(result, "Toilets")
			}
			if len(amenities.Showers) > 0 {
				result = append(result, "Showers")
			}
			if amenities.DumpStation == "true" {
				result = append(result, "Dump Station")
			}
			if amenities.CampStore == "true" {
				result = append(result, "Camp Store")
			}
			if amenities.Laundry == "true" {
				result = append(result, "Laundry")
			}
			if amenities.Amphitheater == "true" {
				result = append(result, "Amphitheater")
			}
			return result
		},
		"fullImageURL": func(url string) string {
			if strings.HasPrefix(url, "http") {
				return url
			}
			if strings.HasPrefix(url, "/") {
				return "https://www.nps.gov" + url
			}
			return url
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"atoi": func(s string) int {
			if i, err := strconv.Atoi(s); err == nil {
				return i
			}
			return 0
		},
		"add": func(a, b int) int {
			return a + b
		},
	}).ParseFiles("web/templates/partials/camping-results.html")

	if err != nil {
		log.Printf("Error parsing camping results template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing camping results template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// NewsPageHandler returns the full News page
func (dm *Dashboard) NewsPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Get available parks for filter dropdown
	parks, err := dm.parkService.GetAllParks()
	if err != nil {
		log.Printf("Failed to get parks for News page: %v", err)
		parks = []database.CachedPark{}
	}

	// Get unique states from parks
	statesMap := make(map[string]bool)
	for _, park := range parks {
		states := strings.Split(park.States, ",")
		for _, state := range states {
			state = strings.TrimSpace(state)
			if state != "" {
				statesMap[state] = true
			}
		}
	}

	states := make([]string, 0, len(statesMap))
	for state := range statesMap {
		states = append(states, state)
	}

	// Sort states alphabetically
	sort.Strings(states)

	// Execute the template with parks and states data
	data := map[string]interface{}{
		"CurrentPage": "news",
		"Parks":       parks,
		"States":      states,
	}

	// Load and parse the news template
	tmpl, err := template.ParseFiles("web/templates/news.html")
	if err != nil {
		log.Printf("Error parsing news template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing news template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// NewsSearchHandler handles search and filtering for news
func (dm *Dashboard) NewsSearchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query().Get("q")
	parkCode := r.URL.Query().Get("park")
	stateCode := r.URL.Query().Get("state")
	newsType := r.URL.Query().Get("news_type")

	limitStr := r.URL.Query().Get("limit")
	startStr := r.URL.Query().Get("start")

	limit := 50
	start := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if startStr != "" {
		if s, err := strconv.Atoi(startStr); err == nil {
			start = s
		}
	}

	// Search news using the park service
	newsData, err := dm.parkService.SearchNews(query, parkCode, stateCode, newsType, limit, start)
	if err != nil {
		log.Printf("Error searching news: %v", err)
		http.Error(w, "Error searching news", http.StatusInternalServerError)
		return
	}

	// Normalize the different data types into a unified structure
	var unifiedData *UnifiedNewsData
	switch newsType {
	case "articles":
		if articleData, ok := newsData.(*nps.ArticleData); ok {
			unifiedData = normalizeArticleData(articleData)
		}
	case "alerts":
		if alertData, ok := newsData.(*nps.AlertResponse); ok {
			unifiedData = normalizeAlertData(alertData)
		}
	default:
		// Default to news releases
		if newsReleaseData, ok := newsData.(*nps.NewsReleaseResponse); ok {
			unifiedData = normalizeNewsReleaseData(newsReleaseData)
		}
	}

	// Handle case where normalization failed
	if unifiedData == nil {
		log.Printf("Failed to normalize news data for type: %s", newsType)
		unifiedData = &UnifiedNewsData{
			Total: "0",
			Data:  []UnifiedNewsItem{},
			Limit: strconv.Itoa(limit),
			Start: strconv.Itoa(start),
		}
	}

	// Create template data
	data := struct {
		NewsData *UnifiedNewsData
	}{
		NewsData: unifiedData,
	}

	// Load and parse the news results template
	tmpl, err := template.New("news-results.html").Funcs(template.FuncMap{
		"unescapeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"formatNewsDate": func(dateStr string) string {
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				return t.Format("January 2, 2006")
			}
			if t, err := time.Parse("2006-01-02T15:04:05Z", dateStr); err == nil {
				return t.Format("January 2, 2006")
			}
			return dateStr
		},
		"fullImageURL": func(url string) string {
			if strings.HasPrefix(url, "http") {
				return url
			}
			if strings.HasPrefix(url, "/") {
				return "https://www.nps.gov" + url
			}
			return url
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"atoi": func(s string) int {
			if i, err := strconv.Atoi(s); err == nil {
				return i
			}
			return 0
		},
		"add": func(a, b int) int {
			return a + b
		},
	}).ParseFiles("web/templates/partials/news-results.html")

	if err != nil {
		log.Printf("Error parsing news results template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing news results template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Helper function to match campground amenity types
func matchesCampgroundAmenity(campground nps.Campground, amenityType string) bool {
	switch amenityType {
	case "showers":
		return len(campground.Amenities.Showers) > 0
	case "dump_station":
		return campground.Amenities.DumpStation == "true"
	case "laundry":
		return campground.Amenities.Laundry == "true"
	case "camp_store":
		return campground.Amenities.CampStore == "true"
	case "rv_hookups":
		return campground.Campsites.ElectricalHookups != "0" && campground.Campsites.ElectricalHookups != ""
	case "reservable":
		return campground.NumberOfSitesReservable != "0" && campground.NumberOfSitesReservable != ""
	default:
		return true
	}
}

// EventDetailsHandler returns detailed information for a specific event
func (dm *Dashboard) EventDetailsHandler(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventID")

	if eventID == "" {
		http.Error(w, "Event ID required", http.StatusBadRequest)
		return
	}

	// Try to get the event by ID first
	targetEvent, err := dm.parkService.GetEventByID(eventID)
	if err != nil {
		// Fallback: search for the event in recent events
		eventsData, searchErr := dm.parkService.SearchEvents("", "", "", "", "", "", 100, 0)
		if searchErr != nil {
			log.Printf("Error searching for event %s: %v", eventID, searchErr)
			http.Error(w, "Error fetching event details", http.StatusInternalServerError)
			return
		}

		// Find the specific event by ID
		for _, event := range eventsData.Data {
			if event.ID == eventID || event.EventID == eventID {
				targetEvent = &event
				break
			}
		}
	}

	if targetEvent == nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	// Create comprehensive template data
	data := map[string]interface{}{
		"Event": targetEvent,
	}

	// Create template with helper functions
	tmpl := template.New("event-details").Funcs(template.FuncMap{
		"unescapeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"formatEventDate": func(dateStr string) string {
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				return t.Format("January 2, 2006")
			}
			if t, err := time.Parse("2006-01-02T15:04:05Z", dateStr); err == nil {
				return t.Format("January 2, 2006")
			}
			return dateStr
		},
		"fullImageURL": func(url string) string {
			if strings.HasPrefix(url, "http") {
				return url
			}
			if strings.HasPrefix(url, "/") {
				return "https://www.nps.gov" + url
			}
			return url
		},
		"formatDateTime": func(dateStr string) string {
			if t, err := time.Parse("2006-01-02T15:04:05Z", dateStr); err == nil {
				return t.Format("January 2, 2006 at 3:04 PM")
			}
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				return t.Format("January 2, 2006")
			}
			return dateStr
		},
	})

	// Inline template for event details modal
	templateHTML := `
	<div class="event-detail-content">
		<h3>{{.Event.Title}}</h3>
		
		{{if .Event.Images}}
		<div class="event-detail-images">
			{{range .Event.Images}}
			<div class="event-detail-image">
				<img src="{{fullImageURL .Url}}" alt="{{.AltText}}" title="{{.Title}}">
				{{if .Caption}}<p class="image-caption">{{.Caption}} {{if .Credit}}({{.Credit}}){{end}}</p>{{end}}
			</div>
			{{end}}
		</div>
		{{end}}

		{{if .Event.Description}}
		<div class="event-detail-description">
			{{.Event.Description | unescapeHTML}}
		</div>
		{{end}}

		<div class="event-detail-info">
			{{if .Event.ParkFullName}}
			<div class="detail-row">
				<strong>Location:</strong> {{.Event.ParkFullName}}
				{{if .Event.Location}} - {{.Event.Location}}{{end}}
			</div>
			{{end}}

			{{if .Event.DateStart}}
			<div class="detail-row">
				<strong>Date:</strong>
				{{if .Event.DateEnd}}
					{{if eq .Event.DateStart .Event.DateEnd}}
						{{.Event.DateStart | formatEventDate}}
					{{else}}
						{{.Event.DateStart | formatEventDate}} - {{.Event.DateEnd | formatEventDate}}
					{{end}}
				{{else}}
					{{.Event.DateStart | formatEventDate}}
				{{end}}
			</div>
			{{end}}

			{{if .Event.Times}}
			<div class="detail-row">
				<strong>Times:</strong>
				<ul class="event-times-list">
				{{range .Event.Times}}
					<li>{{.TimeStart}}{{if .TimeEnd}} - {{.TimeEnd}}{{end}}</li>
				{{end}}
				</ul>
			</div>
			{{end}}

			{{if .Event.Category}}
			<div class="detail-row">
				<strong>Category:</strong> {{.Event.Category}}
			</div>
			{{end}}

			{{if .Event.Types}}
			<div class="detail-row">
				<strong>Event Types:</strong> {{range $index, $type := .Event.Types}}{{if $index}}, {{end}}{{$type}}{{end}}
			</div>
			{{end}}

			<div class="detail-row">
				<strong>Cost:</strong> {{if eq .Event.IsFree "true"}}Free{{else}}Fee Required{{end}}
			</div>

			{{if .Event.FeeInfo}}
			<div class="detail-row">
				<strong>Fee Information:</strong> {{.Event.FeeInfo}}
			</div>
			{{end}}

			{{if eq .Event.IsRegresRequired "true"}}
			<div class="detail-row">
				<strong>Registration:</strong> Required
			</div>
			{{end}}

			{{if .Event.RegresInfo}}
			<div class="detail-row">
				<strong>Registration Info:</strong> {{.Event.RegresInfo}}
			</div>
			{{end}}

			{{if .Event.RegresUrl}}
			<div class="detail-row">
				<strong>Registration:</strong> <a href="{{.Event.RegresUrl}}" target="_blank" rel="noopener">Register Here</a>
			</div>
			{{end}}

			{{if .Event.ContactName}}
			<div class="detail-row">
				<strong>Contact:</strong> {{.Event.ContactName}}
				{{if .Event.ContactTelephoneNumber}} - {{.Event.ContactTelephoneNumber}}{{end}}
				{{if .Event.ContactEmailAddress}} - {{.Event.ContactEmailAddress}}{{end}}
			</div>
			{{end}}

			{{if eq .Event.IsRecurring "true"}}
			<div class="detail-row">
				<strong>Recurring:</strong> Yes
				{{if .Event.RecurrenceDateStart}} ({{.Event.RecurrenceDateStart | formatEventDate}}{{if .Event.RecurrenceDateEnd}} - {{.Event.RecurrenceDateEnd | formatEventDate}}{{end}}){{end}}
			</div>
			{{end}}

			{{if .Event.Tags}}
			<div class="detail-row">
				<strong>Tags:</strong>
				<div class="event-detail-tags">
					{{range .Event.Tags}}
					<span class="event-detail-tag">{{.}}</span>
					{{end}}
				</div>
			</div>
			{{end}}

			{{if .Event.InfoUrl}}
			<div class="detail-row">
				<strong>More Information:</strong> <a href="{{.Event.InfoUrl}}" target="_blank" rel="noopener">Event Details</a>
			</div>
			{{end}}

			{{if .Event.Dates}}
			<div class="detail-row">
				<strong>Additional Dates:</strong>
				<div class="event-additional-dates">
					{{range .Event.Dates}}
					<span class="additional-date">{{. | formatEventDate}}</span>
					{{end}}
				</div>
			</div>
			{{end}}
		</div>
	</div>
	`

	tmpl, err = tmpl.Parse(templateHTML)
	if err != nil {
		log.Printf("Error parsing event details template: %v", err)
		http.Error(w, "Error rendering event details", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing event details template: %v", err)
		http.Error(w, "Error rendering event details", http.StatusInternalServerError)
		return
	}
}
