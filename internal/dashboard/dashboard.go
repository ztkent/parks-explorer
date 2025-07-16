package dashboard

import (
	"github.com/ztkent/go-nps"
	"github.com/ztkent/nps-dashboard/internal/database"
)

type Dashboard struct {
	npsApi      nps.NpsApi
	db          *database.DB
	parkService *ParkService
}

func NewDashboard(apiKey string, dbPath string) *Dashboard {
	// Initialize database
	db, err := database.NewDatabase(dbPath)
	if err != nil {
		panic(err)
	}

	// Initialize NPS API
	npsApi := nps.NewNpsApi(apiKey)

	// Initialize park service
	parkService := NewParkService(npsApi, db)
	go parkService.GetAllParks()

	return &Dashboard{
		npsApi:      npsApi,
		db:          db,
		parkService: parkService,
	}
}
