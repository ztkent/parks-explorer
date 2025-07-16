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

// GetParkMedia fetches multimedia content for a specific park, preferring cache
func (ps *ParkService) GetParkMedia(parkCode string) (map[string]interface{}, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.fetchMediaFromAPI(parkCode)
	}

	result := make(map[string]interface{})
	mediaTypes := []string{"galleries", "videos", "audio", "webcams"}

	// Check each media type
	for _, mediaType := range mediaTypes {
		stale, err := ps.db.IsParkDataStale(parkID, mediaType, "park_media", 24*time.Hour)
		if err != nil || stale {
			// Fetch fresh data and cache it
			ps.refreshMediaType(parkID, parkCode, mediaType)
		}

		// Get cached data
		cachedData, err := ps.db.GetCachedParkData(parkID, mediaType, "park_media")
		if err == nil {
			// Unmarshal into proper response type based on media type
			switch mediaType {
			case "galleries":
				var galleries nps.MultimediaGalleriesResponse
				if err := json.Unmarshal([]byte(cachedData.APIData), &galleries); err == nil {
					result[mediaType] = galleries
				}
			case "videos":
				var videos nps.MultimediaVideosResponse
				if err := json.Unmarshal([]byte(cachedData.APIData), &videos); err == nil {
					result[mediaType] = videos
				}
			case "audio":
				var audio nps.MultimediaAudioResponse
				if err := json.Unmarshal([]byte(cachedData.APIData), &audio); err == nil {
					result[mediaType] = audio
				}
			case "webcams":
				var webcams nps.WebcamResponse
				if err := json.Unmarshal([]byte(cachedData.APIData), &webcams); err == nil {
					result[mediaType] = webcams
				}
			}
		}
	}

	// Fetch gallery assets if we have galleries
	if galleries, exists := result["galleries"]; exists {
		if galleryResp, ok := galleries.(nps.MultimediaGalleriesResponse); ok && len(galleryResp.Data) > 0 {
			galleryAssets := make(map[string]*nps.MultimediaGalleriesAssetsResponse)
			
			for _, gallery := range galleryResp.Data {
				// Check if gallery assets are cached and fresh
				assetKey := fmt.Sprintf("gallery_assets_%s", gallery.ID)
				stale, err := ps.db.IsParkDataStale(parkID, assetKey, "park_media", 24*time.Hour)
				if err != nil || stale {
					// Fetch fresh gallery assets and cache them
					if assets, err := ps.npsApi.GetMultimediaGalleriesAssets("", gallery.ID, []string{parkCode}, nil, "", 0, 50); err == nil {
						ps.db.UpsertParkData(parkID, assetKey, "park_media", assets)
						galleryAssets[gallery.ID] = assets
					}
				} else {
					// Get cached gallery assets
					if cachedAssets, err := ps.db.GetCachedParkData(parkID, assetKey, "park_media"); err == nil {
						var assets nps.MultimediaGalleriesAssetsResponse
						if err := json.Unmarshal([]byte(cachedAssets.APIData), &assets); err == nil {
							galleryAssets[gallery.ID] = &assets
						}
					}
				}
			}
			
			if len(galleryAssets) > 0 {
				result["galleryAssets"] = galleryAssets
			}
		}
	}

	// If any media type is missing, fetch from API as fallback
	if len(result) == 0 {
		return ps.fetchMediaFromAPI(parkCode)
	}

	return result, nil
}

// GetParkNews fetches news and alerts for a specific park, preferring cache
func (ps *ParkService) GetParkNews(parkCode string) (map[string]interface{}, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.fetchNewsFromAPI(parkCode)
	}

	result := make(map[string]interface{})
	newsTypes := []string{"articles", "alerts", "events", "news_releases"}

	// Check each news type
	for _, newsType := range newsTypes {
		stale, err := ps.db.IsParkDataStale(parkID, newsType, "park_news", 24*time.Hour)
		if err != nil || stale {
			// Fetch fresh data and cache it
			ps.refreshNewsType(parkID, parkCode, newsType)
		}

		// Get cached data
		cachedData, err := ps.db.GetCachedParkData(parkID, newsType, "park_news")
		if err == nil {
			// Unmarshal into proper response type based on news type
			switch newsType {
			case "articles":
				var articles nps.ArticleData
				if err := json.Unmarshal([]byte(cachedData.APIData), &articles); err == nil {
					result[newsType] = articles
				}
			case "alerts":
				var alerts nps.AlertResponse
				if err := json.Unmarshal([]byte(cachedData.APIData), &alerts); err == nil {
					result[newsType] = alerts
				}
			case "events":
				var events nps.EventResponse
				if err := json.Unmarshal([]byte(cachedData.APIData), &events); err == nil {
					result[newsType] = events
				}
			case "news_releases":
				var newsReleases nps.NewsReleaseResponse
				if err := json.Unmarshal([]byte(cachedData.APIData), &newsReleases); err == nil {
					result[newsType] = newsReleases
				}
			}
		}
	}

	// If any news type is missing, fetch from API as fallback
	if len(result) == 0 {
		return ps.fetchNewsFromAPI(parkCode)
	}

	return result, nil
}

