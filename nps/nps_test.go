//go:build integration

package nps

import (
	"fmt"
	"os"
	"testing"
)

const (
	testStart    = 0
	testLimit    = 10
	testString   = ""
	testSiteCode = "ACAD"
	testState    = "CA"
)

func TestApiEndpoints(t *testing.T) {
	api := NewNpsApi(os.Getenv("NPS_API_KEY"))

	t.Run("GetActivities", func(t *testing.T) {
		res, err := api.GetActivities(testString, testString, testLimit, testStart, testString)
		checkResponse(t, res, err)
	})

	t.Run("GetActivityParks", func(t *testing.T) {
		res, err := api.GetActivityParks([]string{testString}, testString, testLimit, testStart, testString)
		checkResponse(t, res, err)
	})

	t.Run("GetAlerts", func(t *testing.T) {
		res, err := api.GetAlerts([]string{testString}, []string{testString}, testString, testLimit, testStart)
		checkResponse(t, res, err)
	})

	t.Run("GetAmenities", func(t *testing.T) {
		res, err := api.GetAmenities([]string{testString}, testString, testLimit, testStart)
		checkResponse(t, res, err)
	})

	t.Run("GetAmenitiesParksPlaces", func(t *testing.T) {
		res, err := api.GetAmenitiesParksPlaces([]string{testString}, []string{testString}, testString, testLimit, testStart, testString)
		checkResponse(t, res, err)
	})

	t.Run("GetAmenitiesParksVisitorCenters", func(t *testing.T) {
		res, err := api.GetAmenitiesParksVisitorCenters(testString, testString, testString, testLimit, testStart, []string{testString})
		checkResponse(t, res, err)
	})

	t.Run("GetArticles", func(t *testing.T) {
		res, err := api.GetArticles([]string{testString}, []string{testString}, testString, testLimit, testStart)
		checkResponse(t, res, err)
	})

	t.Run("GetCampgrounds", func(t *testing.T) {
		res, err := api.GetCampgrounds([]string{testString}, []string{testString}, testString, testLimit, testStart, []string{testString})
		checkResponse(t, res, err)
	})

	t.Run("GetEvents", func(t *testing.T) {
		res, err := api.GetEvents([]string{testString}, []string{testString}, []string{testString}, []string{testString}, []string{testString}, []string{testString}, []string{testString}, []string{testString}, testString, testString, []string{testString}, testString, testString, testLimit, testStart, true)
		checkResponse(t, res, err)
	})

	t.Run("GetFeesPasses", func(t *testing.T) {
		res, err := api.GetFeesPasses([]string{testString}, []string{testString}, testString, testStart, testLimit, []string{testString})
		checkResponse(t, res, err)
	})

	t.Run("GetLessonPlans", func(t *testing.T) {
		res, err := api.GetLessonPlans([]string{testString}, []string{testString}, testString, testStart, testLimit, []string{testString})
		checkResponse(t, res, err)
	})

	t.Run("GetParkBoundaries", func(t *testing.T) {
		res, err := api.GetParkBoundaries(testSiteCode)
		checkResponse(t, res, err)
	})

	t.Run("GetMultimediaAudio", func(t *testing.T) {
		res, err := api.GetMultimediaAudio([]string{testString}, []string{testString}, testString, testStart, testLimit)
		checkResponse(t, res, err)
	})

	t.Run("GetMultimediaGalleries", func(t *testing.T) {
		res, err := api.GetMultimediaGalleries([]string{testString}, []string{testString}, testString, testStart, testLimit)
		checkResponse(t, res, err)
	})

	t.Run("GetMultimediaGalleriesAssets", func(t *testing.T) {
		res, err := api.GetMultimediaGalleriesAssets(testString, testString, []string{testString}, []string{testString}, testString, testStart, testLimit)
		checkResponse(t, res, err)
	})

	t.Run("GetMultimediaVideos", func(t *testing.T) {
		res, err := api.GetMultimediaVideos([]string{testString}, []string{testString}, testString, testStart, testLimit)
		checkResponse(t, res, err)
	})

	t.Run("GetNewsReleases", func(t *testing.T) {
		res, err := api.GetNewsReleases([]string{testString}, []string{testString}, testString, testLimit, testStart, []string{testString})
		checkResponse(t, res, err)
	})

	t.Run("GetParkinglots", func(t *testing.T) {
		res, err := api.GetParkinglots([]string{testString}, []string{testString}, testString, testStart, testLimit)
		checkResponse(t, res, err)
	})

	t.Run("GetParks", func(t *testing.T) {
		res, err := api.GetParks([]string{testString}, []string{testString}, testStart, testLimit, testString, []string{testString})
		checkResponse(t, res, err)
	})

	t.Run("GetPassportStampLocations", func(t *testing.T) {
		res, err := api.GetPassportStampLocations([]string{testString}, []string{testString}, testString, testLimit, testStart)
		checkResponse(t, res, err)
	})

	t.Run("GetPeople", func(t *testing.T) {
		res, err := api.GetPeople([]string{testString}, []string{testString}, testString, testLimit, testStart)
		checkResponse(t, res, err)
	})

	t.Run("GetPlaces", func(t *testing.T) {
		res, err := api.GetPlaces([]string{testString}, []string{testString}, testString, testLimit, testStart)
		checkResponse(t, res, err)
	})

	t.Run("GetRoadEvents", func(t *testing.T) {
		res, err := api.GetRoadEvents(testString, testString)
		checkResponse(t, res, err)
	})

	t.Run("GetThingsToDo", func(t *testing.T) {
		res, err := api.GetThingsToDo(testString, testString, testString, testString, testLimit, testStart, []string{testString})
		checkResponse(t, res, err)
	})

	t.Run("GetTopics", func(t *testing.T) {
		res, err := api.GetTopics(testString, testString, testLimit, testStart, testString)
		checkResponse(t, res, err)
	})

	t.Run("GetTopicParks", func(t *testing.T) {
		res, err := api.GetTopicParks([]string{testString}, testString, testLimit, testStart, testString)
		checkResponse(t, res, err)
	})

	t.Run("GetTours", func(t *testing.T) {
		res, err := api.GetTours([]string{testString}, []string{testString}, []string{testString}, testString, testLimit, testStart, []string{testString})
		checkResponse(t, res, err)
	})

	t.Run("GetVisitorCenters", func(t *testing.T) {
		res, err := api.GetVisitorCenters([]string{testString}, []string{testState}, testString, testLimit, testStart, []string{testString})
		checkResponse(t, res, err)
	})

	t.Run("GetWebcams", func(t *testing.T) {
		res, err := api.GetWebcams(testString, []string{testString}, []string{testString}, testString, testLimit, testStart)
		checkResponse(t, res, err)
	})
}

func checkResponse(t *testing.T, res interface{}, err error) {
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %v, Response: %v", err, res))
		t.Errorf("Error: %v, Response: %v", err, res)
	} else if res == nil {
		fmt.Println("Expected results, got nil")
		t.Errorf("Expected results, got nil")
	}
}
