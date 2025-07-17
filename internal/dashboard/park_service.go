package dashboard

import (
	"encoding/json"
	"fmt"
	"strings"
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

// GetEventsWithFilters fetches events with filtering options
func (ps *ParkService) GetEventsWithFilters(parkCodes []string, stateCodes []string, dateStart, dateEnd string, eventTypes []string, query string, pageSize, pageNumber int) (*nps.EventResponse, error) {
	// Ensure we always use today as the minimum start date to avoid showing past events
	today := time.Now().Format("2006-01-02")
	if dateStart == "" || dateStart < today {
		dateStart = today
	}

	return ps.npsApi.GetEvents(parkCodes, stateCodes, nil, nil, nil, nil, nil, nil, dateStart, dateEnd, eventTypes, "", query, pageSize, pageNumber, false)
}

// GetAllParks returns all parks, using cache when possible
func (ps *ParkService) GetAllParks() ([]database.CachedPark, error) {
	// First try to get cached parks
	parks, err := ps.db.GetAllParks()
	if err != nil || len(parks) == 0 {
		// If no cached parks or error, fetch from API
		res, err := ps.npsApi.GetParks(nil, nil, 0, 500, "", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch parks from API: %w", err)
		}

		// Cache the response
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

		// Get the cached parks after insertion
		parks, err = ps.db.GetAllParks()
		if err != nil {
			return nil, fmt.Errorf("failed to get cached parks after API fetch: %w", err)
		}
	}

	return parks, nil
}

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
	return results, nil
}

// GetParkBySlug returns a specific park by slug
func (ps *ParkService) GetParkBySlug(slug string) (*database.CachedPark, error) {
	park, err := ps.db.GetParkBySlug(slug)
	if err == nil {
		return park, nil
	}
	return nil, fmt.Errorf("park not found: %s", slug)
}

// GetParksWithPagination returns parks with pagination support
func (ps *ParkService) GetParksWithPagination(offset, limit int) ([]database.CachedPark, error) {
	return ps.db.GetParksWithPagination(offset, limit)
}

func (ps *ParkService) GetParkArticles(parkCode string) (*nps.ArticleData, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetArticles([]string{parkCode}, nil, "", 10, 0)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "articles", "park_news", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetArticles([]string{parkCode}, nil, "", 10, 0)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "articles", "park_news", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "articles", "park_news")
	if err != nil {
		return ps.npsApi.GetArticles([]string{parkCode}, nil, "", 10, 0)
	}

	var response nps.ArticleData
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetArticles([]string{parkCode}, nil, "", 10, 0)
	}

	return &response, nil
}

// GetParkAlerts fetches alerts for a specific park, preferring cache
func (ps *ParkService) GetParkAlerts(parkCode string) (*nps.AlertResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetAlerts([]string{parkCode}, nil, "", 10, 0)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "alerts", "park_news", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetAlerts([]string{parkCode}, nil, "", 10, 0)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "alerts", "park_news", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "alerts", "park_news")
	if err != nil {
		return ps.npsApi.GetAlerts([]string{parkCode}, nil, "", 10, 0)
	}

	var response nps.AlertResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetAlerts([]string{parkCode}, nil, "", 10, 0)
	}

	return &response, nil
}

// GetParkEventsList fetches events for a specific park, preferring cache
func (ps *ParkService) GetParkEventsList(parkCode string) (*nps.EventResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	today := time.Now().Format("2006-01-02")

	if err != nil {
		return ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, today, "", nil, "", "", 10, 0, false)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "events", "park_news", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, today, "", nil, "", "", 10, 0, false)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "events", "park_news", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "events", "park_news")
	if err != nil {
		return ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, today, "", nil, "", "", 10, 0, false)
	}

	var response nps.EventResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, today, "", nil, "", "", 10, 0, false)
	}

	return &response, nil
}