// GetParkEnhancedDetails fetches comprehensive park details, preferring cache
func (ps *ParkService) GetParkEnhancedDetails(parkCode string) (map[string]interface{}, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.fetchDetailsFromAPI(parkCode)
	}

	result := make(map[string]interface{})
	detailTypes := []string{"visitor_centers", "campgrounds", "fees", "amenities", "parking_lots"}

	// Check each detail type
	for _, detailType := range detailTypes {
		stale, err := ps.db.IsParkDataStale(parkID, detailType, "park_details", 24*time.Hour)
		if err != nil || stale {
			// Fetch fresh data and cache it
			ps.refreshDetailType(parkID, parkCode, detailType)
		}

		// Get cached data
		cachedData, err := ps.db.GetCachedParkData(parkID, detailType, "park_details")
		if err == nil {
			var detailData interface{}
			if err := json.Unmarshal([]byte(cachedData.APIData), &detailData); err == nil {
				result[detailType] = detailData
			}
		}
	}

	// Get basic park info
	parks, err := ps.npsApi.GetParks([]string{parkCode}, nil, 0, 1, "", nil)
	if err == nil && len(parks.Data) > 0 {
		result["park"] = parks.Data[0]
	}

	// If any detail type is missing, fetch from API as fallback
	if len(result) <= 1 { // Only park data
		return ps.fetchDetailsFromAPI(parkCode)
	}

	return result, nil
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

// GetParkMultimediaGalleriesAssets fetches assets for a specific gallery
func (ps *ParkService) GetParkMultimediaGalleriesAssets(galleryId string, parkCode string) (*nps.MultimediaGalleriesAssetsResponse, error) {
	return ps.npsApi.GetMultimediaGalleriesAssets("", galleryId, []string{parkCode}, nil, "", 0, 50)
}

// GetParkEvents fetches events for a specific park, preferring cache
func (ps *ParkService) GetParkEvents(parkCode string) (*nps.EventResponse, error) {
	// Get park ID first
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err != nil {
		return ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 10, 0, false)
	}

	// Check if cached data is fresh
	stale, err := ps.db.IsParkDataStale(parkID, "events", "park_news", 24*time.Hour)
	if err != nil || stale {
		// Fetch fresh data from API
		response, err := ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 10, 0, false)
		if err == nil {
			// Cache the response
			ps.db.UpsertParkData(parkID, "events", "park_news", response)
		}
		return response, err
	}

	// Return cached data
	cachedData, err := ps.db.GetCachedParkData(parkID, "events", "park_news")
	if err != nil {
		return ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 10, 0, false)
	}

	var response nps.EventResponse
	if err := json.Unmarshal([]byte(cachedData.APIData), &response); err != nil {
		return ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 10, 0, false)
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

