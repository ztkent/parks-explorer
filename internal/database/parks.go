package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ztkent/go-nps"
)

// Park represents a cached park from the database
type CachedPark struct {
	ID             int         `json:"id"`
	ParkCode       string      `json:"park_code"`
	Name           string      `json:"name"`
	FullName       string      `json:"full_name"`
	Slug           string      `json:"slug"`
	States         string      `json:"states"`
	Designation    string      `json:"designation"`
	Description    string      `json:"description"`
	WeatherInfo    string      `json:"weather_info"`
	DirectionsInfo string      `json:"directions_info"`
	URL            string      `json:"url"`
	DirectionsURL  string      `json:"directions_url"`
	Latitude       string      `json:"latitude"`
	Longitude      string      `json:"longitude"`
	LatLong        string      `json:"lat_long"`
	RelevanceScore float64     `json:"relevance_score"`
	APIData        string      `json:"api_data"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	LastFetchedAt  time.Time   `json:"last_fetched_at"`
	Images         []ParkImage `json:"images"`
}

// ParkImage represents a cached park image
type ParkImage struct {
	ID         int       `json:"id"`
	ParkID     int       `json:"park_id"`
	URL        string    `json:"url"`
	Title      string    `json:"title"`
	AltText    string    `json:"alt_text"`
	Caption    string    `json:"caption"`
	Credit     string    `json:"credit"`
	ImageOrder int       `json:"image_order"`
	CreatedAt  time.Time `json:"created_at"`
}

// CachedParkData represents cached park-specific data
type CachedParkData struct {
	ID            int       `json:"id"`
	ParkID        int       `json:"park_id"`
	DataType      string    `json:"data_type"`
	APIData       string    `json:"api_data"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	LastFetchedAt time.Time `json:"last_fetched_at"`
}

// UpsertPark inserts or updates a park in the database
func (db *DB) UpsertPark(parkData interface{}, slug string) (*CachedPark, error) {
	// Convert park data to JSON for storage
	apiDataJSON, err := json.Marshal(parkData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal park data: %w", err)
	}

	// Extract fields from the park data (assuming it's a map or struct)
	parkMap, ok := parkData.(map[string]interface{})
	if !ok {
		// Try to convert to map via JSON
		var tempMap map[string]interface{}
		if err := json.Unmarshal(apiDataJSON, &tempMap); err != nil {
			return nil, fmt.Errorf("failed to convert park data to map: %w", err)
		}
		parkMap = tempMap
	}

	// Extract values with safe type assertions
	parkCode := getString(parkMap, "parkCode")
	name := getString(parkMap, "name")
	fullName := getString(parkMap, "fullName")
	states := getString(parkMap, "states")
	designation := getString(parkMap, "designation")
	description := getString(parkMap, "description")
	weatherInfo := getString(parkMap, "weatherInfo")
	directionsInfo := getString(parkMap, "directionsInfo")
	url := getString(parkMap, "url")
	directionsURL := getString(parkMap, "directionsUrl")
	latitude := getString(parkMap, "latitude")
	longitude := getString(parkMap, "longitude")
	latLong := getString(parkMap, "latLong")
	relevanceScore := getFloat64(parkMap, "relevanceScore")

	// Upsert park using SQLite syntax
	query := `
		INSERT OR REPLACE INTO parks (
			park_code, name, full_name, slug, states, designation, description,
			weather_info, directions_info, url, directions_url, latitude, longitude,
			lat_long, relevance_score, api_data, updated_at, last_fetched_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := db.Exec(query,
		parkCode, name, fullName, slug, states, designation, description,
		weatherInfo, directionsInfo, url, directionsURL, latitude, longitude,
		latLong, relevanceScore, string(apiDataJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to upsert park: %w", err)
	}

	// Get the park ID
	var parkID int64
	if parkCode != "" {
		row := db.QueryRow("SELECT id FROM parks WHERE park_code = ?", parkCode)
		if err := row.Scan(&parkID); err != nil {
			return nil, fmt.Errorf("failed to get park ID: %w", err)
		}
	} else {
		parkID, err = result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert ID: %w", err)
		}
	}

	// Handle images
	if images, ok := parkMap["images"].([]interface{}); ok {
		if err := db.UpsertParkImages(int(parkID), images); err != nil {
			return nil, fmt.Errorf("failed to upsert park images: %w", err)
		}
	}

	// Return the cached park
	return db.GetParkByID(int(parkID))
}

// UpsertParkImages inserts or updates park images
func (db *DB) UpsertParkImages(parkID int, images []interface{}) error {
	// Delete existing images for this park
	_, err := db.Exec("DELETE FROM park_images WHERE park_id = ?", parkID)
	if err != nil {
		return fmt.Errorf("failed to delete existing images: %w", err)
	}

	// Insert new images
	query := `
		INSERT INTO park_images (park_id, url, title, alt_text, caption, credit, image_order)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	for i, imgData := range images {
		if imgMap, ok := imgData.(map[string]interface{}); ok {
			url := getString(imgMap, "url")
			title := getString(imgMap, "title")
			altText := getString(imgMap, "altText")
			caption := getString(imgMap, "caption")
			credit := getString(imgMap, "credit")

			_, err := db.Exec(query, parkID, url, title, altText, caption, credit, i)
			if err != nil {
				return fmt.Errorf("failed to insert image: %w", err)
			}
		}
	}

	return nil
}