func (ps *ParkService) GetParkVisitorCenters(parkCode string) (*nps.VisitorCenterResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetVisitorCenters([]string{parkCode}, nil, "", 20, 0, nil)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "visitor_centers", "park_details", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetVisitorCenters([]string{parkCode}, nil, "", 20, 0, nil)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "visitor_centers", "park_details", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "visitor_centers", "park_details")
	if err != nil {
		return ps.npsApi.GetVisitorCenters([]string{parkCode}, nil, "", 20, 0, nil)
	}

	var response nps.VisitorCenterResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetVisitorCenters([]string{parkCode}, nil, "", 20, 0, nil)
	}

	return &response, nil
}

// GetParkFees fetches fees and passes for a specific park, preferring cache
func (ps *ParkService) GetParkFees(parkCode string) (*nps.FeePassResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetFeesPasses([]string{parkCode}, nil, "", 0, 20, nil)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "fees", "park_details", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetFeesPasses([]string{parkCode}, nil, "", 0, 20, nil)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "fees", "park_details", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "fees", "park_details")
	if err != nil {
		return ps.npsApi.GetFeesPasses([]string{parkCode}, nil, "", 0, 20, nil)
	}

	var response nps.FeePassResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetFeesPasses([]string{parkCode}, nil, "", 0, 20, nil)
	}

	return &response, nil
}

// GetParkParking fetches parking lots for a specific park, preferring cache
func (ps *ParkService) GetParkParking(parkCode string) (*nps.ParkinglotResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetParkinglots([]string{parkCode}, nil, "", 0, 20)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "parking_lots", "park_details", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetParkinglots([]string{parkCode}, nil, "", 0, 20)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "parking_lots", "park_details", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "parking_lots", "park_details")
	if err != nil {
		return ps.npsApi.GetParkinglots([]string{parkCode}, nil, "", 0, 20)
	}

	var response nps.ParkinglotResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetParkinglots([]string{parkCode}, nil, "", 0, 20)
	}

	return &response, nil
}

// GetParkThingsToDo fetches things to do for a specific park, preferring cache
func (ps *ParkService) GetParkThingsToDo(parkCode string) (*nps.ThingsToDoResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetThingsToDo("", parkCode, "", "", 20, 0, nil)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "things_to_do", "park_activities", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetThingsToDo("", parkCode, "", "", 20, 0, nil)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "things_to_do", "park_activities", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "things_to_do", "park_activities")
	if err != nil {
		return ps.npsApi.GetThingsToDo("", parkCode, "", "", 20, 0, nil)
	}

	var response nps.ThingsToDoResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetThingsToDo("", parkCode, "", "", 20, 0, nil)
	}

	return &response, nil
}

// GetParkTours fetches tours for a specific park, preferring cache
func (ps *ParkService) GetParkTours(parkCode string) (*nps.TourResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetTours(nil, []string{parkCode}, nil, "", 20, 0, nil)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "tours", "park_activities", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetTours(nil, []string{parkCode}, nil, "", 20, 0, nil)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "tours", "park_activities", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "tours", "park_activities")
	if err != nil {
		return ps.npsApi.GetTours(nil, []string{parkCode}, nil, "", 20, 0, nil)
	}

	var response nps.TourResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetTours(nil, []string{parkCode}, nil, "", 20, 0, nil)
	}

	return &response, nil
}

// GetParkActivities fetches activities available for a specific park, preferring cache
func (ps *ParkService) GetParkActivities(parkCode string) (*nps.ActivityResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetActivities("", "", 20, 0, "")
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "activities", "park_activities", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetActivities("", "", 20, 0, "")
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "activities", "park_activities", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "activities", "park_activities")
	if err != nil {
		return ps.npsApi.GetActivities("", "", 20, 0, "")
	}

	var response nps.ActivityResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetActivities("", "", 20, 0, "")
	}

	return &response, nil
}

// GetParkAmenities fetches amenities for a specific park, preferring cache
func (ps *ParkService) GetParkAmenities(parkCode string) (*nps.AmenityResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetAmenities([]string{parkCode}, "", 20, 0)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "amenities", "park_details", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetAmenities([]string{parkCode}, "", 20, 0)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "amenities", "park_details", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "amenities", "park_details")
	if err != nil {
		return ps.npsApi.GetAmenities([]string{parkCode}, "", 20, 0)
	}

	var response nps.AmenityResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetAmenities([]string{parkCode}, "", 20, 0)
	}

	return &response, nil
}