// GetParkOverview fetches comprehensive overview data with caching support
func (ps *ParkService) GetParkOverview(parkCode string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Get park ID for caching
	parkID, err := ps.db.GetParkIDByCode(parkCode)
	if err == nil {
		// Fetch things to do for overview with caching
		stale, _ := ps.db.IsParkDataStale(parkID, "things_to_do_overview", "park_activities", 24*time.Hour)
		if stale {
			if thingsToDo, err := ps.npsApi.GetThingsToDo("", parkCode, "", "", 6, 0, nil); err == nil {
				ps.db.UpsertParkData(parkID, "things_to_do_overview", "park_activities", thingsToDo)
				result["thingsToDo"] = thingsToDo
			}
		} else {
			if cachedData, err := ps.db.GetCachedParkData(parkID, "things_to_do_overview", "park_activities"); err == nil {
				var thingsToDo interface{}
				if json.Unmarshal([]byte(cachedData.APIData), &thingsToDo) == nil {
					result["thingsToDo"] = thingsToDo
				}
			}
		}

		// Fetch activities with caching
		stale, _ = ps.db.IsParkDataStale(parkID, "activities_overview", "park_activities", 24*time.Hour)
		if stale {
			if activities, err := ps.npsApi.GetActivities("", "", 15, 0, ""); err == nil {
				ps.db.UpsertParkData(parkID, "activities_overview", "park_activities", activities)
				result["activities"] = activities
			}
		} else {
			if cachedData, err := ps.db.GetCachedParkData(parkID, "activities_overview", "park_activities"); err == nil {
				var activities interface{}
				if json.Unmarshal([]byte(cachedData.APIData), &activities) == nil {
					result["activities"] = activities
				}
			}
		}

		// Fetch visitor centers with caching
		stale, _ = ps.db.IsParkDataStale(parkID, "visitor_centers_overview", "park_details", 24*time.Hour)
		if stale {
			if visitorCenters, err := ps.npsApi.GetVisitorCenters([]string{parkCode}, nil, "", 3, 0, nil); err == nil {
				ps.db.UpsertParkData(parkID, "visitor_centers_overview", "park_details", visitorCenters)
				result["visitorCenters"] = visitorCenters
			}
		} else {
			if cachedData, err := ps.db.GetCachedParkData(parkID, "visitor_centers_overview", "park_details"); err == nil {
				var visitorCenters interface{}
				if json.Unmarshal([]byte(cachedData.APIData), &visitorCenters) == nil {
					result["visitorCenters"] = visitorCenters
				}
			}
		}

		// Fetch amenities with caching
		stale, _ = ps.db.IsParkDataStale(parkID, "amenities_overview", "park_details", 24*time.Hour)
		if stale {
			if amenities, err := ps.npsApi.GetAmenities([]string{parkCode}, "", 10, 0); err == nil {
				ps.db.UpsertParkData(parkID, "amenities_overview", "park_details", amenities)
				result["amenities"] = amenities
			}
		} else {
			if cachedData, err := ps.db.GetCachedParkData(parkID, "amenities_overview", "park_details"); err == nil {
				var amenities interface{}
				if json.Unmarshal([]byte(cachedData.APIData), &amenities) == nil {
					result["amenities"] = amenities
				}
			}
		}

		// Fetch tours with caching
		stale, _ = ps.db.IsParkDataStale(parkID, "tours_overview", "park_activities", 24*time.Hour)
		if stale {
			if tours, err := ps.npsApi.GetTours(nil, []string{parkCode}, nil, "", 3, 0, nil); err == nil {
				ps.db.UpsertParkData(parkID, "tours_overview", "park_activities", tours)
				result["tours"] = tours
			}
		} else {
			if cachedData, err := ps.db.GetCachedParkData(parkID, "tours_overview", "park_activities"); err == nil {
				var tours interface{}
				if json.Unmarshal([]byte(cachedData.APIData), &tours) == nil {
					result["tours"] = tours
				}
			}
		}

		// Fetch events with caching
		stale, _ = ps.db.IsParkDataStale(parkID, "events_overview", "park_news", 24*time.Hour)
		if stale {
			if events, err := ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 3, 0, false); err == nil {
				ps.db.UpsertParkData(parkID, "events_overview", "park_news", events)
				result["events"] = events
			}
		} else {
			if cachedData, err := ps.db.GetCachedParkData(parkID, "events_overview", "park_news"); err == nil {
				var events interface{}
				if json.Unmarshal([]byte(cachedData.APIData), &events) == nil {
					result["events"] = events
				}
			}
		}

		// Fetch alerts with caching
		stale, _ = ps.db.IsParkDataStale(parkID, "alerts_overview", "park_news", 24*time.Hour)
		if stale {
			if alerts, err := ps.npsApi.GetAlerts([]string{parkCode}, nil, "", 3, 0); err == nil {
				ps.db.UpsertParkData(parkID, "alerts_overview", "park_news", alerts)
				result["alerts"] = alerts
			}
		} else {
			if cachedData, err := ps.db.GetCachedParkData(parkID, "alerts_overview", "park_news"); err == nil {
				var alerts interface{}
				if json.Unmarshal([]byte(cachedData.APIData), &alerts) == nil {
					result["alerts"] = alerts
				}
			}
		}
	}

	// Fallback to API if no park ID found or cache failed
	if len(result) == 0 {
		// Fallback to direct API calls with error handling
		if thingsToDo, err := ps.npsApi.GetThingsToDo("", parkCode, "", "", 6, 0, nil); err == nil {
			result["thingsToDo"] = thingsToDo
		}
		if activities, err := ps.npsApi.GetActivities("", "", 15, 0, ""); err == nil {
			result["activities"] = activities
		}
		if visitorCenters, err := ps.npsApi.GetVisitorCenters([]string{parkCode}, nil, "", 3, 0, nil); err == nil {
			result["visitorCenters"] = visitorCenters
		}
		if amenities, err := ps.npsApi.GetAmenities([]string{parkCode}, "", 10, 0); err == nil {
			result["amenities"] = amenities
		}
		if tours, err := ps.npsApi.GetTours(nil, []string{parkCode}, nil, "", 3, 0, nil); err == nil {
			result["tours"] = tours
		}
		if events, err := ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 3, 0, false); err == nil {
			result["events"] = events
		}
		if alerts, err := ps.npsApi.GetAlerts([]string{parkCode}, nil, "", 3, 0); err == nil {
			result["alerts"] = alerts
		}
	}

	return result, nil
}

