package nps

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/*
 Go wrapper for the National Park Service API.
 https://www.nps.gov/subjects/developer/api-documentation.htm
*/

// NpsService describes the methods for interacting with the National Park Service API.
type NpsApi interface {
	GetActivities(id, q, sort string, limit, start int) (*ActivityResponse, error)
	GetActivityParks(id []string, q, sort string, limit, start int) (*ActivityParkResponse, error)
	GetAlerts(parkCode, stateCode []string, q string, limit, start int) (*AlertResponse, error)
	GetAmenities(id []string, q string, limit, start int) (*AmenityResponse, error)
	GetAmenitiesParksPlaces(parkCode, id []string, q, sort string, limit, start int) (*AmenityParkPlaceResponse, error)
	GetAmenitiesParksVisitorCenters(parkCode, id, q string, sort []string, limit, start int) (*AmenityParkVisitorCenterResponse, error)
	GetArticles(parkCode, stateCode []string, q string, limit, start int) (*ArticleData, error)
	GetCampgrounds(parkCode, stateCode []string, q string, sort []string, limit, start int) (*CampgroundData, error)
	GetEvents(parkCode, stateCode, organization, subject, portal, tagsAll, tagsOne, tagsNone []string, dateStart, dateEnd string, eventType []string, id, q string, pageSize, pageNumber int, expandRecurring bool) (*EventResponse, error)
	GetFeesPasses(parkCode, stateCode []string, start, limit int, q string, sort []string) (*FeePassResponse, error)
	GetLessonPlans(parkCode, stateCode []string, start, limit int, q string, sort []string) (*LessonPlanResponse, error)
	GetParkBoundaries(sitecode string) (*MapdataParkboundaryResponse, error)
	GetMultimediaAudio(parkCode, stateCode []string, start, limit int, q string) (*MultimediaAudioResponse, error)
	GetMultimediaGalleries(parkCode, stateCode []string, start, limit int, q string) (*MultimediaGalleriesResponse, error)
	GetMultimediaGalleriesAssets(id, galleryId string, parkCode, stateCode []string, start, limit int, q string) (*MultimediaGalleriesAssetsResponse, error)
	GetMultimediaVideos(parkCode, stateCode []string, start, limit int, q string) (*MultimediaVideosResponse, error)
	GetNewsReleases(parkCode, stateCode []string, q string, limit, start int, sort []string) (*NewsReleaseResponse, error)
	GetParkinglots(parkCode, stateCode []string, start, limit int, q string) (*ParkinglotResponse, error)
	GetParks(parkCode, stateCode []string, start, limit int, q string, sort []string) (*ParkResponse, error)
	GetPassportStampLocations(parkCode, stateCode []string, q string, limit, start int) (*PassportStampLocationResponse, error)
	GetPeople(parkCode, stateCode []string, q string, limit, start int) (*PersonResponse, error)
	GetPlaces(parkCode, stateCode []string, q string, limit, start int) (*PlaceResponse, error)
	GetRoadEvents(parkCode, eventType string) (*RoadEventResponse, error)
	GetThingsToDo(id, parkCode, stateCode, q string, limit, start int, sort []string) (*ThingsToDoResponse, error)
	GetTopics(id, q string, limit, start int, sort string) (*TopicResponse, error)
	GetTopicParks(id []string, q string, limit, start int, sort string) (*TopicParkResponse, error)
	GetTours(id, parkCode, stateCode []string, q string, limit, start int, sort []string) (*TourResponse, error)
	GetVisitorCenters(parkCode, stateCode []string, q string, limit, start int, sort []string) (*VisitorCenterResponse, error)
	GetWebcams(id string, parkCode, stateCode []string, q string, limit, start int) (*WebcamResponse, error)
}

// npsApi implements the NpsService interface.
type npsApi struct {
	BaseURL string
	Client  *http.Client
}

func NewNpsApi(baseurl string) NpsApi {
	return &npsApi{
		BaseURL: baseurl,
		Client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

// Activity represents an activity in the National Park Service.
type Activity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ActivityResponse represents the response from the /activities endpoint.
type ActivityResponse struct {
	Total string       `json:"total"`
	Data  [][]Activity `json:"data"`
	Limit string       `json:"limit"`
	Start string       `json:"start"`
}

// GetActivities makes a GET request to the /activities endpoint and returns the activities.
func (api *npsApi) GetActivities(id, q, sort string, limit, start int) (*ActivityResponse, error) {
	url := api.BaseURL + "/activities?id=" + id + "&q=" + q + "&sort=" + sort + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start)
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var activitiesResponse ActivityResponse
	if err := json.NewDecoder(resp.Body).Decode(&activitiesResponse); err != nil {
		return nil, err
	}

	return &activitiesResponse, nil
}

// ActivityPark represents a park related to an activity in the National Park Service.
type ActivityParkData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Parks []struct {
		States      string `json:"states"`
		FullName    string `json:"fullName"`
		URL         string `json:"url"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
		Name        string `json:"name"`
	} `json:"parks"`
}

// ActivityParkResponse represents the response from the /activities/parks endpoint.
type ActivityParkResponse struct {
	Total string             `json:"total"`
	Data  []ActivityParkData `json:"data"`
	Limit string             `json:"limit"`
	Start string             `json:"start"`
}

// GetActivityParks makes a GET request to the /activities/parks endpoint and returns the parks related to the activities.
func (api *npsApi) GetActivityParks(id []string, q, sort string, limit, start int) (*ActivityParkResponse, error) {
	url := api.BaseURL + "/activities/parks?id=" + strings.Join(id, ",") + "&q=" + q + "&sort=" + sort + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start)
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var activityParkResponse ActivityParkResponse
	if err := json.NewDecoder(resp.Body).Decode(&activityParkResponse); err != nil {
		return nil, err
	}

	return &activityParkResponse, nil
}

// Alert represents an alert in the National Park Service.
type Alert struct {
	Category          string `json:"category"`
	Description       string `json:"description"`
	ID                string `json:"id"`
	ParkCode          string `json:"parkCode"`
	Title             string `json:"title"`
	URL               string `json:"url"`
	LastIndexedDate   string `json:"lastIndexedDate"`
	RelatedRoadEvents []struct {
		Title string `json:"title"`
		ID    string `json:"id"`
		Type  string `json:"type"`
		URL   string `json:"url"`
	} `json:"relatedRoadEvents"`
}

// AlertResponse represents the response from the /alerts endpoint.
type AlertResponse struct {
	Total string    `json:"total"`
	Data  [][]Alert `json:"data"`
	Limit string    `json:"limit"`
	Start string    `json:"start"`
}

// GetAlerts makes a GET request to the /alerts endpoint and returns the alerts.
func (api *npsApi) GetAlerts(parkCode, stateCode []string, q string, limit, start int) (*AlertResponse, error) {
	url := api.BaseURL + "/alerts?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&q=" + q + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start)
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var alertsResponse AlertResponse
	if err := json.NewDecoder(resp.Body).Decode(&alertsResponse); err != nil {
		return nil, err
	}

	return &alertsResponse, nil
}

// Amenity represents an amenity in the National Park Service.
type Amenity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AmenityResponse represents the response from the /amenities endpoint.
type AmenityResponse struct {
	Total string    `json:"total"`
	Data  []Amenity `json:"data"`
	Limit string    `json:"limit"`
	Start string    `json:"start"`
}

// GetAmenities makes a GET request to the /amenities endpoint and returns the amenities.
func (api *npsApi) GetAmenities(id []string, q string, limit, start int) (*AmenityResponse, error) {
	url := api.BaseURL + "/amenities?id=" + strings.Join(id, ",") + "&q=" + q + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start)
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var amenitiesResponse AmenityResponse
	if err := json.NewDecoder(resp.Body).Decode(&amenitiesResponse); err != nil {
		return nil, err
	}

	return &amenitiesResponse, nil
}

// Place represents a place in the National Park Service.
type AmenityParkPlace struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Parks []struct {
		States      string `json:"states"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
		FullName    string `json:"fullName"`
		Places      []struct {
			IsManagedByNps   string `json:"isManagedByNps"`
			AudioDescription string `json:"audioDescription"`
			Multimedia       []struct {
				Title string `json:"title"`
				ID    string `json:"id"`
				Type  string `json:"type"`
				Url   string `json:"url"`
			} `json:"multimedia"`
			Latitude             string   `json:"latitude"`
			ManagedByOrg         string   `json:"managedByOrg"`
			Url                  string   `json:"url"`
			Longitude            string   `json:"longitude"`
			BodyText             string   `json:"bodyText"`
			GeometryPoiId        string   `json:"geometryPoiId"`
			NpmapId              string   `json:"npmapId"`
			RelatedOrganizations []string `json:"relatedOrganizations"`
			Amenities            []string `json:"amenities"`
			Title                string   `json:"title"`
			Images               []string `json:"images"`
			ListingDescription   string   `json:"listingDescription"`
			IsOpenToPublic       string   `json:"isOpenToPublic"`
			Tags                 []string `json:"tags"`
			ManagedByUrl         string   `json:"managedByUrl"`
			QuickFacts           string   `json:"quickFacts"`
			LatLong              string   `json:"latLong"`
			ID                   string   `json:"id"`
			RelatedParks         []struct {
				States      string `json:"states"`
				ParkCode    string `json:"parkCode"`
				Designation string `json:"designation"`
				FullName    string `json:"fullName"`
				Url         string `json:"url"`
				Name        string `json:"name"`
			} `json:"relatedParks"`
		} `json:"places"`
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"parks"`
}

// AmenityParkPlaceResponse represents the response from the /amenities/parksplaces endpoint.
type AmenityParkPlaceResponse struct {
	Total string             `json:"total"`
	Data  []AmenityParkPlace `json:"data"`
	Limit string             `json:"limit"`
	Start string             `json:"start"`
}

// GetAmenitiesParksPlaces makes a GET request to the /amenities/parksplaces endpoint and returns the parks with places related to the amenities.
func (api *npsApi) GetAmenitiesParksPlaces(parkCode, id []string, q, sort string, limit, start int) (*AmenityParkPlaceResponse, error) {
	url := api.BaseURL + "/amenities/parksplaces?parkCode=" + strings.Join(parkCode, ",") + "&id=" + strings.Join(id, ",") + "&q=" + q + "&sort=" + sort + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start)
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var amenityParkPlaceResponse AmenityParkPlaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&amenityParkPlaceResponse); err != nil {
		return nil, err
	}

	return &amenityParkPlaceResponse, nil
}

// VisitorCenter represents a visitor center in the National Park Service.
type AmenityParkVisitorCenterData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Parks []struct {
		States         string `json:"states"`
		ParkCode       string `json:"parkCode"`
		Designation    string `json:"designation"`
		VisitorCenters []struct {
			URL  string `json:"url"`
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"visitorcenters"`
		FullName string `json:"fullName"`
		URL      string `json:"url"`
		Name     string `json:"Name"`
	} `json:"parks"`
}