// GetParkNewsReleases fetches news releases for a specific park, preferring cache
func (ps *ParkService) GetParkNewsReleases(parkCode string) (*nps.NewsReleaseResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetNewsReleases([]string{parkCode}, nil, "", 10, 0, nil)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "news_releases", "park_news", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetNewsReleases([]string{parkCode}, nil, "", 10, 0, nil)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "news_releases", "park_news", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "news_releases", "park_news")
	if err != nil {
		return ps.npsApi.GetNewsReleases([]string{parkCode}, nil, "", 10, 0, nil)
	}

	var response nps.NewsReleaseResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetNewsReleases([]string{parkCode}, nil, "", 10, 0, nil)
	}

	return &response, nil
}

// GetParkMultimediaAudio fetches audio content for a specific park, preferring cache
func (ps *ParkService) GetParkMultimediaAudio(parkCode string) (*nps.MultimediaAudioResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetMultimediaAudio([]string{parkCode}, nil, "", 0, 10)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "audio", "park_media", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetMultimediaAudio([]string{parkCode}, nil, "", 0, 10)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "audio", "park_media", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "audio", "park_media")
	if err != nil {
		return ps.npsApi.GetMultimediaAudio([]string{parkCode}, nil, "", 0, 10)
	}

	var response nps.MultimediaAudioResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetMultimediaAudio([]string{parkCode}, nil, "", 0, 10)
	}

	return &response, nil
}

// GetParkMultimediaGalleries fetches multimedia galleries for a specific park, preferring cache
func (ps *ParkService) GetParkMultimediaGalleries(parkCode string) (*nps.MultimediaGalleriesResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetMultimediaGalleries([]string{parkCode}, nil, "", 0, 10)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "galleries", "park_media", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetMultimediaGalleries([]string{parkCode}, nil, "", 0, 10)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "galleries", "park_media", response)
		}
		// For each gallery, fetch assets if available
		for i, gallery := range response.Data {
			if gallery.ID != "" {
				assets, err := ps.getParkMultimediaGalleriesAssets(gallery.ID, parkCode)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch assets for gallery %s: %w", gallery.ID, err)
				}
				for _, asset := range assets.Data {
					response.Data[i].Images = append(response.Data[i].Images, struct {
						Url         string "json:\"url\""
						AltText     string "json:\"altText\""
						Title       string "json:\"title\""
						Description string "json:\"description\""
					}{
						Url:         asset.FileInfo.Url,
						AltText:     asset.AltText,
						Title:       asset.Title,
						Description: asset.Description,
					})
				}
			}
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "galleries", "park_media")
	if err != nil {
		return ps.npsApi.GetMultimediaGalleries([]string{parkCode}, nil, "", 0, 10)
	}

	var response nps.MultimediaGalleriesResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetMultimediaGalleries([]string{parkCode}, nil, "", 0, 10)
	}

	// For each gallery, fetch assets if available
	for i, gallery := range response.Data {
		if gallery.ID != "" {
			assets, err := ps.getParkMultimediaGalleriesAssets(gallery.ID, parkCode)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch assets for gallery %s: %w", gallery.ID, err)
			}
			for _, asset := range assets.Data {
				response.Data[i].Images = append(response.Data[i].Images, struct {
					Url         string "json:\"url\""
					AltText     string "json:\"altText\""
					Title       string "json:\"title\""
					Description string "json:\"description\""
				}{
					Url:         asset.FileInfo.Url,
					AltText:     asset.AltText,
					Title:       asset.Title,
					Description: asset.Description,
				})
			}
		}
	}

	return &response, nil
}