// GetParkMediaComplete fetches comprehensive multimedia content, preferring cache
func (ps *ParkService) GetParkMediaComplete(parkCode string) (map[string]interface{}, error) {
	return ps.GetParkMedia(parkCode)
}

// GetParkNewsComplete fetches comprehensive news data, preferring cache
func (ps *ParkService) GetParkNewsComplete(parkCode string) (map[string]interface{}, error) {
	return ps.GetParkNews(parkCode)
}

// Helper function to fetch media from API
func (ps *ParkService) fetchMediaFromAPI(parkCode string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Fetch galleries
	galleries, _ := ps.npsApi.GetMultimediaGalleries([]string{parkCode}, nil, "", 0, 10)
	result["galleries"] = galleries

	// Fetch gallery assets for each gallery
	if galleries != nil && len(galleries.Data) > 0 {
		galleryAssets := make(map[string]*nps.MultimediaGalleriesAssetsResponse)
		for _, gallery := range galleries.Data {
			if assets, err := ps.npsApi.GetMultimediaGalleriesAssets("", gallery.ID, []string{parkCode}, nil, "", 0, 50); err == nil {
				galleryAssets[gallery.ID] = assets
			}
		}
		if len(galleryAssets) > 0 {
			result["galleryAssets"] = galleryAssets
		}
	}

	// Fetch videos
	videos, _ := ps.npsApi.GetMultimediaVideos([]string{parkCode}, nil, "", 0, 10)
	result["videos"] = videos

	// Fetch audio
	audio, _ := ps.npsApi.GetMultimediaAudio([]string{parkCode}, nil, "", 0, 10)
	result["audio"] = audio

	// Fetch webcams
	webcams, _ := ps.npsApi.GetWebcams("", []string{parkCode}, nil, "", 10, 0)
	result["webcams"] = webcams

	return result, nil
}

// Helper function to refresh specific media type
func (ps *ParkService) refreshMediaType(parkID int, parkCode, mediaType string) {
	switch mediaType {
	case "galleries":
		if data, err := ps.npsApi.GetMultimediaGalleries([]string{parkCode}, nil, "", 0, 10); err == nil {
			ps.db.UpsertParkData(parkID, mediaType, "park_media", data)
		}
	case "videos":
		if data, err := ps.npsApi.GetMultimediaVideos([]string{parkCode}, nil, "", 0, 10); err == nil {
			ps.db.UpsertParkData(parkID, mediaType, "park_media", data)
		}
	case "audio":
		if data, err := ps.npsApi.GetMultimediaAudio([]string{parkCode}, nil, "", 0, 10); err == nil {
			ps.db.UpsertParkData(parkID, mediaType, "park_media", data)
		}
	case "webcams":
		if data, err := ps.npsApi.GetWebcams("", []string{parkCode}, nil, "", 10, 0); err == nil {
			ps.db.UpsertParkData(parkID, mediaType, "park_media", data)
		}
	}
}