// AmenityParkVisitorCenterResponse represents the response from the /amenities/parksvisitorcenters endpoint.
type AmenityParkVisitorCenterResponse struct {
	Total string                         `json:"total"`
	Data  []AmenityParkVisitorCenterData `json:"data"`
	Limit string                         `json:"limit"`
	Start string                         `json:"start"`
}

// GetAmenitiesParksVisitorCenters makes a GET request to the /amenities/parksvisitorcenters endpoint and returns the parks with visitor centers related to the amenities.
func (api *npsApi) GetAmenitiesParksVisitorCenters(parkCode, id, q string, sort []string, limit, start int) (*AmenityParkVisitorCenterResponse, error) {
	url := api.BaseURL + "/amenities/parksvisitorcenters?parkCode=" + parkCode + "&id=" + id + "&q=" + q + "&sort=" + strings.Join(sort, ",") + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start)
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var amenityParkVisitorCenterResponse AmenityParkVisitorCenterResponse
	if err := json.NewDecoder(resp.Body).Decode(&amenityParkVisitorCenterResponse); err != nil {
		return nil, err
	}

	return &amenityParkVisitorCenterResponse, nil
}

// Article represents an article in the National Park Service.
type Article struct {
	URL                string `json:"url"`
	Title              string `json:"title"`
	ID                 string `json:"id"`
	GeometryPoiId      string `json:"geometryPoiId"`
	ListingDescription string `json:"listingDescription"`
	ListingImage       struct {
		URL         string `json:"url"`
		Credit      string `json:"credit"`
		AltText     string `json:"altText"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Caption     string `json:"caption"`
	} `json:"listingImage"`
	RelatedParks []struct {
		States      string `json:"states"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
		FullName    string `json:"fullName"`
		URL         string `json:"url"`
		Name        string `json:"name"`
	} `json:"relatedParks"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	LatLong   string  `json:"latLong"`
}

// ArticleData represents the data in the response from the /articles endpoint.
type ArticleData struct {
	Total string    `json:"total"`
	Data  []Article `json:"data"`
	Limit string    `json:"limit"`
	Start string    `json:"start"`
}

// GetArticles makes a GET request to the /articles endpoint and returns the articles.
func (api *npsApi) GetArticles(parkCode, stateCode []string, q string, limit, start int) (*ArticleData, error) {
	url := api.BaseURL + "/articles?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&q=" + q + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start)
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var articleData ArticleData
	if err := json.NewDecoder(resp.Body).Decode(&articleData); err != nil {
		return nil, err
	}

	return &articleData, nil
}

// Campground represents a campground in the National Park Service.
type Campground struct {
	Accessibility []struct {
		Wheelchairaccess string   `json:"wheelchairaccess"`
		Internetinfo     string   `json:"internetinfo"`
		Rvallowed        int      `json:"rvallowed"`
		Cellphoneinfo    string   `json:"cellphoneinfo"`
		Firestovepolicy  string   `json:"firestovepolicy"`
		Rvmaxlength      int      `json:"rvmaxlength"`
		Additionalinfo   string   `json:"additionalinfo"`
		Trailermaxlength int      `json:"trailermaxlength"`
		Adainfo          string   `json:"adainfo"`
		Rvinfo           string   `json:"rvinfo"`
		Accessroads      []string `json:"accessroads"`
		Trailerallowed   int      `json:"trailerallowed"`
		Classifications  []string `json:"classifications"`
	} `json:"accessibility"`
	Addresses []struct {
		PostalCode  string `json:"postalCode"`
		City        string `json:"city"`
		StateCode   string `json:"stateCode"`
		CountryCode string `json:"countryCode"`
		Line1       string `json:"line1"`
		Line2       string `json:"line2"`
		Line3       string `json:"line3"`
		Type        string `json:"type"`
	} `json:"addresses"`
	Amenities []struct {
		Trashrecyclingcollection   string   `json:"trashrecyclingcollection"`
		Toilets                    []string `json:"toilets"`
		Internetconnectivity       bool     `json:"internetconnectivity"`
		Showers                    []string `json:"showers"`
		Cellphonereception         bool     `json:"cellphonereception"`
		Laundry                    bool     `json:"laundry"`
		Amphitheater               string   `json:"amphitheater"`
		Dumpstation                bool     `json:"dumpstation"`
		Campstore                  bool     `json:"campstore"`
		Stafforvolunteerhostonsite string   `json:"stafforvolunteerhostonsite"`
		Potablewater               []string `json:"potablewater"`
		Iceavailableforsale        bool     `json:"iceavailableforsale"`
		Firewoodforsale            bool     `json:"firewoodforsale"`
		Ampitheater                string   `json:"ampitheater"`
		Foodstoragelockers         string   `json:"foodstoragelockers"`
	} `json:"amenities"`
	Campsites []struct {
		Other             int `json:"other"`
		Group             int `json:"group"`
		Horse             int `json:"horse"`
		Totalsites        int `json:"totalsites"`
		Tentonly          int `json:"tentonly"`
		Electricalhookups int `json:"electricalhookups"`
		Rvonly            int `json:"rvonly"`
		Walkboatto        int `json:"walkboatto"`
	} `json:"campsites"`
	Contacts []struct {
		PhoneNumbers []struct {
			PhoneNumber string `json:"phoneNumber"`
			Description string `json:"description"`
			Extension   string `json:"extension"`
			Type        string `json:"type"`
		} `json:"phoneNumbers"`
		EmailAddresses []struct {
			EmailAddress string `json:"emailAddress"`
			Description  string `json:"description"`
		} `json:"emailAddresses"`
	} `json:"contacts"`
	Description        string `json:"description"`
	Directionsoverview string `json:"directionsoverview"`
	DirectionsUrl      string `json:"directionsUrl"`
	Id                 int    `json:"id"`
	LatLong            string `json:"latLong"`
	Name               string `json:"name"`
	ParkCode           string `json:"parkCode"`
	Weatheroverview    string `json:"weatheroverview"`
}

// CampgroundData represents the data in the response from the /campgrounds endpoint.
type CampgroundData struct {
	Total string       `json:"total"`
	Data  []Campground `json:"data"`
	Limit string       `json:"limit"`
	Start string       `json:"start"`
}

// GetCampgrounds makes a GET request to the /campgrounds endpoint and returns the campgrounds.
func (api *npsApi) GetCampgrounds(parkCode, stateCode []string, q string, sort []string, limit, start int) (*CampgroundData, error) {
	url := api.BaseURL + "/campgrounds?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&q=" + q + "&sort=" + strings.Join(sort, ",") + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start)
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var campgroundData CampgroundData
	if err := json.NewDecoder(resp.Body).Decode(&campgroundData); err != nil {
		return nil, err
	}

	return &campgroundData, nil
}

// Event represents an event in the National Park Service.
type Event struct {
	Category    string   `json:"category"`
	CategoryID  string   `json:"categoryid"`
	Date        string   `json:"date"`
	DateEnd     string   `json:"dateend"`
	Dates       []string `json:"dates"`
	DateStart   string   `json:"datestart"`
	Description string   `json:"description"`
	EventID     string   `json:"eventid"`
	ID          string   `json:"id"`
	IsAllDay    bool     `json:"isallday"`
	IsFree      bool     `json:"isfree"`
	IsRecurring bool     `json:"isrecurring"`
	Location    string   `json:"location"`
	Title       string   `json:"title"`
}

// EventResponse represents the response from the /events endpoint.
type EventResponse struct {
	Total      string    `json:"total"`
	Errors     []string  `json:"errors"` // Assuming errors are strings.
	Data       [][]Event `json:"data"`
	Dates      string    `json:"dates"`
	PageNumber string    `json:"pagenumber"`
	PageSize   string    `json:"pagesize"`
}

// GetEvents makes a GET request to the /events endpoint and returns the events.
func (api *npsApi) GetEvents(parkCode, stateCode, organization, subject, portal, tagsAll, tagsOne, tagsNone []string, dateStart, dateEnd string, eventType []string, id, q string, pageSize, pageNumber int, expandRecurring bool) (*EventResponse, error) {
	url := api.BaseURL + "/events?parkCode=" + strings.Join(parkCode, ",") + "&organization=" + strings.Join(organization, ",") + "&subject=" + strings.Join(subject, ",") + "&portal=" + strings.Join(portal, ",") + "&tagsAll=" + strings.Join(tagsAll, ",") + "&tagsOne=" + strings.Join(tagsOne, ",") + "&tagsNone=" + strings.Join(tagsNone, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&dateStart=" + dateStart + "&dateEnd=" + dateEnd + "&eventType=" + strings.Join(eventType, ",") + "&id=" + id + "&q=" + q + "&pageSize=" + strconv.Itoa(pageSize) + "&pageNumber=" + strconv.Itoa(pageNumber) + "&expandRecurring=" + strconv.FormatBool(expandRecurring)
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var eventsResponse EventResponse
	if err := json.NewDecoder(resp.Body).Decode(&eventsResponse); err != nil {
		return nil, err
	}

	return &eventsResponse, nil
}

// FeePass represents a fee or pass in the National Park Service.
type FeePass struct {
	CustomFeeLinkUrl                     string `json:"customFeeLinkUrl"`
	TimedEntryDescription                string `json:"timedEntryDescription"`
	ParkingDetailsUrl                    string `json:"parkingDetailsUrl"`
	EntrancePassesDescription            string `json:"entrancePassesDescription"`
	TimedEntryHeading                    string `json:"timedEntryHeading"`
	CustomFeeDescription                 string `json:"customFeeDescription"`
	IsParkingFeePossible                 bool   `json:"isParkingFeePossible"`
	EntranceFeeDescription               string `json:"entranceFeeDescription"`
	Cashless                             string `json:"cashless"`
	CustomFeeHeading                     string `json:"customFeeHeading"`
	IsInteragencyPassAccepted            bool   `json:"isInteragencyPassAccepted"`
	PaidParkingDescription               string `json:"paidParkingDescription"`
	IsFeeFreePark                        bool   `json:"isFeeFreePark"`
	PaidParkingHeading                   string `json:"paidParkingHeading"`
	ParkCode                             string `json:"parkCode"`
	ContentOrderOrdinals                 []int  `json:"contentOrderOrdinals"`
	IsParkingOrTransportationFeePossible bool   `json:"isParkingOrTransportationFeePossible"`
	CustomFeeLinkText                    string `json:"customFeeLinkText"`
	FeesAtWorkUrl                        string `json:"feesAtWorkUrl"`
}

// FeePassResponse represents the response from the /feespasses endpoint.
type FeePassResponse struct {
	Total string      `json:"total"`
	Data  [][]FeePass `json:"data"`
	Start string      `json:"start"`
	Limit string      `json:"limit"`
}

// GetFeesPasses makes a GET request to the /feespasses endpoint and returns the fees and passes.
func (api *npsApi) GetFeesPasses(parkCode, stateCode []string, start, limit int, q string, sort []string) (*FeePassResponse, error) {
	url := api.BaseURL + "/feespasses?parkCode=" + strings.Join(parkCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q + "&sort=" + strings.Join(sort, ",") + "&stateCode=" + strings.Join(stateCode, ",")
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var feesPassesResponse FeePassResponse
	if err := json.NewDecoder(resp.Body).Decode(&feesPassesResponse); err != nil {
		return nil, err
	}

	return &feesPassesResponse, nil
}

// LessonPlan represents a lesson plan in the National Park Service.
type LessonPlan struct {
	Parks       []string `json:"parks"`
	Description string   `json:"description"`
	CommonCore  struct {
		StateStandards      string   `json:"statestandards"`
		MathStandards       []string `json:"mathstandards"`
		AdditionalStandards string   `json:"additionalstandards"`
		ElaStandards        []string `json:"elastandards"`
	} `json:"commoncore"`
	Subject           string `json:"subject"`
	GradeLevel        string `json:"gradelevel"`
	Url               string `json:"url"`
	QuestionObjective string `json:"questionobjective"`
	Duration          string `json:"duration"`
	Title             string `json:"title"`
	ID                string `json:"id"`
}

// LessonPlanResponse represents the response from the /lessonplans endpoint.
type LessonPlanResponse struct {
	Total string         `json:"total"`
	Data  [][]LessonPlan `json:"data"`
	Start string         `json:"start"`
	Limit string         `json:"limit"`
}

// GetLessonPlans makes a GET request to the /lessonplans endpoint and returns the lesson plans.
func (api *npsApi) GetLessonPlans(parkCode, stateCode []string, start, limit int, q string, sort []string) (*LessonPlanResponse, error) {
	url := api.BaseURL + "/lessonplans?parkCode=" + strings.Join(parkCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q + "&sort=" + strings.Join(sort, ",") + "&stateCode=" + strings.Join(stateCode, ",")
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var lessonPlansResponse LessonPlanResponse
	if err := json.NewDecoder(resp.Body).Decode(&lessonPlansResponse); err != nil {
		return nil, err
	}

	return &lessonPlansResponse, nil
}

// MapdataParkboundary represents a park boundary in the National Park Service.
type MapdataParkboundary struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Properties struct {
		ParkClass string `json:"parkClass"`
	} `json:"properties"`
	Geometry struct {
		Coordinates [][][]struct {
			Longitude float64 `json:"0"`
			Latitude  float64 `json:"1"`
		} `json:"coordinates"`
		Type string `json:"type"`
	} `json:"geometry"`
}

// MapdataParkboundaryResponse represents the response from the /mapdata/parkboundaries/{sitecode} endpoint.
type MapdataParkboundaryResponse struct {
	Total string            `json:"total"`
	Data  []MultimediaAudio `json:"data"`
	Start string            `json:"start"`
	Limit string            `json:"limit"`
}

// GetParkBoundaries makes a GET request to the /mapdata/parkboundaries/{sitecode} endpoint and returns the park boundaries.
func (api *npsApi) GetParkBoundaries(sitecode string) (*MapdataParkboundaryResponse, error) {
	url := api.BaseURL + "/mapdata/parkboundaries/" + sitecode
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parkBoundariesResponse MapdataParkboundaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&parkBoundariesResponse); err != nil {
		return nil, err
	}

	return &parkBoundariesResponse, nil
}

// MultimediaAudio represents an audio file in the National Park Service.
type MultimediaAudio struct {
	CallToActionUrl string  `json:"callToActionUrl"`
	PermalinkUrl    string  `json:"permalinkUrl"`
	Latitude        float64 `json:"latitude"`
	CallToAction    string  `json:"callToAction"`
	Longitude       float64 `json:"longitude"`
	GeometryPoiId   string  `json:"geometryPoiId"`
	SplashImage     struct {
		Url string `json:"url"`
	} `json:"splashImage"`
	Transcript string   `json:"transcript"`
	Title      string   `json:"title"`
	Tags       []string `json:"tags"`
	Credit     string   `json:"credit"`
	DurationMs int64    `json:"durationMs"`
	ID         string   `json:"id"`
	Versions   struct {
		FileSize int64  `json:"fileSize"`
		FileType string `json:"fileType"`
		Url      string `json:"url"`
	} `json:"versions"`
	Description  string `json:"description"`
	RelatedParks []struct {
		States      string `json:"states"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
	} `json:"relatedParks"`
}

