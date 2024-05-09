# Go NPS
Go wrapper for the [National Park Service API](https://www.nps.gov/subjects/developer/index.htm).   
Handles HTTP requests to the NPS API, provides a structured response.  

https://www.nps.gov/subjects/developer/api-documentation.htm

## About
The National Park Service API is designed to provide authoritative NPS data and content for internal and external developers creating apps, maps, and websites. They feature photos and essential information about NPS sites including visitor centers, campgrounds, events, news, alerts, and more, as well as detailed articles about NPS natural and cultural features and important people and places.   
[Content Policy](https://www.nps.gov/aboutus/disclaimer.htm)

## Usage 
### Installation
```bash
go get github.com/ztkent/go-nps
```

### Example
```go
// Create a new NPS API client
nps := NewNpsApi(os.Getenv("NPS_API_KEY"))

// Get a list of all national parks, starting from the 5th park and returning 10 parks
parks, err := nps.GetParks(nil, nil, 5, 10, "", nil)

// Get a list of all activities available at Acadia National Park
activities, err := nps.GetActivityParks([]string{"acad"}, "", "", 0, 0)

```

### Endpoints
| Method | Description |
| --- | --- |
| `GetActivities` | Retrieves a list of activities available in the parks. |
| `GetActivityParks` | Retrieves a list of parks where a specific activity is available. |
| `GetAlerts` | Retrieves a list of current alerts and warnings for the parks. |
| `GetAmenities` | Retrieves a list of amenities available in the parks. |
| `GetAmenitiesParksPlaces` | Retrieves a list of places in the parks with specific amenities. |
| `GetAmenitiesParksVisitorCenters` | Retrieves a list of visitor centers in the parks with specific amenities. |
| `GetArticles` | Retrieves a list of articles related to the parks. |
| `GetCampgrounds` | Retrieves a list of campgrounds available in the national parks. |
| `GetEvents` | Retrieves a list of current and upcoming events happening in the national parks. |
| `GetFeesPasses` | Retrieves detailed information about the fees and passes required for visiting the national parks. |
| `GetLessonPlans` | Retrieves a list of lesson plans for educational activities available in the national parks. |
| `GetParkBoundaries` | Retrieves detailed information about the geographical boundaries of the national parks. |
| `GetMultimediaAudio` | Retrieves a list of audio files related to the history, wildlife, and other aspects of the national parks. |
| `GetMultimediaGalleries` | Retrieves a list of image galleries showcasing the beauty and diversity of the national parks. |
| `GetMultimediaGalleriesAssets` | Retrieves a list of assets (images, descriptions, etc.) for the image galleries related to the national parks. |
| `GetMultimediaVideos` | Retrieves a list of videos that provide a visual representation of various features and aspects of the national parks. |
| `GetNewsReleases` | Retrieves a list of news releases or updates related to the national parks. |
| `GetParkinglots` | Retrieves a list of parking lots available in the national parks. |
| `GetParks` | Retrieves a comprehensive list of all national parks. |
| `GetPassportStampLocations` | Retrieves a list of locations within the national parks where visitors can get their passports stamped. |
| `GetPeople` | Retrieves a list of notable individuals who have a significant association or history with the national parks. |
| `GetPlaces` |Retrieves a list of notable places within the parks, such as landmarks, historical sites, and natural wonders. |
| `GetRoadEvents` |Retrieves a list of current road events in the parks, including road closures, construction, and other traffic-related incidents. |
| `GetThingsToDo` |Retrieves a list of recommended activities in the parks, such as hiking, camping, bird watching, etc. |
| `GetTopics` |Retrieves a list of topics related to the parks, such as conservation, wildlife, geology, history, etc. |
| `GetTopicParks` |Retrieves a list of parks related to a specific topic. For example, if the topic is 'wildlife', it will return parks known for their wildlife. |
| `GetTours` |Retrieves a list of available tours in the parks, including guided tours, self-guided tours, and virtual tours. |
| `GetVisitorCenters` |Retrieves a list of visitor centers in the parks, including their locations, hours of operation, and available services. |
| `GetWebcams` |Retrieves a list of webcams in the parks, allowing users to view real-time images of different areas within the parks. |