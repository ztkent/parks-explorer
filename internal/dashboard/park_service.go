package dashboard

import (
	"fmt"
	"time"

	"github.com/ztkent/go-nps"
	"github.com/ztkent/nps-dashboard/internal/database"
)

// ParkService handles park data with caching
type ParkService struct {
	npsApi nps.NpsApi
	db     *database.DB
}

// NewParkService creates a new park service
func NewParkService(npsApi nps.NpsApi, db *database.DB) *ParkService {
	return &ParkService{
		npsApi: npsApi,
		db:     db,
	}
}

// GetAllParks returns all parks, using cache when possible
func (ps *ParkService) GetAllParks() ([]database.CachedPark, error) {
	// Check if we need to refresh data (older than 24 hours)
	stale, err := ps.db.IsDataStale(24 * time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to check data staleness: %w", err)
	}

	if stale {
		if err := ps.RefreshParksFromAPI(); err != nil {
			// If API fails, try to return cached data anyway
			fmt.Printf("Warning: Failed to refresh parks from API: %v\n", err)
		}
	}

	return ps.db.GetAllParks()
}

// GetFeaturedParks returns the first 6 parks for the featured section
func (ps *ParkService) GetFeaturedParks() ([]database.CachedPark, error) {
	allParks, err := ps.GetAllParks()
	if err != nil {
		return nil, err
	}

	// Return first 6 parks
	if len(allParks) > 6 {
		return allParks[:6], nil
	}
	return allParks, nil
}

// SearchParks searches for parks by name
func (ps *ParkService) SearchParks(query string) ([]database.CachedPark, error) {
	// Always try cache first for search
	results, err := ps.db.SearchParks(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search cached parks: %w", err)
	}

	// If we have results, return them
	if len(results) > 0 {
		return results, nil
	}

	// If no cached results and data is stale, refresh and try again
	stale, err := ps.db.IsDataStale(24 * time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to check data staleness: %w", err)
	}

	if stale {
		if err := ps.RefreshParksFromAPI(); err != nil {
			return nil, fmt.Errorf("failed to refresh parks from API: %w", err)
		}
		// Try search again after refresh
		return ps.db.SearchParks(query)
	}

	return results, nil
}

// GetParkBySlug returns a specific park by slug
func (ps *ParkService) GetParkBySlug(slug string) (*database.CachedPark, error) {
	// Try cache first
	park, err := ps.db.GetParkBySlug(slug)
	if err == nil {
		return park, nil
	}

	// If not found and data is stale, refresh and try again
	stale, err := ps.db.IsDataStale(24 * time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to check data staleness: %w", err)
	}

	if stale {
		if err := ps.RefreshParksFromAPI(); err != nil {
			return nil, fmt.Errorf("failed to refresh parks from API: %w", err)
		}
		return ps.db.GetParkBySlug(slug)
	}

	return nil, fmt.Errorf("park not found: %s", slug)
}

// RefreshParksFromAPI fetches fresh data from the NPS API and caches it
func (ps *ParkService) RefreshParksFromAPI() error {
	fmt.Println("Refreshing parks data from NPS API...")

	// Fetch parks from API
	res, err := ps.npsApi.GetParks(nil, nil, 0, 500, "", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch parks from API: %w", err)
	}

	fmt.Printf("Fetched %d parks from API\n", len(res.Data))

	// Cache each park
	for _, park := range res.Data {
		slug := createSlug(park.Name)

		// Convert park to map for easier handling
		parkMap := map[string]interface{}{
			"parkCode":       park.ParkCode,
			"name":           park.Name,
			"fullName":       park.FullName,
			"states":         park.States,
			"designation":    park.Designation,
			"description":    park.Description,
			"weatherInfo":    park.WeatherInfo,
			"directionsInfo": park.DirectionsInfo,
			"url":            park.Url,
			"directionsUrl":  park.DirectionsUrl,
			"latitude":       park.Latitude,
			"longitude":      park.Longitude,
			"latLong":        park.LatLong,
			"relevanceScore": park.RelevanceScore,
			"images":         convertImagesToInterface(park.Images),
		}

		_, err := ps.db.UpsertPark(parkMap, slug)
		if err != nil {
			fmt.Printf("Warning: Failed to cache park %s: %v\n", park.Name, err)
			continue
		}
	}

	fmt.Printf("Successfully cached %d parks\n", len(res.Data))
	return nil
}

// convertImagesToInterface converts the park images to interface{} for database storage
func convertImagesToInterface(images []struct {
	Credit  string `json:"credit"`
	AltText string `json:"altText"`
	Title   string `json:"title"`
	Caption string `json:"caption"`
	Url     string `json:"url"`
}) []interface{} {
	result := make([]interface{}, len(images))
	for i, img := range images {
		result[i] = map[string]interface{}{
			"credit":  img.Credit,
			"altText": img.AltText,
			"title":   img.Title,
			"caption": img.Caption,
			"url":     img.Url,
		}
	}
	return result
}