// GetParkByID retrieves a park by ID with its images
func (db *DB) GetParkByID(id int) (*CachedPark, error) {
	query := `
		SELECT id, park_code, name, full_name, slug, states, designation, description,
			   weather_info, directions_info, url, directions_url, latitude, longitude,
			   lat_long, relevance_score, api_data, created_at, updated_at, last_fetched_at
		FROM parks WHERE id = ?
	`

	var park CachedPark
	row := db.QueryRow(query, id)
	err := row.Scan(
		&park.ID, &park.ParkCode, &park.Name, &park.FullName, &park.Slug,
		&park.States, &park.Designation, &park.Description, &park.WeatherInfo,
		&park.DirectionsInfo, &park.URL, &park.DirectionsURL, &park.Latitude,
		&park.Longitude, &park.LatLong, &park.RelevanceScore, &park.APIData,
		&park.CreatedAt, &park.UpdatedAt, &park.LastFetchedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get park: %w", err)
	}

	// Load images
	images, err := db.GetParkImages(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get park images: %w", err)
	}
	park.Images = images

	return &park, nil
}

// GetParkBySlug retrieves a park by slug with its images
func (db *DB) GetParkBySlug(slug string) (*CachedPark, error) {
	query := `
		SELECT id, park_code, name, full_name, slug, states, designation, description,
			   weather_info, directions_info, url, directions_url, latitude, longitude,
			   lat_long, relevance_score, api_data, created_at, updated_at, last_fetched_at
		FROM parks WHERE slug = ?
	`

	var park CachedPark
	row := db.QueryRow(query, slug)
	err := row.Scan(
		&park.ID, &park.ParkCode, &park.Name, &park.FullName, &park.Slug,
		&park.States, &park.Designation, &park.Description, &park.WeatherInfo,
		&park.DirectionsInfo, &park.URL, &park.DirectionsURL, &park.Latitude,
		&park.Longitude, &park.LatLong, &park.RelevanceScore, &park.APIData,
		&park.CreatedAt, &park.UpdatedAt, &park.LastFetchedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get park: %w", err)
	}

	// Load images
	images, err := db.GetParkImages(park.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get park images: %w", err)
	}
	park.Images = images

	return &park, nil
}

// GetParkImages retrieves images for a park
func (db *DB) GetParkImages(parkID int) ([]ParkImage, error) {
	query := `
		SELECT id, park_id, url, title, alt_text, caption, credit, image_order, created_at
		FROM park_images WHERE park_id = ? ORDER BY image_order
	`

	rows, err := db.Query(query, parkID)
	if err != nil {
		return nil, fmt.Errorf("failed to query park images: %w", err)
	}
	defer rows.Close()

	var images []ParkImage
	for rows.Next() {
		var img ParkImage
		err := rows.Scan(&img.ID, &img.ParkID, &img.URL, &img.Title,
			&img.AltText, &img.Caption, &img.Credit, &img.ImageOrder, &img.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, img)
	}

	return images, nil
}

