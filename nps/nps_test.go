package nps

import (
	"testing"
)

func TestGetActivitiesIntegration(t *testing.T) {
	api := NewNpsApi()
	res, err := api.GetActivities("", "", "", 10, 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	} else if res.Total == "0" {
		t.Errorf("Expected results, got %v", res.Total)
	}
}
