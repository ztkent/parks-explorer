package main

import (
	"os"

	"github.com/Ztkent/go-nps"
	"github.com/Ztkent/nps-dashboard/internal/dashboard"
)

func main() {
	npsApi := nps.NewNpsApi(os.Getenv("NPS_API_KEY"))
	dashboard.ListImportantPeopleInPark(npsApi, "ACAD")
	dashboard.GetVisitorCentersInState(npsApi, "CA")
	dashboard.ListAllParks(npsApi)
}