// GetAllParks retrieves all cached parks
func (db *DB) GetAllParks() ([]CachedPark, error) {
	return db.GetParksWithPagination(0, -1) // -1 means no limit for backward compatibility
}

// GetParksWithPagination retrieves cached parks with pagination support
func (db *DB) GetParksWithPagination(offset, limit int) ([]CachedPark, error) {
	query := `
		SELECT id, park_code, name, full_name, slug, states, designation, description,
			   weather_info, directions_info, url, directions_url, latitude, longitude,
			   lat_long, relevance_score, api_data, created_at, updated_at, last_fetched_at
		FROM parks ORDER BY name
	`

	var args []interface{}
	if limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query parks: %w", err)
	}
	defer rows.Close()

	var parks []CachedPark
	for rows.Next() {
		var park CachedPark
		err := rows.Scan(
			&park.ID, &park.ParkCode, &park.Name, &park.FullName, &park.Slug,
			&park.States, &park.Designation, &park.Description, &park.WeatherInfo,
			&park.DirectionsInfo, &park.URL, &park.DirectionsURL, &park.Latitude,
			&park.Longitude, &park.LatLong, &park.RelevanceScore, &park.APIData,
			&park.CreatedAt, &park.UpdatedAt, &park.LastFetchedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan park: %w", err)
		}

		// Load images for each park
		images, err := db.GetParkImages(park.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get park images: %w", err)
		}
		park.Images = images

		parks = append(parks, park)
	}

	return parks, nil
}

// SearchParks searches cached parks by name
func (db *DB) SearchParks(query string) ([]CachedPark, error) {
	searchQuery := `
		SELECT id, park_code, name, full_name, slug, states, designation, description,
			   weather_info, directions_info, url, directions_url, latitude, longitude,
			   lat_long, relevance_score, api_data, created_at, updated_at, last_fetched_at
		FROM parks 
		WHERE name LIKE ? OR full_name LIKE ? OR description LIKE ?
		ORDER BY 
			CASE 
				WHEN name LIKE ? THEN 1
				WHEN full_name LIKE ? THEN 2
				ELSE 3
			END,
			name
	`

	searchTerm := "%" + strings.ToLower(query) + "%"
	exactTerm := strings.ToLower(query) + "%"

	rows, err := db.Query(searchQuery, searchTerm, searchTerm, searchTerm, exactTerm, exactTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search parks: %w", err)
	}
	defer rows.Close()

	var parks []CachedPark
	for rows.Next() {
		var park CachedPark
		err := rows.Scan(
			&park.ID, &park.ParkCode, &park.Name, &park.FullName, &park.Slug,
			&park.States, &park.Designation, &park.Description, &park.WeatherInfo,
			&park.DirectionsInfo, &park.URL, &park.DirectionsURL, &park.Latitude,
			&park.Longitude, &park.LatLong, &park.RelevanceScore, &park.APIData,
			&park.CreatedAt, &park.UpdatedAt, &park.LastFetchedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan park: %w", err)
		}

		// Load images for each park
		images, err := db.GetParkImages(park.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get park images: %w", err)
		}
		park.Images = images

		parks = append(parks, park)
	}

	return parks, nil
}

// Helper functions for safe type assertions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

// GetParkIDByCode retrieves park ID by park code
func (db *DB) GetParkIDByCode(parkCode string) (int, error) {
	var parkID int
	query := "SELECT id FROM parks WHERE park_code = ?"
	err := db.QueryRow(query, parkCode).Scan(&parkID)
	if err != nil {
		return 0, fmt.Errorf("failed to get park ID for code %s: %w", parkCode, err)
	}
	return parkID, nil
}

