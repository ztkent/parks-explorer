package main

import (
	"os"

	"github.com/Ztkent/nps-dashboard/nps"
)

func main() {
	npsApi := nps.NewNpsApi(os.Getenv("NPS_API_KEY"))
	ListImportantPeopleInPark(npsApi, "ACAD")
	GetVisitorCentersInState(npsApi, "CA")
	ListAllParks(npsApi)
}

func ListAllParks(npsApi nps.NpsApi) {
	res, err := npsApi.GetParks(nil, nil, 0, 0, "", nil)
	if err != nil {
		panic(err)
	}
	for _, park := range res.Data {
		println(park.Name)
	}
}

func GetVisitorCentersInState(npsApi nps.NpsApi, state string) {
	res, err := npsApi.GetVisitorCenters(nil, []string{state}, "", 0, 1, nil)
	if err != nil {
		panic(err)
	}
	for _, visitorCenter := range res.Data {
		println(visitorCenter.Name)
	}
}

func ListImportantPeopleInPark(npsApi nps.NpsApi, siteCode string) {
	res, err := npsApi.GetPeople([]string{siteCode}, nil, "", 0, 0)
	if err != nil {
		panic(err)
	}
	for _, person := range res.Data {
		println(person.FirstName + " " + person.LastName)
	}
}