// getParkMultimediaGalleriesAssets fetches assets for a specific gallery, preferring cache
func (ps *ParkService) getParkMultimediaGalleriesAssets(galleryId string, parkCode string) (*nps.MultimediaGalleriesAssetsResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetMultimediaGalleriesAssets("", galleryId, []string{parkCode}, nil, "", 0, 500)
	}

	// Check if cached data is fresh using the dedicated gallery assets table
	stale, err := ps.db.IsGalleryAssetsStale(parkID, galleryId, 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetMultimediaGalleriesAssets("", galleryId, []string{parkCode}, nil, "", 0, 500)
		if err == nil {
			ps.db.UpsertGalleryAssets(parkID, galleryId, response)
		}
		return response, err
	}

	// Return cached data from the dedicated table
	cachedData, err := ps.db.GetCachedGalleryAssets(parkID, galleryId)
	if err != nil {
		return ps.npsApi.GetMultimediaGalleriesAssets("", galleryId, []string{parkCode}, nil, "", 0, 500)
	}

	var response nps.MultimediaGalleriesAssetsResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetMultimediaGalleriesAssets("", galleryId, []string{parkCode}, nil, "", 0, 500)
	}

	return &response, nil
}

// GetParkMultimediaVideos fetches multimedia videos for a specific park, preferring cache
func (ps *ParkService) GetParkMultimediaVideos(parkCode string) (*nps.MultimediaVideosResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetMultimediaVideos([]string{parkCode}, nil, "", 0, 10)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "videos", "park_media", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetMultimediaVideos([]string{parkCode}, nil, "", 0, 10)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "videos", "park_media", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "videos", "park_media")
	if err != nil {
		return ps.npsApi.GetMultimediaVideos([]string{parkCode}, nil, "", 0, 10)
	}

	var response nps.MultimediaVideosResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetMultimediaVideos([]string{parkCode}, nil, "", 0, 10)
	}

	return &response, nil
}

// GetParkWebcams fetches webcams for a specific park, preferring cache
func (ps *ParkService) GetParkWebcams(parkCode string) (*nps.WebcamResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetWebcams("", []string{parkCode}, nil, "", 10, 0)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "webcams", "park_media", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetWebcams("", []string{parkCode}, nil, "", 10, 0)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "webcams", "park_media", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "webcams", "park_media")
	if err != nil {
		return ps.npsApi.GetWebcams("", []string{parkCode}, nil, "", 10, 0)
	}

	var response nps.WebcamResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetWebcams("", []string{parkCode}, nil, "", 10, 0)
	}

	return &response, nil
}

// GetParkEvents fetches events for a specific park, preferring cache
func (ps *ParkService) GetParkEvents(parkCode string) (*nps.EventResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	today := time.Now().Format("2006-01-02")

	if err != nil {
		return ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, today, "", nil, "", "", 10, 0, false)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "events", "park_news", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, today, "", nil, "", "", 10, 0, false)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "events", "park_news", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "events", "park_news")
	if err != nil {
		return ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, today, "", nil, "", "", 10, 0, false)
	}

	var response nps.EventResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, today, "", nil, "", "", 10, 0, false)
	}

	return &response, nil
}

// GetParkCampgrounds fetches campgrounds for a specific park, preferring cache
func (ps *ParkService) GetParkCampgrounds(parkCode string) (*nps.CampgroundData, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetCampgrounds([]string{parkCode}, nil, "", 10, 0, nil)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "campgrounds", "park_details", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetCampgrounds([]string{parkCode}, nil, "", 10, 0, nil)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "campgrounds", "park_details", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "campgrounds", "park_details")
	if err != nil {
		return ps.npsApi.GetCampgrounds([]string{parkCode}, nil, "", 10, 0, nil)
	}

	var response nps.CampgroundData
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetCampgrounds([]string{parkCode}, nil, "", 10, 0, nil)
	}

	return &response, nil
}

// GetAllActivities fetches all available activities from the NPS API
func (ps *ParkService) GetAllActivities() (*nps.ActivityResponse, error) {
	return ps.npsApi.GetActivities("", "", 500, 0, "name")
}