// Helper function to fetch news from API
func (ps *ParkService) fetchNewsFromAPI(parkCode string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Fetch news releases
	newsReleases, _ := ps.npsApi.GetNewsReleases([]string{parkCode}, nil, "", 10, 0, nil)
	result["news_releases"] = newsReleases

	// Fetch articles
	articles, _ := ps.npsApi.GetArticles([]string{parkCode}, nil, "", 10, 0)
	result["articles"] = articles

	// Fetch alerts
	alerts, _ := ps.npsApi.GetAlerts([]string{parkCode}, nil, "", 10, 0)
	result["alerts"] = alerts

	// Fetch events
	events, _ := ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 10, 0, false)
	result["events"] = events

	return result, nil
}

// Helper function to refresh specific news type
func (ps *ParkService) refreshNewsType(parkID int, parkCode, newsType string) {
	switch newsType {
	case "articles":
		if data, err := ps.npsApi.GetArticles([]string{parkCode}, nil, "", 10, 0); err == nil {
			ps.db.UpsertParkData(parkID, newsType, "park_news", data)
		}
	case "alerts":
		if data, err := ps.npsApi.GetAlerts([]string{parkCode}, nil, "", 10, 0); err == nil {
			ps.db.UpsertParkData(parkID, newsType, "park_news", data)
		}
	case "events":
		if data, err := ps.npsApi.GetEvents([]string{parkCode}, nil, nil, nil, nil, nil, nil, nil, "", "", nil, "", "", 10, 0, false); err == nil {
			ps.db.UpsertParkData(parkID, newsType, "park_news", data)
		}
	case "news_releases":
		if data, err := ps.npsApi.GetNewsReleases([]string{parkCode}, nil, "", 10, 0, nil); err == nil {
			ps.db.UpsertParkData(parkID, newsType, "park_news", data)
		}
	}
}

// Helper function to fetch details from API
func (ps *ParkService) fetchDetailsFromAPI(parkCode string) (map[string]interface{}, error) {
	// Get basic park info
	parks, err := ps.npsApi.GetParks([]string{parkCode}, nil, 0, 1, "", nil)
	if err != nil || len(parks.Data) == 0 {
		return nil, fmt.Errorf("park not found")
	}
	park := parks.Data[0]

	// Fetch visitor centers
	visitorCenters, _ := ps.npsApi.GetVisitorCenters([]string{parkCode}, nil, "", 10, 0, nil)

	// Fetch campgrounds
	campgrounds, _ := ps.npsApi.GetCampgrounds([]string{parkCode}, nil, "", 10, 0, nil)

	// Fetch fees and passes
	fees, _ := ps.npsApi.GetFeesPasses([]string{parkCode}, nil, "", 0, 10, nil)

	// Fetch parking lots
	parkingLots, _ := ps.npsApi.GetParkinglots([]string{parkCode}, nil, "", 0, 10)

	return map[string]interface{}{
		"park":           park,
		"visitorCenters": visitorCenters,
		"campgrounds":    campgrounds,
		"fees":           fees,
		"parkingLots":    parkingLots,
	}, nil
}

// Helper function to refresh specific detail type
func (ps *ParkService) refreshDetailType(parkID int, parkCode, detailType string) {
	switch detailType {
	case "visitor_centers":
		if data, err := ps.npsApi.GetVisitorCenters([]string{parkCode}, nil, "", 10, 0, nil); err == nil {
			ps.db.UpsertParkData(parkID, detailType, "park_details", data)
		}
	case "campgrounds":
		if data, err := ps.npsApi.GetCampgrounds([]string{parkCode}, nil, "", 10, 0, nil); err == nil {
			ps.db.UpsertParkData(parkID, detailType, "park_details", data)
		}
	case "fees":
		if data, err := ps.npsApi.GetFeesPasses([]string{parkCode}, nil, "", 0, 10, nil); err == nil {
			ps.db.UpsertParkData(parkID, detailType, "park_details", data)
		}
	case "amenities":
		if data, err := ps.npsApi.GetAmenities([]string{parkCode}, "", 20, 0); err == nil {
			ps.db.UpsertParkData(parkID, detailType, "park_details", data)
		}
	case "parking_lots":
		if data, err := ps.npsApi.GetParkinglots([]string{parkCode}, nil, "", 0, 10); err == nil {
			ps.db.UpsertParkData(parkID, detailType, "park_details", data)
		}
	}
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