// MultimediaAudioResponse represents the response from the /multimedia/audio endpoint.
type MultimediaAudioResponse struct {
	Total string            `json:"total"`
	Data  []MultimediaAudio `json:"data"`
	Start string            `json:"start"`
	Limit string            `json:"limit"`
}

// GetMultimediaAudio makes a GET request to the /multimedia/audio endpoint and returns the audio files.
func (api *npsApi) GetMultimediaAudio(parkCode, stateCode []string, start, limit int, q string) (*MultimediaAudioResponse, error) {
	url := api.BaseURL + "/multimedia/audio?parkCode=" + strings.Join(parkCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q + "&stateCode=" + strings.Join(stateCode, ",")
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var multimediaAudioResponse MultimediaAudioResponse
	if err := json.NewDecoder(resp.Body).Decode(&multimediaAudioResponse); err != nil {
		return nil, err
	}

	return &multimediaAudioResponse, nil
}

// MultimediaGalleries represents a gallery in the National Park Service.
type MultimediaGalleries struct {
	ConstraintsInfo struct {
		Constraint     string `json:"constraint"`
		GrantingRights string `json:"grantingRights"`
	} `json:"constraintsInfo"`
	Copyright string `json:"copyright"`
	Url       string `json:"url"`
	Title     string `json:"title"`
	Images    []struct {
		Url         string `json:"url"`
		AltText     string `json:"altText"`
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"images"`
	Tags         []string `json:"tags"`
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	RelatedParks []struct {
		States      string `json:"states"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
		FullName    string `json:"fullName"`
		Url         string `json:"url"`
		Name        string `json:"name"`
	} `json:"relatedParks"`
	AssetCount string `json:"assetCount"`
}

// MultimediaGalleriesResponse represents the response from the /multimedia/galleries endpoint.
type MultimediaGalleriesResponse struct {
	Total string                `json:"total"`
	Data  []MultimediaGalleries `json:"data"`
	Start string                `json:"start"`
	Limit string                `json:"limit"`
}

// GetMultimediaGalleries makes a GET request to the /multimedia/galleries endpoint and returns the galleries.
func (api *npsApi) GetMultimediaGalleries(parkCode, stateCode []string, start, limit int, q string) (*MultimediaGalleriesResponse, error) {
	url := api.BaseURL + "/multimedia/galleries?parkCode=" + strings.Join(parkCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q + "&stateCode=" + strings.Join(stateCode, ",")
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var multimediaGalleriesResponse MultimediaGalleriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&multimediaGalleriesResponse); err != nil {
		return nil, err
	}

	return &multimediaGalleriesResponse, nil
}

// MultimediaGalleriesAssets represents a gallery asset in the National Park Service.
type MultimediaGalleriesAssets struct {
	ConstraintsInfo struct {
		Constraint     string `json:"constraint"`
		GrantingRights string `json:"grantingRights"`
	} `json:"constraintsInfo"`
	PermalinkUrl string `json:"permalinkUrl"`
	Copyright    string `json:"copyright"`
	FileInfo     struct {
		Url          string `json:"url"`
		FileType     string `json:"fileType"`
		WidthPixels  string `json:"widthPixels"`
		HeightPixels string `json:"heightPixels"`
		FileSizeKb   string `json:"fileSizeKb"`
	} `json:"fileInfo"`
	Ordinal      string   `json:"ordinal"`
	AltText      string   `json:"altText"`
	Title        string   `json:"title"`
	Tags         []string `json:"tags"`
	Credit       string   `json:"credit"`
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	RelatedParks []struct {
		States      string `json:"states"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
		FullName    string `json:"fullName"`
		Url         string `json:"url"`
		Name        string `json:"name"`
	} `json:"relatedParks"`
}

// MultimediaGalleriesAssetsResponse represents the response from the /multimedia/galleries/assets endpoint.
type MultimediaGalleriesAssetsResponse struct {
	Total string                      `json:"total"`
	Data  []MultimediaGalleriesAssets `json:"data"`
	Start string                      `json:"start"`
	Limit string                      `json:"limit"`
}

// GetMultimediaGalleriesAssets makes a GET request to the /multimedia/galleries/assets endpoint and returns the gallery assets.
func (api *npsApi) GetMultimediaGalleriesAssets(id, galleryId string, parkCode, stateCode []string, start, limit int, q string) (*MultimediaGalleriesAssetsResponse, error) {
	url := api.BaseURL + "/multimedia/galleries/assets?id=" + id + "&galleryId=" + galleryId + "&parkCode=" + strings.Join(parkCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q + "&stateCode=" + strings.Join(stateCode, ",")
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var multimediaGalleriesAssetsResponse MultimediaGalleriesAssetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&multimediaGalleriesAssetsResponse); err != nil {
		return nil, err
	}

	return &multimediaGalleriesAssetsResponse, nil
}

// MultimediaVideos represents a video in the National Park Service.
type MultimediaVideos struct {
	CallToActionURL       string   `json:"callToActionURL"`
	AudioDescribedBuiltIn bool     `json:"audioDescribedBuiltIn"`
	DescriptiveTranscript string   `json:"descriptiveTranscript"`
	PermalinkUrl          string   `json:"permalinkUrl"`
	Audiodescription      string   `json:"audiodescription"`
	AudioDescriptionUrl   string   `json:"audioDescriptionUrl"`
	Latitude              float64  `json:"latitude"`
	CallToAction          string   `json:"callToAction"`
	Longitude             float64  `json:"longitude"`
	HasOpenCaptions       bool     `json:"hasOpenCaptions"`
	GeometryPoiId         string   `json:"geometryPoiId"`
	IsBRoll               bool     `json:"isBRoll"`
	Transcript            string   `json:"transcript"`
	Title                 string   `json:"title"`
	Tags                  []string `json:"tags"`
	Credit                string   `json:"credit"`
	DurationMs            int      `json:"durationMs"`
	ID                    string   `json:"id"`
	Description           string   `json:"description"`
	RelatedParks          []struct {
		States      string `json:"states"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
		FullName    string `json:"fullName"`
		Url         string `json:"url"`
		Name        string `json:"name"`
	} `json:"relatedParks"`
}

// MultimediaVideosResponse represents the response from the /multimedia/videos endpoint.
type MultimediaVideosResponse struct {
	Total string             `json:"total"`
	Data  []MultimediaVideos `json:"data"`
	Start string             `json:"start"`
	Limit string             `json:"limit"`
}

// GetMultimediaVideos makes a GET request to the /multimedia/videos endpoint and returns the videos.
func (api *npsApi) GetMultimediaVideos(parkCode, stateCode []string, start, limit int, q string) (*MultimediaVideosResponse, error) {
	url := api.BaseURL + "/multimedia/videos?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var multimediaVideosResponse MultimediaVideosResponse
	if err := json.NewDecoder(resp.Body).Decode(&multimediaVideosResponse); err != nil {
		return nil, err
	}

	return &multimediaVideosResponse, nil
}

type NewsRelease struct {
	Abstract      string `json:"abstract"`
	Latitude      string `json:"latitude"`
	Url           string `json:"url"`
	Longitude     string `json:"longitude"`
	GeometryPoiId string `json:"geometryPoiId"`
	ReleaseDate   string `json:"releaseDate"`
	ParkCode      string `json:"parkCode"`
	Title         string `json:"title"`
	RelatedOrgs   []struct {
		ID   string `json:"id"`
		Url  string `json:"url"`
		Name string `json:"name"`
	} `json:"relatedOrgs"`
	ID    string `json:"id"`
	Image struct {
		Credit      string `json:"credit"`
		AltText     string `json:"altText"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Caption     string `json:"caption"`
		Url         string `json:"url"`
	} `json:"image"`
	RelatedParks []struct {
		States      string `json:"states"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
		FullName    string `json:"fullName"`
		Url         string `json:"url"`
		Name        string `json:"name"`
	} `json:"relatedParks"`
}

// NewsReleaseResponse represents the response from the /newsreleases endpoint.
type NewsReleaseResponse struct {
	Total string        `json:"total"`
	Data  []NewsRelease `json:"data"`
	Start string        `json:"start"`
	Limit string        `json:"limit"`
}

// GetNewsReleases makes a GET request to the /newsreleases endpoint and returns the news releases.
func (api *npsApi) GetNewsReleases(parkCode, stateCode []string, q string, limit, start int, sort []string) (*NewsReleaseResponse, error) {
	url := api.BaseURL + "/newsreleases?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q + "&sort=" + strings.Join(sort, ",")
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var newsReleaseResponse NewsReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsReleaseResponse); err != nil {
		return nil, err
	}

	return &newsReleaseResponse, nil
}

