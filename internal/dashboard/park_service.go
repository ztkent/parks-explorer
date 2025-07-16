package dashboard

import (
	"fmt"
	"time"

	"github.com/ztkent/go-nps"
	"github.com/ztkent/nps-dashboard/internal/database"
)

// GetFeaturedParks returns the first 12 parks for the featured section
func (ps *ParkService) GetFeaturedParks() ([]database.CachedPark, error) {
	// Use pagination to get exactly 12 featured parks
	return ps.GetParksWithPagination(0, 12)
}

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

// GetParksWithPagination returns parks with pagination support
func (ps *ParkService) GetParksWithPagination(offset, limit int) ([]database.CachedPark, error) {
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

	return ps.db.GetParksWithPagination(offset, limit)
}

// GetTotalParksCount returns the total number of parks
func (ps *ParkService) GetTotalParksCount() (int, error) {
	return ps.db.GetTotalParksCount()
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

// GetParkThingsToDo fetches things to do for a specific park
func (ps *ParkService) GetParkThingsToDo(parkCode string) (*nps.ThingsToDoResponse, error) {
	return ps.npsApi.GetThingsToDo("", parkCode, "", "", 20, 0, nil)
}

// GetParkTours fetches tours for a specific park
func (ps *ParkService) GetParkTours(parkCode string) (*nps.TourResponse, error) {
	return ps.npsApi.GetTours(nil, []string{parkCode}, nil, "", 20, 0, nil)
}

// GetParkActivities fetches activities available for a specific park
func (ps *ParkService) GetParkActivities(parkCode string) (*nps.ActivityResponse, error) {
	return ps.npsApi.GetActivities("", "", 20, 0, "")
}

// GetParkMedia fetches multimedia content for a specific park
func (ps *ParkService) GetParkMedia(parkCode string) (map[string]interface{}, error) {
	// Fetch galleries
	galleries, err := ps.npsApi.GetMultimediaGalleries([]string{parkCode}, nil, "", 0, 10)
	if err != nil {
		galleries = nil // Continue even if galleries fail
	}

	// Fetch videos
	videos, err := ps.npsApi.GetMultimediaVideos([]string{parkCode}, nil, "", 0, 10)
	if err != nil {
		videos = nil // Continue even if videos fail
	}

	// Fetch webcams
	webcams, err := ps.npsApi.GetWebcams("", []string{parkCode}, nil, "", 10, 0)
	if err != nil {
		webcams = nil // Continue even if webcams fail
	}

	return map[string]interface{}{
		"galleries": galleries,
		"videos":    videos,
		"webcams":   webcams,
	}, nil
}

// GetParkNews fetches news and alerts for a specific park
func (ps *ParkService) GetParkNews(parkCode string) (map[string]interface{}, error) {
	// Fetch articles
	articles, err := ps.npsApi.GetArticles([]string{parkCode}, nil, "", 10, 0)
	if err != nil {
		articles = nil // Continue even if articles fail
	}

	// Fetch alerts
	alerts, err := ps.npsApi.GetAlerts([]string{parkCode}, nil, "", 10, 0)
	if err != nil {
		alerts = nil // Continue even if alerts fail
	}

	// Fetch events
	events, err := ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 10, 0, false)
	if err != nil {
		events = nil // Continue even if events fail
	}

	return map[string]interface{}{
		"articles": articles,
		"alerts":   alerts,
		"events":   events,
	}, nil
}

// GetParkEnhancedDetails fetches comprehensive park details
func (ps *ParkService) GetParkEnhancedDetails(parkCode string) (map[string]interface{}, error) {
	// Get basic park info
	parks, err := ps.npsApi.GetParks([]string{parkCode}, nil, 0, 1, "", nil)
	if err != nil || len(parks.Data) == 0 {
		return nil, fmt.Errorf("park not found")
	}
	park := parks.Data[0]

	// Fetch visitor centers
	visitorCenters, err := ps.npsApi.GetVisitorCenters([]string{parkCode}, nil, "", 10, 0, nil)
	if err != nil {
		visitorCenters = nil
	}

	// Fetch campgrounds
	campgrounds, err := ps.npsApi.GetCampgrounds([]string{parkCode}, nil, "", 10, 0, nil)
	if err != nil {
		campgrounds = nil
	}

	// Fetch fees and passes
	fees, err := ps.npsApi.GetFeesPasses([]string{parkCode}, nil, "", 0, 10, nil)
	if err != nil {
		fees = nil
	}

	// Fetch parking lots
	parkingLots, err := ps.npsApi.GetParkinglots([]string{parkCode}, nil, "", 0, 10)
	if err != nil {
		parkingLots = nil
	}

	return map[string]interface{}{
		"park":           park,
		"visitorCenters": visitorCenters,
		"campgrounds":    campgrounds,
		"fees":           fees,
		"parkingLots":    parkingLots,
	}, nil
}

