package dashboard

import "github.com/ztkent/go-nps"

type Dashboard struct {
	npsApi nps.NpsApi
}

func NewDashboard(apiKey string) *Dashboard {
	return &Dashboard{
		npsApi: nps.NewNpsApi(apiKey),
	}
}

func (d *Dashboard) ListAllParks() {
	res, err := d.npsApi.GetParks(nil, nil, 0, 0, "", nil)
	if err != nil {
		panic(err)
	}
	for _, park := range res.Data {
		println(park.Name)
	}
}

func (d *Dashboard) GetVisitorCentersInState(state string) {
	res, err := d.npsApi.GetVisitorCenters(nil, []string{state}, "", 0, 1, nil)
	if err != nil {
		panic(err)
	}
	for _, visitorCenter := range res.Data {
		println(visitorCenter.Name)
	}
}

func (d *Dashboard) ListImportantPeopleInPark(siteCode string) {
	res, err := d.npsApi.GetPeople([]string{siteCode}, nil, "", 0, 0)
	if err != nil {
		panic(err)
	}
	for _, person := range res.Data {
		println(person.FirstName + " " + person.LastName)
	}
}
