package main

import (
	"os"

	"github.com/Ztkent/nps-dashboard/nps"
)

func main() {
	npsApi := nps.NewNpsApi(os.Getenv("NPS_API_KEY"))
	ListAllParks(npsApi)
}

func ListAllParks(npsApi nps.NpsApi) {
}