// Parkinglot represents a parking lot in the National Park Service.
type Parkinglot struct {
	ManagedByOrganization string `json:"managedByOrganization"`
	Name                  string `json:"name"`
	Latitude              string `json:"latitude"`
	Fees                  []struct {
		Cost        string `json:"cost"`
		Description string `json:"description"`
		Title       string `json:"title"`
	} `json:"fees"`
	Accessibility struct {
		IsLotAccessibleToDisabled      bool `json:"isLotAccessibleToDisabled"`
		NumberOfOversizeVehicleSpaces  int  `json:"numberOfOversizeVehicleSpaces"`
		NumberofAdaSpaces              int  `json:"numberofAdaSpaces"`
		NumberofAdaStepFreeSpaces      int  `json:"numberofAdaStepFreeSpaces"`
		NumberofAdaVanAccessbileSpaces int  `json:"numberofAdaVanAccessbileSpaces"`
		TotalSpaces                    int  `json:"totalSpaces"`
	} `json:"accessibility"`
	OperatingHours []struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		StandardHours struct {
			Sunday    string `json:"sunday"`
			Monday    string `json:"monday"`
			Tuesday   string `json:"tuesday"`
			Wednesday string `json:"wednesday"`
			Thursday  string `json:"thursday"`
			Friday    string `json:"friday"`
			Saturday  string `json:"saturday"`
		} `json:"standardHours"`
		Exceptions []struct {
			Name           string `json:"name"`
			StartDate      string `json:"startDate"`
			EndDate        string `json:"endDate"`
			ExceptionHours struct {
				Sunday    string `json:"sunday"`
				Monday    string `json:"monday"`
				Tuesday   string `json:"tuesday"`
				Wednesday string `json:"wednesday"`
				Thursday  string `json:"thursday"`
				Friday    string `json:"friday"`
				Saturday  string `json:"saturday"`
			} `json:"exceptionHours"`
		} `json:"exceptions"`
	} `json:"operatingHours"`
	Longitude string `json:"longitude"`
	Contacts  struct {
		PhoneNumbers []struct {
			PhoneNumber string `json:"phoneNumber"`
			Description string `json:"description"`
			Extension   string `json:"extension"`
			Type        string `json:"type"`
		} `json:"phoneNumbers"`
		EmailAddresses []struct {
			EmailAddress string `json:"emailAddress"`
			Description  string `json:"description"`
		} `json:"emailAddresses"`
	} `json:"contacts"`
	GeometryPoiId string `json:"geometryPoiId"`
	WebcamUrl     string `json:"webcamUrl"`
	AltName       string `json:"altName"`
	TimeZone      string `json:"timeZone"`
	ID            string `json:"id"`
	Description   string `json:"description"`
	RelatedParks  []struct {
		States      string `json:"states"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
		FullName    string `json:"fullName"`
		Url         string `json:"url"`
		Name        string `json:"name"`
	} `json:"relatedParks"`
	LiveStatus struct {
		Description                string `json:"description"`
		EstimatedWaitTimeInMinutes int    `json:"estimatedWaitTimeInMinutes"`
		ExpirationDate             string `json:"expirationDate"`
		IsActive                   bool   `json:"isActive"`
		Occupancy                  string `json:"occupancy"`
	} `json:"liveStatus"`
}

// ParkinglotResponse represents the response from the /parkinglots endpoint.
type ParkinglotResponse struct {
	Total string       `json:"total"`
	Data  []Parkinglot `json:"data"`
	Start string       `json:"start"`
	Limit string       `json:"limit"`
}

// GetParkinglots makes a GET request to the /parkinglots endpoint and returns the parking lots.
func (api *npsApi) GetParkinglots(parkCode, stateCode []string, start, limit int, q string) (*ParkinglotResponse, error) {
	url := api.BaseURL + "/parkinglots?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parkinglotResponse ParkinglotResponse
	if err := json.NewDecoder(resp.Body).Decode(&parkinglotResponse); err != nil {
		return nil, err
	}

	return &parkinglotResponse, nil
}

// Park represents a park in the National Park Service.
type Park struct {
	States         string `json:"states"`
	WeatherInfo    string `json:"weatherInfo"`
	DirectionsInfo string `json:"directionsInfo"`
	Addresses      []struct {
		CountryCode           string `json:"countryCode"`
		City                  string `json:"city"`
		ProvinceTerritoryCode string `json:"provinceTerritoryCode"`
		PostalCode            string `json:"postalCode"`
		Type                  string `json:"type"`
		Line1                 string `json:"line1"`
		StateCode             string `json:"stateCode"`
		Line2                 string `json:"line2"`
		Line3                 string `json:"line3"`
	} `json:"addresses"`
	EntranceFees []struct {
		Cost        string `json:"cost"`
		Description string `json:"description"`
		Title       string `json:"title"`
	} `json:"entranceFees"`
	Topics []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"topics"`
	Multimedia []struct {
		Title string `json:"title"`
		ID    string `json:"id"`
		Type  string `json:"type"`
		Url   string `json:"url"`
	} `json:"multimedia"`
	Name       string `json:"name"`
	Latitude   string `json:"latitude"`
	Activities []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"activities"`
	OperatingHours []struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		StandardHours struct {
			Sunday    string `json:"sunday"`
			Monday    string `json:"monday"`
			Tuesday   string `json:"tuesday"`
			Wednesday string `json:"wednesday"`
			Thursday  string `json:"thursday"`
			Friday    string `json:"friday"`
			Saturday  string `json:"saturday"`
		} `json:"standardHours"`
		Exceptions []struct {
			Name           string `json:"name"`
			StartDate      string `json:"startDate"`
			EndDate        string `json:"endDate"`
			ExceptionHours struct {
				Sunday    string `json:"sunday"`
				Monday    string `json:"monday"`
				Tuesday   string `json:"tuesday"`
				Wednesday string `json:"wednesday"`
				Thursday  string `json:"thursday"`
				Friday    string `json:"friday"`
				Saturday  string `json:"saturday"`
			} `json:"exceptionHours"`
		} `json:"exceptions"`
	} `json:"operatingHours"`
	Url       string `json:"url"`
	Longitude string `json:"longitude"`
	Contacts  struct {
		PhoneNumbers []struct {
			PhoneNumber string `json:"phoneNumber"`
			Description string `json:"description"`
			Extension   string `json:"extension"`
			Type        string `json:"type"`
		} `json:"phoneNumbers"`
		EmailAddresses []struct {
			EmailAddress string `json:"emailAddress"`
			Description  string `json:"description"`
		} `json:"emailAddresses"`
	} `json:"contacts"`
	EntrancePasses []struct {
		Cost        string `json:"cost"`
		Description string `json:"description"`
		Title       string `json:"title"`
	} `json:"entrancePasses"`
	ParkCode    string `json:"parkCode"`
	Designation string `json:"designation"`
	Images      []struct {
		Credit  string `json:"credit"`
		AltText string `json:"altText"`
		Title   string `json:"title"`
		Caption string `json:"caption"`
		Url     string `json:"url"`
	} `json:"images"`
	RelevanceScore int    `json:"relevanceScore"`
	FullName       string `json:"fullName"`
	LatLong        string `json:"latLong"`
	ID             string `json:"id"`
	DirectionsUrl  string `json:"directionsUrl"`
	Description    string `json:"description"`
}

// ParkResponse represents the response from the /parks endpoint.
type ParkResponse struct {
	Total string `json:"total"`
	Data  []Park `json:"data"`
	Start string `json:"start"`
	Limit string `json:"limit"`
}

// GetParks makes a GET request to the /parks endpoint and returns the parks.
func (api *npsApi) GetParks(parkCode, stateCode []string, start, limit int, q string, sort []string) (*ParkResponse, error) {
	url := api.BaseURL + "/parks?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q + "&sort=" + strings.Join(sort, ",")
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parkResponse ParkResponse
	if err := json.NewDecoder(resp.Body).Decode(&parkResponse); err != nil {
		return nil, err
	}

	return &parkResponse, nil
}

type PassportStampLocation struct {
	Label string `json:"label"`
	ID    string `json:"id"`
	Type  string `json:"type"`
	Parks []struct {
		States      string `json:"states"`
		Designation string `json:"designation"`
		ParkCode    string `json:"parkCode"`
		FullName    string `json:"fullName"`
		Url         string `json:"url"`
		Name        string `json:"name"`
	} `json:"parks"`
}

// PassportStampLocationResponse represents the response from the /passportstamplocations endpoint.
type PassportStampLocationResponse struct {
	Total string                  `json:"total"`
	Data  []PassportStampLocation `json:"data"`
	Start string                  `json:"start"`
	Limit string                  `json:"limit"`
}

// GetPassportStampLocations makes a GET request to the /passportstamplocations endpoint and returns the passport stamp locations.
func (api *npsApi) GetPassportStampLocations(parkCode, stateCode []string, q string, limit, start int) (*PassportStampLocationResponse, error) {
	url := api.BaseURL + "/passportstamplocations?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var passportStampLocationResponse PassportStampLocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&passportStampLocationResponse); err != nil {
		return nil, err
	}

	return &passportStampLocationResponse, nil
}

// Person represents a person related to national parks.
type Person struct {
	MiddleName           string   `json:"middleName"`
	LastName             string   `json:"lastName"`
	Latitude             string   `json:"latitude"`
	Url                  string   `json:"url"`
	Longitude            string   `json:"longitude"`
	BodyText             string   `json:"bodyText"`
	GeometryPoiId        string   `json:"geometryPoiId"`
	RelatedOrganizations []string `json:"relatedOrganizations"`
	Title                string   `json:"title"`
	Images               []struct {
		Credit string `json:"credit"`
		Crops  []struct {
			AspectRatio int    `json:"aspectratio"`
			URL         string `json:"url"`
		} `json:"crops"`
		AltText string `json:"altText"`
		Title   string `json:"title"`
		Caption string `json:"caption"`
		URL     string `json:"url"`
	} `json:"images"`
	ListingDescription string `json:"listingDescription"`
	QuickFacts         []struct {
		ID    string `json:"id"`
		Value string `json:"value"`
		Name  string `json:"name"`
	} `json:"quickFacts"`
	LatLong      string `json:"latLong"`
	ID           string `json:"id"`
	FirstName    string `json:"firstName"`
	RelatedParks []Park `json:"relatedParks"`
}

// PersonResponse represents the response from the /people endpoint.
type PersonResponse struct {
	Total string   `json:"total"`
	Data  []Person `json:"data"`
	Start string   `json:"start"`
	Limit string   `json:"limit"`
}

// GetPeople makes a GET request to the /people endpoint and returns the people.
func (api *npsApi) GetPeople(parkCode, stateCode []string, q string, limit, start int) (*PersonResponse, error) {
	url := api.BaseURL + "/people?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var personResponse PersonResponse
	if err := json.NewDecoder(resp.Body).Decode(&personResponse); err != nil {
		return nil, err
	}

	return &personResponse, nil
}

type Place struct {
	IsManagedByNps   string `json:"isManagedByNps"`
	AudioDescription string `json:"audioDescription"`
	Multimedia       []struct {
		Title string `json:"title"`
		ID    string `json:"id"`
		Type  string `json:"type"`
		Url   string `json:"url"`
	} `json:"multimedia"`
	Latitude             string   `json:"latitude"`
	ManagedByOrg         string   `json:"managedByOrg"`
	Url                  string   `json:"url"`
	Longitude            string   `json:"longitude"`
	BodyText             string   `json:"bodyText"`
	GeometryPoiId        string   `json:"geometryPoiId"`
	NpmapId              string   `json:"npmapId"`
	RelatedOrganizations []string `json:"relatedOrganizations"`
	Amenities            []string `json:"amenities"`
	Title                string   `json:"title"`
	Images               []string `json:"images"`
	ListingDescription   string   `json:"listingDescription"`
	IsOpenToPublic       string   `json:"isOpenToPublic"`
	Tags                 []string `json:"tags"`
	ManagedByUrl         string   `json:"managedByUrl"`
	QuickFacts           string   `json:"quickFacts"`
	LatLong              string   `json:"latLong"`
	ID                   string   `json:"id"`
	RelatedParks         []struct {
		States      string `json:"states"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
		FullName    string `json:"fullName"`
		Url         string `json:"url"`
		Name        string `json:"name"`
	} `json:"relatedParks"`
}

