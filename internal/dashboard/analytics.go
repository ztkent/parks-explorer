package dashboard

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// AnalyticsConfig represents the configuration for Google Analytics
type AnalyticsConfig struct {
	GoogleAnalyticsID string `json:"googleAnalyticsId"`
	Enabled           bool   `json:"enabled"`
	Debug             bool   `json:"debug"`
}

// GetAnalyticsConfig returns the current analytics configuration
func GetAnalyticsConfig() *AnalyticsConfig {
	config := &AnalyticsConfig{
		GoogleAnalyticsID: os.Getenv("GOOGLE_ANALYTICS_ID"),
		Enabled:           os.Getenv("GOOGLE_ANALYTICS_ENABLED") == "true",
		Debug:             os.Getenv("GOOGLE_ANALYTICS_DEBUG") == "true",
	}

	// Default to enabled if the environment variable is not set but ID is provided
	if config.GoogleAnalyticsID != "" && os.Getenv("GOOGLE_ANALYTICS_ENABLED") == "" {
		config.Enabled = true
	}

	return config
}

// AnalyticsConfigHandler returns the analytics configuration as JSON for client-side use
func (dm *Dashboard) AnalyticsConfigHandler(w http.ResponseWriter, r *http.Request) {
	config := GetAnalyticsConfig()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		log.Printf("Failed to encode analytics config: %v", err)
		http.Error(w, "Failed to get analytics configuration", http.StatusInternalServerError)
		return
	}
}

// TrackPageView logs a page view for server-side tracking
func TrackPageView(r *http.Request, page string) {
	config := GetAnalyticsConfig()
	if !config.Enabled || config.Debug {
		if config.Debug {
			log.Printf("GA Debug: Page view tracked - Page: %s, User-Agent: %s, IP: %s",
				page, r.UserAgent(), r.RemoteAddr)
		}
		return
	}
}

// TrackEvent logs an event for server-side tracking
func TrackEvent(r *http.Request, action, category, label string, value int) {
	config := GetAnalyticsConfig()
	if !config.Enabled || config.Debug {
		if config.Debug {
			log.Printf("GA Debug: Event tracked - Action: %s, Category: %s, Label: %s, Value: %d, User-Agent: %s",
				action, category, label, value, r.UserAgent())
		}
		return
	}

	// In a production environment, you might want to send this data to Google Analytics
	// via the Measurement Protocol API for server-side tracking
}