// GetParkAmenities fetches amenities for a specific park
func (ps *ParkService) GetParkAmenities(parkCode string) (*nps.AmenityResponse, error) {
	return ps.npsApi.GetAmenities([]string{parkCode}, "", 20, 0)
}

// GetParkNewsReleases fetches news releases for a specific park
func (ps *ParkService) GetParkNewsReleases(parkCode string) (*nps.NewsReleaseResponse, error) {
	return ps.npsApi.GetNewsReleases([]string{parkCode}, nil, "", 10, 0, nil)
}

// GetParkMultimediaAudio fetches audio content for a specific park
func (ps *ParkService) GetParkMultimediaAudio(parkCode string) (*nps.MultimediaAudioResponse, error) {
	return ps.npsApi.GetMultimediaAudio([]string{parkCode}, nil, "", 0, 10)
}

// GetParkMultimediaGalleriesAssets fetches assets for a specific gallery
func (ps *ParkService) GetParkMultimediaGalleriesAssets(galleryId string, parkCode string) (*nps.MultimediaGalleriesAssetsResponse, error) {
	return ps.npsApi.GetMultimediaGalleriesAssets("", galleryId, []string{parkCode}, nil, "", 0, 50)
}

// GetParkEvents fetches events for a specific park
func (ps *ParkService) GetParkEvents(parkCode string) (*nps.EventResponse, error) {
	return ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 10, 0, false)
}

// GetParkCampgrounds fetches campgrounds for a specific park
func (ps *ParkService) GetParkCampgrounds(parkCode string) (*nps.CampgroundData, error) {
	return ps.npsApi.GetCampgrounds([]string{parkCode}, nil, "", 10, 0, nil)
}

// GetParkOverview fetches comprehensive overview data with all available fields
func (ps *ParkService) GetParkOverview(parkCode string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Fetch things to do for overview with error handling
	thingsToDo, err := ps.npsApi.GetThingsToDo("", parkCode, "", "", 6, 0, nil)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch things to do for park %s: %v\n", parkCode, err)
		result["thingsToDo"] = map[string]interface{}{
			"Data":  []interface{}{},
			"Total": "0",
			"Limit": "6",
			"Start": "0",
		}
	} else {
		result["thingsToDo"] = thingsToDo
	}

	// Fetch activities with error handling
	activities, err := ps.npsApi.GetActivities("", "", 15, 0, "")
	if err != nil {
		fmt.Printf("Warning: Failed to fetch activities for park %s: %v\n", parkCode, err)
		result["activities"] = map[string]interface{}{
			"Data":  []interface{}{},
			"Total": "0",
			"Limit": "15",
			"Start": "0",
		}
	} else {
		result["activities"] = activities
	}

	// Fetch visitor centers for services overview with error handling
	visitorCenters, err := ps.npsApi.GetVisitorCenters([]string{parkCode}, nil, "", 3, 0, nil)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch visitor centers for park %s: %v\n", parkCode, err)
		result["visitorCenters"] = map[string]interface{}{
			"Data":  []interface{}{},
			"Total": "0",
			"Limit": "3",
			"Start": "0",
		}
	} else {
		result["visitorCenters"] = visitorCenters
	}

	// Fetch amenities for facilities overview with error handling
	amenities, err := ps.npsApi.GetAmenities([]string{parkCode}, "", 10, 0)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch amenities for park %s: %v\n", parkCode, err)
		result["amenities"] = map[string]interface{}{
			"Data":  []interface{}{},
			"Total": "0",
			"Limit": "10",
			"Start": "0",
		}
	} else {
		result["amenities"] = amenities
	}

	// Fetch tours for enhanced overview with error handling
	tours, err := ps.npsApi.GetTours(nil, []string{parkCode}, nil, "", 3, 0, nil)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch tours for park %s: %v\n", parkCode, err)
		result["tours"] = map[string]interface{}{
			"Data":  []interface{}{},
			"Total": "0",
			"Limit": "3",
			"Start": "0",
		}
	} else {
		result["tours"] = tours
	}

	// Fetch events for overview with error handling
	events, err := ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 3, 0, false)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch events for park %s: %v\n", parkCode, err)
		result["events"] = map[string]interface{}{
			"Data":  []interface{}{},
			"Total": "0",
			"Limit": "3",
			"Start": "0",
		}
	} else {
		result["events"] = events
	}

	// Fetch alerts for overview with error handling
	alerts, err := ps.npsApi.GetAlerts([]string{parkCode}, nil, "", 3, 0)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch alerts for park %s: %v\n", parkCode, err)
		result["alerts"] = map[string]interface{}{
			"Data":  []interface{}{},
			"Total": "0",
			"Limit": "3",
			"Start": "0",
		}
	} else {
		result["alerts"] = alerts
	}

	return result, nil
}