// PlaceResponse represents the response from the /places endpoint.
type PlaceResponse struct {
	Total string  `json:"total"`
	Data  []Place `json:"data"`
	Start string  `json:"start"`
	Limit string  `json:"limit"`
}

// GetPlaces makes a GET request to the /places endpoint and returns the places.
func (api *npsApi) GetPlaces(parkCode, stateCode []string, q string, limit, start int) (*PlaceResponse, error) {
	url := api.BaseURL + "/places?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&start=" + strconv.Itoa(start) + "&limit=" + strconv.Itoa(limit) + "&q=" + q
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var placeResponse PlaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&placeResponse); err != nil {
		return nil, err
	}

	return &placeResponse, nil
}

// RoadEvent represents a road event.
type RoadEvent struct {
	Geometry struct {
		Coordinates [][]float64 `json:"coordinates"`
		Type        string      `json:"type"`
	} `json:"geometry"`
	ID         string `json:"id"`
	Properties struct {
		IsEndDateVerified bool `json:"is_end_date_verified"`
		CoreDetails       struct {
			DataSourceID string   `json:"data_source_id"`
			Description  string   `json:"description"`
			Direction    string   `json:"direction"`
			EventType    string   `json:"event_type"`
			Name         string   `json:"name"`
			RoadNames    []string `json:"road_names"`
		} `json:"core_details"`
		IsEndPositionVerified   bool   `json:"is_end_position_verified"`
		IsStartDateVerified     bool   `json:"is_start_date_verified"`
		LocationMethod          string `json:"location_method"`
		StartDate               string `json:"start_date"`
		IsStartPositionVerified bool   `json:"is_start_position_verified"`
		EndDate                 string `json:"end_date"`
		TypesOfWork             []struct {
			TypeName string `json:"type_name"`
		} `json:"types_of_work"`
		VehicleImpact string `json:"vehicle_impact"`
	} `json:"properties"`
	Type string `json:"type"`
}