// UpsertParkData inserts or updates cached park data
func (db *DB) UpsertParkData(parkID int, dataType, tableName string, data interface{}) error {
	apiDataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// SQLite UPSERT syntax using INSERT OR REPLACE
	query := fmt.Sprintf(`
		INSERT OR REPLACE INTO %s (park_id, data_type, api_data, updated_at, last_fetched_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, tableName)

	_, err = db.Exec(query, parkID, dataType, string(apiDataJSON))
	if err != nil {
		return fmt.Errorf("failed to upsert park data: %w", err)
	}
	return nil
}

// GetCachedParkData retrieves cached park data
func (db *DB) GetCachedParkData(parkID int, dataType, tableName string) (*CachedParkData, error) {
	query := fmt.Sprintf(`
		SELECT id, park_id, data_type, api_data, created_at, updated_at, last_fetched_at
		FROM %s WHERE park_id = ? AND data_type = ?
	`, tableName)

	var data CachedParkData
	row := db.QueryRow(query, parkID, dataType)
	err := row.Scan(
		&data.ID, &data.ParkID, &data.DataType, &data.APIData,
		&data.CreatedAt, &data.UpdatedAt, &data.LastFetchedAt,
	)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// IsParkDataStale checks if specific park data needs refreshing
func (db *DB) IsParkDataStale(parkID int, dataType, tableName string, maxAge time.Duration) (bool, error) {
	query := fmt.Sprintf(`
		SELECT last_fetched_at FROM %s 
	`, tableName)
	if parkID > 0 && dataType != "" {
		query += "WHERE park_id = ? AND data_type = ?"
	} else if parkID > 0 {
		query += "WHERE park_id = ?"
	} else if dataType != "" {
		query += "WHERE data_type = ?"
	}

	var lastFetched time.Time
	err := db.QueryRow(query, parkID, dataType).Scan(&lastFetched)
	if err != nil {
		// If no data exists, it's stale
		return true, nil
	}

	return time.Since(lastFetched) > maxAge, nil
}

type CachedGalleryAsset struct {
	ID            int       `json:"id"`
	ParkID        int       `json:"park_id"`
	GalleryID     string    `json:"gallery_id"`
	AssetID       string    `json:"asset_id"`
	Title         string    `json:"title"`
	AltText       string    `json:"alt_text"`
	Caption       string    `json:"caption"`
	Credit        string    `json:"credit"`
	URL           string    `json:"url"`
	AssetType     string    `json:"asset_type"`
	FileSize      int       `json:"file_size"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	APIData       string    `json:"api_data"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	LastFetchedAt time.Time `json:"last_fetched_at"`
}

func (db *DB) IsGalleryAssetsStale(parkID int, galleryID string, maxAge time.Duration) (bool, error) {
	var lastFetched time.Time
	query := `SELECT last_fetched_at FROM park_gallery_assets 
              WHERE park_id = ? AND gallery_id = ? 
              ORDER BY last_fetched_at DESC LIMIT 1`

	err := db.QueryRow(query, parkID, galleryID).Scan(&lastFetched)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil // No cached data, so it's stale
		}
		return true, err
	}

	return time.Since(lastFetched) > maxAge, nil
}

func (db *DB) GetCachedGalleryAssets(parkID int, galleryID string) (*CachedGalleryAsset, error) {
	var asset CachedGalleryAsset
	query := `SELECT api_data, last_fetched_at FROM park_gallery_assets 
              WHERE park_id = ? AND gallery_id = ? 
              ORDER BY last_fetched_at DESC LIMIT 1`

	err := db.QueryRow(query, parkID, galleryID).Scan(&asset.APIData, &asset.LastFetchedAt)
	return &asset, err
}

func (db *DB) UpsertGalleryAssets(parkID int, galleryID string, response *nps.MultimediaGalleriesAssetsResponse) error {
	// Marshal the entire response to JSON
	apiDataJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal gallery assets: %w", err)
	}
	query := `
        INSERT OR REPLACE INTO park_gallery_assets (
            park_id, gallery_id, api_data, updated_at, last_fetched_at
        ) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `

	_, err = db.Exec(query, parkID, galleryID, string(apiDataJSON))
	if err != nil {
		return fmt.Errorf("failed to upsert gallery assets: %w", err)
	}
	return nil
}