// GetEventsWithDateRange fetches events with optional date filtering
func (ps *ParkService) GetEventsWithDateRange(parkCode, stateCode, dateStart, dateEnd string, pageSize, pageNumber int) (*nps.EventResponse, error) {
	var parkCodes []string
	var stateCodes []string

	if parkCode != "" {
		parkCodes = []string{parkCode}
	}
	if stateCode != "" {
		stateCodes = []string{stateCode}
	}

	// Ensure we always use today as the minimum start date to avoid showing past events
	today := time.Now().Format("2006-01-02")
	if dateStart == "" || dateStart < today {
		dateStart = today
	}

	return ps.npsApi.GetEvents(
		parkCodes,  // parkCode
		stateCodes, // stateCode
		nil,        // organization
		nil,        // subject
		nil,        // portal
		nil,        // tagsAll
		nil,        // tagsOne
		nil,        // tagsNone
		dateStart,  // dateStart
		dateEnd,    // dateEnd
		nil,        // eventType
		"",         // id
		"",         // q
		pageSize,   // pageSize
		pageNumber, // pageNumber
		true,       // expandRecurring
	)
}

// SearchEvents searches for events with filters
func (ps *ParkService) SearchEvents(query, parkCode, stateCode, eventType, dateStart, dateEnd string, limit, start int) (*nps.EventResponse, error) {
	var parkCodes []string
	var stateCodes []string
	var eventTypes []string

	if parkCode != "" {
		parkCodes = []string{parkCode}
	}
	if stateCode != "" {
		stateCodes = []string{stateCode}
	}
	if eventType != "" {
		eventTypes = []string{eventType}
	}

	// Ensure we always use today as the minimum start date to avoid showing past events
	today := time.Now().Format("2006-01-02")
	if dateStart == "" || dateStart < today {
		dateStart = today
	}

	pageSize := limit
	if pageSize == 0 {
		pageSize = 50
	}

	pageNumber := start / pageSize

	return ps.npsApi.GetEvents(
		parkCodes,  // parkCode
		stateCodes, // stateCode
		nil,        // organization
		nil,        // subject
		nil,        // portal
		nil,        // tagsAll
		nil,        // tagsOne
		nil,        // tagsNone
		dateStart,  // dateStart
		dateEnd,    // dateEnd
		eventTypes, // eventType
		"",         // id
		query,      // q
		pageSize,   // pageSize
		pageNumber, // pageNumber
		true,       // expandRecurring
	)
}

// SearchThingsToDo searches for things to do with filters
func (ps *ParkService) SearchThingsToDo(activityId, parkCode, stateCode, query string, limit, start int) (*nps.ThingsToDoResponse, error) {
	// If no specific filters are provided, get general things to do
	if activityId == "" && parkCode == "" && stateCode == "" && query == "" {
		// Get some popular things to do by default
		return ps.npsApi.GetThingsToDo("", "", "", "", limit, start, nil)
	}

	// Get things to do without activity filter first
	// The NPS API doesn't support direct activity filtering, so we'll filter client-side
	response, err := ps.npsApi.GetThingsToDo("", parkCode, stateCode, query, limit*2, start, nil)
	if err != nil {
		return nil, err
	}

	// If no activity filter is specified, return the results as-is
	if activityId == "" {
		return response, nil
	}

	// Filter results by activity ID
	var filteredResults []nps.ThingsToDo
	for _, item := range response.Data {
		for _, activity := range item.Activities {
			if activity.ID == activityId {
				filteredResults = append(filteredResults, item)
				break
			}
		}
	}

	// Limit the results to the requested amount
	if len(filteredResults) > limit {
		filteredResults = filteredResults[:limit]
	}

	// Create a new response with filtered data
	filteredResponse := &nps.ThingsToDoResponse{
		Total: fmt.Sprintf("%d", len(filteredResults)),
		Data:  filteredResults,
		Limit: response.Limit,
		Start: response.Start,
	}

	return filteredResponse, nil
}

// createSlug creates a URL-friendly slug from text
func createSlug(text string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "&", "and")
	// Remove other special characters except hyphens and alphanumeric
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	// Clean up multiple consecutive hyphens
	cleanSlug := result.String()
	for strings.Contains(cleanSlug, "--") {
		cleanSlug = strings.ReplaceAll(cleanSlug, "--", "-")
	}
	// Trim hyphens from start and end
	cleanSlug = strings.Trim(cleanSlug, "-")
	return cleanSlug
}