// UpdateGetParkMedia to include audio content
func (ps *ParkService) GetParkMediaComplete(parkCode string) (map[string]interface{}, error) {
	// Fetch galleries
	galleries, err := ps.npsApi.GetMultimediaGalleries([]string{parkCode}, nil, "", 0, 10)
	if err != nil {
		galleries = nil
	}

	// Fetch gallery assets for each gallery if galleries were successfully retrieved
	var galleryAssets map[string]interface{}
	if galleries != nil && len(galleries.Data) > 0 {
		galleryAssets = make(map[string]interface{})
		for _, gallery := range galleries.Data {
			// Only fetch assets if the gallery has assets to avoid unnecessary API calls
			if gallery.AssetCount > 0 {
				assets, err := ps.npsApi.GetMultimediaGalleriesAssets("", gallery.ID, []string{parkCode}, nil, "", 0, 50)
				if err != nil {
					fmt.Printf("Warning: Failed to fetch assets for gallery %s: %v\n", gallery.ID, err)
					// Still include an empty assets response to maintain data structure
					galleryAssets[gallery.ID] = map[string]interface{}{
						"Data":  []interface{}{},
						"Total": "0",
						"Limit": "50",
						"Start": "0",
					}
				} else {
					galleryAssets[gallery.ID] = assets
				}
			}
		}
	}

	// Fetch videos
	videos, err := ps.npsApi.GetMultimediaVideos([]string{parkCode}, nil, "", 0, 10)
	if err != nil {
		videos = nil
	}

	// Fetch audio content
	audio, err := ps.npsApi.GetMultimediaAudio([]string{parkCode}, nil, "", 0, 10)
	if err != nil {
		audio = nil
	}

	// Fetch webcams
	webcams, err := ps.npsApi.GetWebcams("", []string{parkCode}, nil, "", 10, 0)
	if err != nil {
		webcams = nil
	}

	return map[string]interface{}{
		"galleries":     galleries,
		"galleryAssets": galleryAssets,
		"videos":        videos,
		"audio":         audio,
		"webcams":       webcams,
	}, nil
}

// UpdateGetParkNews to include news releases
func (ps *ParkService) GetParkNewsComplete(parkCode string) (map[string]interface{}, error) {
	// Fetch news releases
	newsReleases, err := ps.npsApi.GetNewsReleases([]string{parkCode}, nil, "", 10, 0, nil)
	if err != nil {
		newsReleases = nil
	}

	// Fetch articles
	articles, err := ps.npsApi.GetArticles([]string{parkCode}, nil, "", 10, 0)
	if err != nil {
		articles = nil
	}

	// Fetch alerts
	alerts, err := ps.npsApi.GetAlerts([]string{parkCode}, nil, "", 10, 0)
	if err != nil {
		alerts = nil
	}

	// Fetch events
	events, err := ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 10, 0, false)
	if err != nil {
		events = nil
	}

	return map[string]interface{}{
		"newsReleases": newsReleases,
		"articles":     articles,
		"alerts":       alerts,
		"events":       events,
	}, nil
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