// RoadEventResponse represents the response from the /roadevents endpoint.
type RoadEventResponse struct {
	Features []RoadEvent `json:"features"`
	Type     string      `json:"type"`
}

// GetRoadEvents makes a GET request to the /roadevents endpoint and returns the road events.
func (api *npsApi) GetRoadEvents(parkCode, eventType string) (*RoadEventResponse, error) {
	url := api.BaseURL + "/roadevents?parkCode=" + parkCode + "&type=" + eventType
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var roadEventResponse RoadEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&roadEventResponse); err != nil {
		return nil, err
	}

	return &roadEventResponse, nil
}

type ThingsToDo struct {
	ShortDescription      string   `json:"shortDescription"`
	LongDescription       string   `json:"longDescription"`
	IsReservationRequired string   `json:"isReservationRequired"`
	Season                []string `json:"season"`
	Topics                []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"topics"`
	TimeOfDayDescription string `json:"timeOfDayDescription"`
	LocationDescription  string `json:"locationDescription"`
	PetsDescription      string `json:"petsDescription"`
	DurationDescription  string `json:"durationDescription"`
	Latitude             string `json:"latitude"`
	ActivityDescription  string `json:"activityDescription"`
	Activities           []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"activities"`
	URL                    string `json:"url"`
	Longitude              string `json:"longitude"`
	ReservationDescription string `json:"reservationDescription"`
	ArePetsPermitted       string `json:"arePetsPermitted"`
	GeometryPoiId          string `json:"geometryPoiId"`
	Duration               string `json:"duration"`
	Location               string `json:"location"`
	FeeDescription         string `json:"feeDescription"`
	DoFeesApply            string `json:"doFeesApply"`
	Title                  string `json:"title"`
	Images                 []struct {
		Credit string `json:"credit"`
		Crops  []struct {
			AspectRatio int    `json:"aspectratio"`
			URL         string `json:"url"`
		} `json:"crops"`
		AltText string `json:"altText"`
		Title   string `json:"title"`
		Caption string `json:"caption"`
		URL     string `json:"url"`
	} `json:"images"`
	TimeOfDay                        []string `json:"timeOfDay"`
	Tags                             []string `json:"tags"`
	SeasonDescription                string   `json:"seasonDescription"`
	RelevanceScore                   float64  `json:"relevanceScore"`
	ID                               string   `json:"id"`
	ArePetsPermittedWithRestrictions string   `json:"arePetsPermittedwithRestrictions"`
	AgeDescription                   string   `json:"ageDescription"`
	RelatedParks                     []struct {
		States      string `json:"states"`
		FullName    string `json:"fullName"`
		URL         string `json:"url"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
		Name        string `json:"name"`
	} `json:"relatedParks"`
	AccessibilityInformation string `json:"accessibilityInformation"`
	Age                      string `json:"age"`
}

type ThingsToDoResponse struct {
	Total string       `json:"total"`
	Data  []ThingsToDo `json:"data"`
	Limit string       `json:"limit"`
	Start string       `json:"start"`
}

// GetThingsToDo makes a GET request to the /thingstodo endpoint and returns the suggested things to do.
func (api *npsApi) GetThingsToDo(id, parkCode, stateCode, q string, limit, start int, sort []string) (*ThingsToDoResponse, error) {
	url := api.BaseURL + "/thingstodo?parkCode=" + parkCode + "&stateCode=" + stateCode + "&q=" + q + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start) + "&sort=" + strings.Join(sort, ",")
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var thingsToDoResponse ThingsToDoResponse
	if err := json.NewDecoder(resp.Body).Decode(&thingsToDoResponse); err != nil {
		return nil, err
	}

	return &thingsToDoResponse, nil
}

type Topic struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TopicResponse struct {
	Total string  `json:"total"`
	Data  []Topic `json:"data"`
	Limit string  `json:"limit"`
	Start string  `json:"start"`
}

// GetTopics makes a GET request to the /topics endpoint and returns the topics.
func (api *npsApi) GetTopics(id, q string, limit, start int, sort string) (*TopicResponse, error) {
	url := api.BaseURL + "/topics?id=" + id + "&q=" + q + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start) + "&sort=" + sort
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var topicResponse TopicResponse
	if err := json.NewDecoder(resp.Body).Decode(&topicResponse); err != nil {
		return nil, err
	}

	return &topicResponse, nil
}

type TopicPark struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Parks []struct {
		States      string `json:"states"`
		FullName    string `json:"fullName"`
		URL         string `json:"url"`
		ParkCode    string `json:"parkCode"`
		Designation string `json:"designation"`
		Name        string `json:"name"`
	} `json:"parks"`
}

type TopicParkResponse struct {
	Total string      `json:"total"`
	Data  []TopicPark `json:"data"`
	Limit string      `json:"limit"`
	Start string      `json:"start"`
}

// GetTopicParks makes a GET request to the /topics/parks endpoint and returns the topic parks.
func (api *npsApi) GetTopicParks(id []string, q string, limit, start int, sort string) (*TopicParkResponse, error) {
	url := api.BaseURL + "/topics/parks?id=" + strings.Join(id, ",") + "&q=" + q + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start) + "&sort=" + sort
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var topicParkResponse TopicParkResponse
	if err := json.NewDecoder(resp.Body).Decode(&topicParkResponse); err != nil {
		return nil, err
	}

	return &topicParkResponse, nil
}

type Tour struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	DurationMin    string  `json:"durationMin"`
	DurationMax    string  `json:"durationMax"`
	DurationUnit   string  `json:"durationUnit"`
	RelevanceScore float64 `json:"relevanceScore"`
	Topics         []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"topics"`
	Park struct {
		States      string `json:"states"`
		Designation string `json:"designation"`
		ParkCode    string `json:"parkCode"`
		FullName    string `json:"fullName"`
		URL         string `json:"url"`
		Name        string `json:"name"`
	} `json:"park"`
	Activities []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"activities"`
	Stops []struct {
		Significance         string `json:"significance"`
		AssetID              string `json:"assetId"`
		AssetName            string `json:"assetName"`
		AssetType            string `json:"assetType"`
		ID                   string `json:"id"`
		Ordinal              string `json:"ordinal"`
		DirectionsToNextStop string `json:"directionsToNextStop"`
	} `json:"stops"`
	Images []struct {
		Credit string `json:"credit"`
		Crops  []struct {
			AspectRatio string `json:"aspectratio"`
			URL         string `json:"url"`
		} `json:"crops"`
		AltText string `json:"altText"`
		Title   string `json:"title"`
		Caption string `json:"caption"`
		URL     string `json:"url"`
	} `json:"images"`
}

type TourResponse struct {
	Total string `json:"total"`
	Data  []Tour `json:"data"`
	Limit string `json:"limit"`
	Start string `json:"start"`
}

// GetTours makes a GET request to the /tours endpoint and returns the tours.
func (api *npsApi) GetTours(id, parkCode, stateCode []string, q string, limit, start int, sort []string) (*TourResponse, error) {
	url := api.BaseURL + "/tours?id=" + strings.Join(id, ",") + "&parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&q=" + q + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start) + "&sort=" + strings.Join(sort, ",")
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tourResponse TourResponse
	if err := json.NewDecoder(resp.Body).Decode(&tourResponse); err != nil {
		return nil, err
	}

	return &tourResponse, nil
}

type VisitorCenter struct {
	Total string `json:"total"`
	Data  []struct {
		DirectionsInfo string `json:"directionsInfo"`
		Addresses      []struct {
			CountryCode           string `json:"countryCode"`
			City                  string `json:"city"`
			ProvinceTerritoryCode string `json:"provinceTerritoryCode"`
			PostalCode            string `json:"postalCode"`
			Type                  string `json:"type"`
			Line1                 string `json:"line1"`
			StateCode             string `json:"stateCode"`
			Line2                 string `json:"line2"`
			Line3                 string `json:"line3"`
		} `json:"addresses"`
		AudioDescription    string `json:"audioDescription"`
		PassportStampImages []struct {
			Credit      string `json:"credit"`
			Description string `json:"description"`
			Crops       []struct {
				AspectRatio float64 `json:"aspectRatio"`
				URL         string  `json:"url"`
			} `json:"crops"`
			AltText string `json:"altText"`
			Title   string `json:"title"`
			Caption string `json:"caption"`
			URL     string `json:"url"`
		} `json:"passportStampImages"`
		LastIndexedDate string `json:"lastIndexedDate"`
		Multimedia      []struct {
			Title string `json:"title"`
			ID    string `json:"id"`
			Type  string `json:"type"`
			URL   string `json:"url"`
		} `json:"multimedia"`
		Name           string `json:"name"`
		Latitude       string `json:"latitude"`
		OperatingHours []struct {
			Name          string            `json:"name"`
			Description   string            `json:"description"`
			StandardHours map[string]string `json:"standardHours"`
			Exceptions    []struct {
				Name           string            `json:"name"`
				StartDate      string            `json:"startDate"`
				EndDate        string            `json:"endDate"`
				ExceptionHours map[string]string `json:"exceptionHours"`
			} `json:"exceptions"`
		} `json:"operatingHours"`
		URL                              string   `json:"url"`
		Longitude                        string   `json:"longitude"`
		Contacts                         []string `json:"contacts"`
		GeometryPoiId                    string   `json:"geometryPoiId"`
		PassportStampLocationDescription string   `json:"passportStampLocationDescription"`
		ParkCode                         string   `json:"parkCode"`
		Amenities                        []string `json:"amenities"`
		Images                           []struct {
			Credit string `json:"credit"`
			Crops  []struct {
			} `json:"crops"`
			Title   string `json:"title"`
			AltText string `json:"altText"`
			Caption string `json:"caption"`
			URL     string `json:"url"`
		} `json:"images"`
		RelevanceScore          float64 `json:"relevanceScore"`
		LatLong                 string  `json:"latLong"`
		ID                      string  `json:"id"`
		DirectionsURL           string  `json:"directionsUrl"`
		IsPassportStampLocation bool    `json:"isPassportStampLocation"`
		Description             string  `json:"description"`
	} `json:"data"`
	Limit string `json:"limit"`
	Start string `json:"start"`
}

type VisitorCenterResponse struct {
	Total string          `json:"total"`
	Data  []VisitorCenter `json:"data"`
	Limit string          `json:"limit"`
	Start string          `json:"start"`
}

// GetVisitorCenters makes a GET request to the /visitorcenters endpoint and returns the visitor centers.
func (api *npsApi) GetVisitorCenters(parkCode, stateCode []string, q string, limit, start int, sort []string) (*VisitorCenterResponse, error) {
	url := api.BaseURL + "/visitorcenters?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&q=" + q + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start) + "&sort=" + strings.Join(sort, ",")
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var visitorCenterResponse VisitorCenterResponse
	if err := json.NewDecoder(resp.Body).Decode(&visitorCenterResponse); err != nil {
		return nil, err
	}

	return &visitorCenterResponse, nil
}

type Webcam struct {
	Total string `json:"total"`
	Data  []struct {
		Latitude      float64 `json:"latitude"`
		URL           string  `json:"url"`
		Longitude     float64 `json:"longitude"`
		Status        string  `json:"status"`
		GeometryPoiId string  `json:"geometryPoiId"`
		StatusMessage string  `json:"statusMessage"`
		Title         string  `json:"title"`
		IsStreaming   bool    `json:"isStreaming"`
		Images        []struct {
			URL         string `json:"url"`
			Credit      string `json:"credit"`
			AltText     string `json:"altText"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Caption     string `json:"caption"`
			Crops       []struct {
				AspectRatio float64 `json:"aspectRatio"`
				URL         string  `json:"url"`
			} `json:"crops"`
		} `json:"images"`
		Tags         []string `json:"tags"`
		ID           string   `json:"id"`
		Description  string   `json:"description"`
		RelatedParks []string `json:"relatedParks"`
	} `json:"data"`
	Limit string `json:"limit"`
	Start string `json:"start"`
}

type WebcamResponse struct {
	Total string   `json:"total"`
	Data  []Webcam `json:"data"`
	Limit string   `json:"limit"`
	Start string   `json:"start"`
}

// GetWebcams makes a GET request to the /webcams endpoint and returns the webcams.
func (api *npsApi) GetWebcams(id string, parkCode, stateCode []string, q string, limit, start int) (*WebcamResponse, error) {
	url := api.BaseURL + "/webcams?parkCode=" + strings.Join(parkCode, ",") + "&stateCode=" + strings.Join(stateCode, ",") + "&id=" + id + "&limit=" + strconv.Itoa(limit) + "&start=" + strconv.Itoa(start) + "&q=" + q
	resp, err := api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var webcamResponse WebcamResponse
	if err := json.NewDecoder(resp.Body).Decode(&webcamResponse); err != nil {
		return nil, err
	}

	return &webcamResponse, nil
}
