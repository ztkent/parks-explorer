package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// LiveParkCamsHandler returns a list of live park cameras.
type ParkCam struct {
	Title string `json:"title"`
	Image string `json:"image"`
	Link  string `json:"link"`
}

func (dm *Dashboard) LiveParkCamsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parkCams := []ParkCam{
			{
				Title: "Park Cam 1",
				Image: "https://via.placeholder.com/100",
				Link:  "#",
			},
			{
				Title: "Park Cam 2",
				Image: "https://via.placeholder.com/100",
				Link:  "#",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(parkCams)
	}
}

// ParkListHandler returns a list of parks.
type ParkList struct {
	Parks []string `json:"parks"`
}

func (dm *Dashboard) ParkListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := dm.npsApi.GetParks(nil, nil, 0, 500, "", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		parkList := ParkList{}
		for _, park := range res.Data {
			parkList.Parks = append(parkList.Parks, park.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(parkList)
	}
}

// Unique identifier for the user, stored in a cookie.
func (dm *Dashboard) EnsureUUIDHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("uuid")
		if err == http.ErrNoCookie {
			// Cookie does not exist, set it
			token := uuid.New().String()
			http.SetCookie(w, &http.Cookie{
				Name:     "uuid",
				Value:    token,
				HttpOnly: true,
				Secure:   true, // Set to true if your site uses HTTPS
				SameSite: http.SameSiteStrictMode,
			})
		} else if err != nil {
			// Some other error occurred
			http.Error(w, "Failed to read cookie", http.StatusInternalServerError)
		}
	}
}
